package itinerary_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

var day = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

func at(hour, minute int) time.Time {
	return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

func item(id string, startHour, endHour int) itinerary.Item {
	return itinerary.Item{
		ID: id, TripID: "trip", Day: 1, StartAt: at(startHour, 0), EndAt: at(endHour, 0),
		Kind: itinerary.Activity, Title: id, PlaceID: "place-" + id,
		Status: itinerary.Planned, Source: itinerary.Generated,
	}
}

func constraints() itinerary.Constraints {
	return itinerary.Constraints{Start: day, End: day.AddDate(0, 0, 1), BufferMinutes: 15}
}

func codes(violations []itinerary.Violation) map[itinerary.ViolationCode]int {
	found := map[itinerary.ViolationCode]int{}
	for _, violation := range violations {
		found[violation.Code]++
	}
	return found
}

func TestAValidDayPasses(t *testing.T) {
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 11), item("b", 13, 15)}}
	if violations := itinerary.Validate(plan, constraints()); len(violations) != 0 {
		t.Fatalf("valid plan reported %v", violations)
	}
}

func TestOverlappingItemsAreRejected(t *testing.T) {
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 12), item("b", 11, 13)}}
	if codes(itinerary.Validate(plan, constraints()))[itinerary.Overlap] != 1 {
		t.Fatal("overlapping items were accepted")
	}
}

func TestTravelTimeMustFitBetweenItems(t *testing.T) {
	second := item("b", 11, 13)
	second.RouteFromPrevious = &route.Evidence{
		Status: route.Found, Mode: route.Walk, DurationSeconds: 3600, DistanceMeters: 4000,
	}
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 11), second}}
	if codes(itinerary.Validate(plan, constraints()))[itinerary.TravelTime] != 1 {
		t.Fatal("an impossible transfer was accepted")
	}
	relaxed := second
	relaxed.StartAt = at(12, 30)
	relaxed.EndAt = at(14, 0)
	fits := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 11), relaxed}}
	if len(itinerary.Validate(fits, constraints())) != 0 {
		t.Fatal("a feasible transfer was rejected")
	}
}

func TestUnknownRouteDoesNotFailTheDay(t *testing.T) {
	second := item("b", 11, 13)
	second.RouteFromPrevious = &route.Evidence{Status: route.Unavailable, Mode: route.Walk, DurationSeconds: -1}
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 10), second}}
	if codes(itinerary.Validate(plan, constraints()))[itinerary.TravelTime] != 0 {
		t.Fatal("unknown route evidence was treated as a failure")
	}
}

func TestClosedPlacesAreHardFailures(t *testing.T) {
	rules := constraints()
	rules.Openings = map[string]place.OpeningHours{
		"place-a": {Known: true, Intervals: []place.Interval{{Weekday: day.Weekday(), Opens: 14 * 60, Closes: 18 * 60}}},
	}
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 11)}}
	if codes(itinerary.Validate(plan, rules))[itinerary.PlaceClosed] != 1 {
		t.Fatal("an item scheduled outside opening hours was accepted")
	}
	open := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 15, 17)}}
	if len(itinerary.Validate(open, rules)) != 0 {
		t.Fatal("an item inside opening hours was rejected")
	}
}

func TestUnknownOpeningHoursAreNotTreatedAsClosed(t *testing.T) {
	rules := constraints()
	rules.Openings = map[string]place.OpeningHours{"place-a": {Known: false}}
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{item("a", 9, 11)}}
	if codes(itinerary.Validate(plan, rules))[itinerary.PlaceClosed] != 0 {
		t.Fatal("unknown opening hours were treated as closed")
	}
}

func TestItemsMustStayInsideTheTripWindow(t *testing.T) {
	outside := item("a", 9, 11)
	outside.StartAt = day.AddDate(0, 0, -1)
	outside.EndAt = day.AddDate(0, 0, -1).Add(2 * time.Hour)
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{outside}}
	if codes(itinerary.Validate(plan, constraints()))[itinerary.OutsideTripDates] != 1 {
		t.Fatal("an item before the trip was accepted")
	}
}

func TestWeatherBlockedAndUnresolvedPlacesAreReported(t *testing.T) {
	blocked := item("a", 9, 11)
	blocked.WeatherImpact = weather.HighImpact
	unresolved := item("b", 13, 15)
	unresolved.PlaceID = ""
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{blocked, unresolved}}
	found := codes(itinerary.Validate(plan, constraints()))
	if found[itinerary.WeatherBlocked] != 1 || found[itinerary.MissingPlace] != 1 {
		t.Fatalf("violations = %v", found)
	}
}

func TestMovingALockedItemIsReported(t *testing.T) {
	original := item("a", 9, 11)
	rules := constraints()
	rules.Locked = map[string]itinerary.Item{"a": original}
	moved := original
	moved.StartAt = at(14, 0)
	moved.EndAt = at(16, 0)
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{moved}}
	if codes(itinerary.Validate(plan, rules))[itinerary.LockedItemMoved] != 1 {
		t.Fatal("a locked item was silently moved")
	}
	unchanged := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{original}}
	if len(itinerary.Validate(unchanged, rules)) != 0 {
		t.Fatal("an unchanged locked item was reported")
	}
}

func TestInvalidItemsAreReportedNotPanicked(t *testing.T) {
	broken := itinerary.Item{ID: "a", TripID: "trip", Day: 1, Title: "", Kind: itinerary.Activity, Status: itinerary.Planned}
	plan := itinerary.Itinerary{TripID: "trip", Items: []itinerary.Item{broken}}
	if codes(itinerary.Validate(plan, constraints()))[itinerary.InvalidItem] != 1 {
		t.Fatal("an invalid item was accepted")
	}
}

func TestDaysAreOrderedAndLockedItemsExposed(t *testing.T) {
	locked := item("c", 16, 17)
	locked.UserLocked = true
	plan := itinerary.Itinerary{
		TripID: "trip", Timezone: "Atlantic/Madeira",
		Items: []itinerary.Item{item("b", 13, 15), locked, item("a", 9, 11)},
	}
	days := plan.Days()
	if len(days) != 1 || len(days[0].Items) != 3 {
		t.Fatalf("days = %v", days)
	}
	if days[0].Items[0].ID != "a" || days[0].Items[2].ID != "c" {
		t.Fatal("items were not ordered by start time")
	}
	if len(plan.LockedItems()) != 1 {
		t.Fatal("locked items were not exposed")
	}
}

func TestStatusTransitions(t *testing.T) {
	if !itinerary.Planned.CanTransitionTo(itinerary.Reserved) || !itinerary.Reserved.CanTransitionTo(itinerary.Completed) {
		t.Fatal("a legitimate transition was refused")
	}
	if itinerary.Completed.CanTransitionTo(itinerary.Planned) || itinerary.Cancelled.CanTransitionTo(itinerary.Reserved) {
		t.Fatal("a terminal state was reopened")
	}
	if itinerary.Planned.CanTransitionTo("invented") {
		t.Fatal("an invented status was accepted")
	}
}
