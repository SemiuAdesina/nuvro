package alerting_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
)

var day = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)

func at(hour, minute int) time.Time {
	return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

func item(id string, startHour int, category place.Category) itinerary.Item {
	return itinerary.Item{
		ID: id, TripID: "trip", Day: 3, StartAt: at(startHour, 0), EndAt: at(startHour+2, 0),
		Kind: itinerary.Activity, Title: id, PlaceID: "place-" + id, Category: category,
		Location: geo.Point{Latitude: 32.6, Longitude: -16.9},
		Status:   itinerary.Planned, Source: itinerary.Generated,
	}
}

func window(condition weather.Condition, hours ...int) weather.Window {
	observations := make([]weather.Observation, 0, len(hours))
	for _, hour := range hours {
		observations = append(observations, weather.Observation{
			At: at(hour, 0), Condition: condition, TemperatureCelsius: 19, PrecipitationChance: 80,
		})
	}
	return weather.Window{Availability: weather.Available, Hourly: observations}
}

func input(items ...itinerary.Item) alerting.Input {
	return alerting.Input{
		TripID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f", Timezone: "Atlantic/Madeira",
		Now: at(12, 0), Itinerary: itinerary.Itinerary{TripID: "trip", Items: items},
	}
}

func codes(alerts []alert.Alert) map[alert.Type]int {
	found := map[alert.Type]int{}
	for _, value := range alerts {
		found[value.Type]++
	}
	return found
}

func TestWeatherAlertOnlyWhenItAffectsAnActivity(t *testing.T) {
	garden := item("garden", 15, place.Park)
	state := input(garden)
	state.Weather = window(weather.HeavyRain, 15)
	alerts := alerting.New(alerting.WeatherImpact{}).Evaluate(state)
	if codes(alerts)[alert.WeatherImpact] != 1 {
		t.Fatalf("alerts = %v", alerts)
	}
	produced := alerts[0]
	if produced.Severity != alert.Important || produced.RelatedItemID != garden.ID {
		t.Fatalf("alert = %+v", produced)
	}
	if produced.MessageData["activity"] != garden.Title {
		t.Fatalf("message data = %v", produced.MessageData)
	}
	if produced.TitleCode == "" || produced.Action != alert.AdjustPlan {
		t.Fatal("the alert carries no title code or action")
	}
}

func TestPleasantWeatherProducesNoAlert(t *testing.T) {
	state := input(item("garden", 15, place.Park))
	state.Weather = window(weather.Clear, 15)
	if alerts := alerting.New(alerting.WeatherImpact{}).Evaluate(state); len(alerts) != 0 {
		t.Fatalf("clear weather produced %v", alerts)
	}
}

func TestIndoorActivitiesAreNotFlaggedForRain(t *testing.T) {
	state := input(item("museum", 15, place.Museum))
	state.Weather = window(weather.Rain, 15)
	if alerts := alerting.New(alerting.WeatherImpact{}).Evaluate(state); len(alerts) != 0 {
		t.Fatalf("an indoor visit was flagged for rain: %v", alerts)
	}
}

func TestUnknownWeatherProducesNoAlert(t *testing.T) {
	state := input(item("garden", 15, place.Park))
	state.Weather = weather.Unavailable(weather.ProviderUnavailable)
	if alerts := alerting.New(alerting.WeatherImpact{}).Evaluate(state); len(alerts) != 0 {
		t.Fatalf("unknown weather produced %v", alerts)
	}
}

func TestFinishedActivitiesAreNotFlagged(t *testing.T) {
	past := item("garden", 9, place.Park)
	completed := item("done", 15, place.Park)
	completed.Status = itinerary.Completed
	state := input(past, completed)
	state.Weather = window(weather.Storm, 9, 15)
	if alerts := alerting.New(alerting.WeatherImpact{}).Evaluate(state); len(alerts) != 0 {
		t.Fatalf("past or completed activities were flagged: %v", alerts)
	}
}

func TestWeatherAlertsDeduplicateAcrossRepeatedEvaluations(t *testing.T) {
	state := input(item("garden", 15, place.Park))
	state.Weather = window(weather.HeavyRain, 15)
	engine := alerting.New(alerting.WeatherImpact{})
	first := engine.Evaluate(state)
	for offset := range 10 {
		state.Now = at(12, offset)
		again := engine.Evaluate(state)
		if len(again) != 1 || again[0].Fingerprint != first[0].Fingerprint {
			t.Fatalf("repeated evaluation produced a different alert: %v", again)
		}
	}
	state.Weather = window(weather.Clear, 15)
	if len(engine.Evaluate(state)) != 0 {
		t.Fatal("the alert survived the condition clearing")
	}
}

func TestLeaveReminderOnlyInsideTheReminderWindow(t *testing.T) {
	next := item("cable-car", 14, place.Attraction)
	state := input(next)
	state.NextItem = &next
	state.Departure = travel.Departure{
		Status: travel.LeaveSoon, LeaveBy: at(12, 20), HasLeaveBy: true,
		MinutesUntilLeave: 20, HasRoute: true,
		Route: route.Evidence{Status: route.Found, Mode: route.Walk, DurationSeconds: 900},
	}
	alerts := alerting.New(alerting.LeaveTime{}).Evaluate(state)
	if codes(alerts)[alert.LeaveReminder] != 1 {
		t.Fatalf("alerts = %v", alerts)
	}
	if alerts[0].MessageData["leaveBy"] == "" || alerts[0].Action != alert.NavigateTo {
		t.Fatalf("alert = %+v", alerts[0])
	}
	early := state
	early.Departure.Status = travel.OnTime
	early.Departure.LeaveBy = at(16, 0)
	if len(alerting.New(alerting.LeaveTime{}).Evaluate(early)) != 0 {
		t.Fatal("a reminder fired while the traveller was comfortably on time")
	}
}

func TestLeaveReminderRequiresRouteEvidence(t *testing.T) {
	next := item("cable-car", 14, place.Attraction)
	state := input(next)
	state.NextItem = &next
	state.Departure = travel.Departure{Status: travel.RouteUnavailable, MinutesUntilLeave: -1}
	if alerts := alerting.New(alerting.LeaveTime{}).Evaluate(state); len(alerts) != 0 {
		t.Fatalf("a reminder was produced without route evidence: %v", alerts)
	}
}

func TestLeaveReminderDoesNotRepeatEveryMinute(t *testing.T) {
	next := item("cable-car", 14, place.Attraction)
	state := input(next)
	state.NextItem = &next
	engine := alerting.New(alerting.LeaveTime{})
	fingerprints := map[string]struct{}{}
	for offset := range 10 {
		state.Now = at(12, offset)
		state.Departure = travel.Departure{
			Status: travel.LeaveSoon, LeaveBy: at(12, 25), HasLeaveBy: true, HasRoute: true,
		}
		for _, produced := range engine.Evaluate(state) {
			fingerprints[produced.Fingerprint] = struct{}{}
		}
	}
	if len(fingerprints) > 2 {
		t.Fatalf("ten evaluations produced %d distinct reminders", len(fingerprints))
	}
}

func TestScheduleConflictsAreReportedOnce(t *testing.T) {
	first := item("first", 14, place.Museum)
	second := item("second", 15, place.Museum)
	state := input(first, second)
	alerts := alerting.New(alerting.ScheduleConflict{}).Evaluate(state)
	if codes(alerts)[alert.ScheduleConflict] == 0 {
		t.Fatalf("an overlap produced no alert: %v", alerts)
	}
	engine := alerting.New(alerting.ScheduleConflict{})
	firstRun := engine.Evaluate(state)
	secondRun := engine.Evaluate(state)
	if len(firstRun) != len(secondRun) || firstRun[0].Fingerprint != secondRun[0].Fingerprint {
		t.Fatal("conflict alerts are not stable across evaluations")
	}
}

func TestAlertsAreOrderedBySeverity(t *testing.T) {
	garden := item("garden", 15, place.Park)
	next := item("next", 14, place.Attraction)
	state := input(garden, next)
	state.Weather = window(weather.HeavyRain, 15)
	state.NextItem = &next
	state.Departure = travel.Departure{Status: travel.LeaveSoon, LeaveBy: at(12, 25), HasLeaveBy: true, HasRoute: true}
	alerts := alerting.New().Evaluate(state)
	if len(alerts) < 2 {
		t.Fatalf("alerts = %v", alerts)
	}
	if alerts[0].Severity != alert.Important {
		t.Fatalf("the most severe alert is not first: %v", alerts)
	}
	for _, produced := range alerts {
		if produced.Validate() != nil {
			t.Fatalf("an invalid alert was emitted: %+v", produced)
		}
	}
}

func TestDefaultEngineRegistersEveryEvaluator(t *testing.T) {
	names := alerting.New().Names()
	if len(names) != 3 {
		t.Fatalf("evaluators = %v", names)
	}
}
