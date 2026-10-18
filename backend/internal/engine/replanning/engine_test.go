package replanning_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var day = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

func at(hour int) time.Time { return day.Add(time.Duration(hour) * time.Hour) }

type finder struct {
	places     []place.Place
	err        error
	categories []place.Category
}

func (f *finder) Nearby(_ context.Context, _ geo.Point, categories []place.Category, _ string, _ int) ([]place.Place, error) {
	f.categories = categories
	return f.places, f.err
}

type router struct{ seconds int }

func (r router) Compute(_ context.Context, origin, destination geo.Point, mode route.Mode, _ time.Time) route.Evidence {
	if r.seconds < 0 {
		return route.UnavailableEvidence(origin, destination, mode)
	}
	return route.Evidence{
		Status: route.Found, Origin: origin, Destination: destination, Mode: mode,
		DurationSeconds: r.seconds, DistanceMeters: r.seconds, Confidence: route.High,
	}
}

type impacts struct{ byPlace map[string]weather.Impact }

func (i impacts) Impact(_ weather.Window, _ time.Time, sensitivity weather.Sensitivity) weather.Impact {
	if impact, found := i.byPlace[string(sensitivity)]; found {
		return impact
	}
	return weather.NoImpact
}

type interpreter struct {
	intent engine.Intent
	err    error
	seen   string
}

func (i *interpreter) Interpret(_ context.Context, message string) (engine.Intent, error) {
	i.seen = message
	return i.intent, i.err
}

func candidate(id string, minutes int) place.Place {
	return place.Place{
		ID: id, Name: id, Category: place.Museum, Location: funchal, Status: place.Open,
		Rating: 4.5, ReviewCount: 200, TypicalMinutes: minutes,
	}
}

func item(id string, dayNumber, startHour int, category place.Category) itinerary.Item {
	return itinerary.Item{
		ID: id, TripID: "trip", Day: dayNumber,
		StartAt: at(startHour), EndAt: at(startHour + 2),
		Kind: itinerary.Activity, Title: id, PlaceID: "place-" + id, Category: category,
		Location: funchal, Status: itinerary.Planned, Source: itinerary.Generated,
	}
}

func plan(items ...itinerary.Item) itinerary.Itinerary {
	return itinerary.Itinerary{
		TripID: "trip", Timezone: "Atlantic/Madeira", Version: 4,
		AlgorithmVersion: "itinerary-v1", Items: items,
	}
}

func request(trigger replan.Trigger, items ...itinerary.Item) engine.Request {
	return engine.Request{
		TripID: "trip", Trigger: trigger, Itinerary: plan(items...),
		Weather: weather.Window{Availability: weather.Available},
		Now:     at(12), Currency: "EUR", Timezone: "Atlantic/Madeira",
	}
}

func newEngine(t *testing.T, source *finder, routes engine.RouteFinder, weatherReader engine.WeatherReader, reader engine.Interpreter) *engine.Engine {
	t.Helper()
	return engine.New(source, routes, weatherReader, reader, clock.NewFixed(at(12)))
}

func outdoorRain() impacts {
	return impacts{byPlace: map[string]weather.Impact{string(weather.Outdoor): weather.HighImpact}}
}

func TestRainReplacesTheOutdoorActivityAndLeavesDinnerAlone(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	dinner := item("dinner", 3, 19, place.Restaurant)
	dinner.ReservationLocked = true
	dinner.Status = itinerary.Reserved
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden, dinner))
	if len(proposal.Changes) != 1 {
		t.Fatalf("changes = %d, want the single minimal change", len(proposal.Changes))
	}
	change := proposal.Changes[0]
	if change.Kind != replan.Replace || change.Original.ID != "garden" {
		t.Fatalf("change = %+v", change)
	}
	if change.Proposed.PlaceID != "museum" || change.Proposed.Category != place.Museum {
		t.Fatalf("replacement is not the indoor candidate: %+v", change.Proposed)
	}
	if !change.Proposed.StartAt.Equal(garden.StartAt) {
		t.Fatal("the replacement moved the time slot")
	}
	if len(proposal.Reasons) == 0 || proposal.Reasons[0] != replan.WeatherConflict {
		t.Fatalf("reasons = %v", proposal.Reasons)
	}
	if !proposal.Valid() {
		t.Fatalf("proposal is invalid: %v", proposal.Violations)
	}
	if proposal.Applied() {
		t.Fatal("a proposal must not apply itself")
	}
	if proposal.ItineraryVersion != 4 {
		t.Fatalf("version = %d", proposal.ItineraryVersion)
	}
}

func TestLockedAndReservedItemsAreNeverProposedForChange(t *testing.T) {
	locked := item("locked-garden", 3, 15, place.Park)
	locked.UserLocked = true
	reserved := item("reserved-tour", 3, 17, place.Experience)
	reserved.ReservationLocked = true
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	proposal := newEngine(t, source, router{seconds: 600}, impacts{byPlace: map[string]weather.Impact{
		string(weather.Outdoor): weather.HighImpact, string(weather.Sheltered): weather.HighImpact,
	}}, nil).Propose(t.Context(), request(replan.Weather, locked, reserved))
	if len(proposal.Changes) != 0 {
		t.Fatalf("locked items were changed: %+v", proposal.Changes)
	}
	if len(proposal.Preserved) == 0 {
		t.Fatal("preserved constraints were not reported")
	}
}

func TestNoCandidateFallsBackToMovingOrRemoving(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	tomorrow := item("tomorrow-visit", 4, 16, place.Museum)
	source := &finder{}
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden, tomorrow))
	if len(proposal.Changes) != 1 {
		t.Fatalf("changes = %+v", proposal.Changes)
	}
	change := proposal.Changes[0]
	if change.Kind != replan.Move || change.Proposed.Day <= garden.Day {
		t.Fatalf("expected a move to a later day, got %+v", change)
	}
	if change.Proposed.ID != garden.ID {
		t.Fatal("a moved item changed identity")
	}
	single := newEngine(t, &finder{}, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden))
	if len(single.Changes) != 1 || single.Changes[0].Kind != replan.Remove {
		t.Fatalf("with nowhere to move, expected removal: %+v", single.Changes)
	}
}

func TestUnresolvableCandidatesAreNeverInvented(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	failing := &finder{err: errors.New("place provider down")}
	proposal := newEngine(t, failing, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden))
	for _, change := range proposal.Changes {
		if change.Proposed != nil && change.Proposed.PlaceID != garden.PlaceID {
			t.Fatalf("a place was invented without provider evidence: %+v", change.Proposed)
		}
	}
	unconfigured := engine.New(nil, nil, outdoorRain(), nil, clock.NewFixed(at(12))).
		Propose(t.Context(), request(replan.Weather, garden))
	for _, change := range unconfigured.Changes {
		if change.Kind == replan.Replace {
			t.Fatal("a replacement was produced with no place provider")
		}
	}
}

func TestClosedPlaceCandidatesAreExcluded(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	closed := candidate("closed-museum", 90)
	closed.Status = place.Closed
	open := candidate("open-museum", 90)
	source := &finder{places: []place.Place{closed, open}}
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden))
	if len(proposal.Changes) != 1 || proposal.Changes[0].Proposed.PlaceID != "open-museum" {
		t.Fatalf("a closed candidate was proposed: %+v", proposal.Changes)
	}
}

func TestCandidatesClosedAtTheSlotHourAreExcluded(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	evening := candidate("evening-museum", 90)
	evening.Hours = place.OpeningHours{
		Known: true, Intervals: []place.Interval{{Weekday: garden.StartAt.Weekday(), Opens: 20 * 60, Closes: 23 * 60}},
	}
	source := &finder{places: []place.Place{evening}}
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden))
	for _, change := range proposal.Changes {
		if change.Proposed != nil && change.Proposed.PlaceID == "evening-museum" {
			t.Fatal("a place closed at the proposed hour was suggested")
		}
	}
}

func TestReducedWalkingPushesFarCandidatesOut(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	far := candidate("far-museum", 90)
	far.Location = geo.Point{Latitude: 33.4, Longitude: -16.9}
	near := candidate("near-museum", 90)
	source := &finder{places: []place.Place{far, near}}
	input := request(replan.Weather, garden)
	input.ReducedWalking = true
	input.MaximumWalkMeters = 1500
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).Propose(t.Context(), input)
	if len(proposal.Changes) != 1 || proposal.Changes[0].Proposed.PlaceID != "near-museum" {
		t.Fatalf("accessibility was ignored: %+v", proposal.Changes)
	}
	preserved := false
	for _, constraint := range proposal.Preserved {
		if constraint == replan.Accessibility {
			preserved = true
		}
	}
	if !preserved {
		t.Fatal("the accessibility constraint was not reported as preserved")
	}
}

func TestTiredTravellerLosesTheLastFlexibleActivity(t *testing.T) {
	first := item("morning", 3, 14, place.Museum)
	last := item("evening", 3, 17, place.Attraction)
	source := &finder{places: []place.Place{candidate("other", 60)}}
	proposal := newEngine(t, source, router{seconds: 600}, impacts{}, nil).
		Propose(t.Context(), request(replan.Tired, first, last))
	if len(proposal.Changes) != 1 {
		t.Fatalf("changes = %+v", proposal.Changes)
	}
	if proposal.Changes[0].Kind != replan.Remove || proposal.Changes[0].Original.ID != "evening" {
		t.Fatalf("a tired traveller lost the wrong item: %+v", proposal.Changes[0])
	}
	if proposal.Changes[0].Reasons[0] != replan.TravelerRequestedLowerPace {
		t.Fatalf("reason = %v", proposal.Changes[0].Reasons)
	}
}

func TestMissedActivityIsReportedWithItsOwnReason(t *testing.T) {
	missed := item("missed", 3, 11, place.Museum)
	missed.EndAt = at(13)
	source := &finder{places: []place.Place{candidate("later-museum", 90)}}
	proposal := newEngine(t, source, router{seconds: 600}, impacts{}, nil).
		Propose(t.Context(), request(replan.Missed, missed))
	if len(proposal.Changes) != 1 {
		t.Fatalf("changes = %+v", proposal.Changes)
	}
	if proposal.Changes[0].Reasons[0] != replan.MissedStartTime {
		t.Fatalf("reason = %v", proposal.Changes[0].Reasons)
	}
}

func TestNoConflictProducesNoChanges(t *testing.T) {
	sunny := item("beach", 3, 15, place.Beach)
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	proposal := newEngine(t, source, router{seconds: 600}, impacts{}, nil).
		Propose(t.Context(), request(replan.Weather, sunny))
	if len(proposal.Changes) != 0 {
		t.Fatalf("changes were proposed with no conflict: %+v", proposal.Changes)
	}
	if len(proposal.Reasons) != 1 || proposal.Reasons[0] != replan.NoAlternativeFound {
		t.Fatalf("reasons = %v", proposal.Reasons)
	}
}

func TestBudgetDeltaIsReportedInTheTripCurrency(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	garden.EstimatedCostMinor = 1000
	garden.CostCurrency = "EUR"
	replacement := candidate("museum", 90)
	source := &finder{places: []place.Place{replacement}}
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden))
	if !proposal.HasBudgetDelta {
		t.Fatal("no budget delta was reported")
	}
	if proposal.BudgetDelta.Currency().String() != "EUR" {
		t.Fatalf("currency = %q", proposal.BudgetDelta.Currency())
	}
	if proposal.BudgetDelta.Minor() != -1000 {
		t.Fatalf("delta = %d, want the removed cost", proposal.BudgetDelta.Minor())
	}
}

func TestInvalidInterpreterOutputIsDiscarded(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	for name, reader := range map[string]*interpreter{
		"provider failure":  {err: errors.New("model unavailable")},
		"invented trigger":  {intent: engine.Intent{Trigger: "vibes"}},
		"invented category": {intent: engine.Intent{Trigger: replan.Weather, Categories: []place.Category{"spaceport"}}},
	} {
		input := request(replan.Weather, garden)
		input.Message = "we are soaked"
		proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), reader).Propose(t.Context(), input)
		if proposal.Trigger != replan.Weather {
			t.Fatalf("%s changed the trigger to %q", name, proposal.Trigger)
		}
		if len(proposal.Changes) != 1 || proposal.Changes[0].Proposed.PlaceID != "museum" {
			t.Fatalf("%s corrupted the proposal: %+v", name, proposal.Changes)
		}
	}
}

func TestProposalCarriesAnExpiryAndIsDeterministic(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	subject := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil)
	first := subject.Propose(t.Context(), request(replan.Weather, garden))
	if !first.ExpiresAt.After(first.CreatedAt) {
		t.Fatal("the proposal has no expiry")
	}
	if !first.Stale(first.ExpiresAt.Add(time.Minute)) || first.Stale(first.CreatedAt) {
		t.Fatal("staleness is wrong")
	}
	for range 10 {
		again := subject.Propose(t.Context(), request(replan.Weather, garden))
		if len(again.Changes) != len(first.Changes) ||
			again.Changes[0].Proposed.PlaceID != first.Changes[0].Proposed.PlaceID {
			t.Fatal("proposals are not deterministic")
		}
	}
}

func TestApplyProducesThePlanTheProposalDescribes(t *testing.T) {
	garden := item("garden", 3, 15, place.Park)
	dinner := item("dinner", 3, 19, place.Restaurant)
	source := &finder{places: []place.Place{candidate("museum", 90)}}
	original := plan(garden, dinner)
	proposal := newEngine(t, source, router{seconds: 600}, outdoorRain(), nil).
		Propose(t.Context(), request(replan.Weather, garden, dinner))
	updated := engine.Apply(original, proposal)
	if len(updated.Items) != 2 {
		t.Fatalf("items = %d", len(updated.Items))
	}
	for _, value := range updated.Items {
		if value.ID == garden.ID && value.PlaceID == garden.PlaceID {
			t.Fatal("the replaced item survived unchanged")
		}
	}
	found := false
	for _, value := range updated.Items {
		if value.ID == dinner.ID && value.StartAt.Equal(dinner.StartAt) {
			found = true
		}
	}
	if !found {
		t.Fatal("dinner was disturbed")
	}
}
