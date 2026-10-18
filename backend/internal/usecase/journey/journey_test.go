package journey_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
)

func TestMadeiraTripProducesAValidStoredItinerary(t *testing.T) {
	repository := &store{}
	subject := newHarness(t, madeiraDependencies(t, repository))
	result, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Violations) != 0 {
		t.Fatalf("the generated plan is invalid: %v", result.Violations)
	}
	if len(result.Itinerary.Items) == 0 {
		t.Fatal("no items were generated")
	}
	if result.Candidates == 0 {
		t.Fatal("candidate count not reported")
	}
	assertGrounded(t, result)
	if !repository.present || repository.saved.TripID != tripID {
		t.Fatal("the itinerary was not persisted")
	}
	if repository.saved.ID == "" || repository.saved.Timezone != "Atlantic/Madeira" {
		t.Fatalf("stored plan = %+v", repository.saved)
	}
}

func assertGrounded(t *testing.T, result usecase.GenerateResult) {
	t.Helper()
	for _, item := range result.Itinerary.Items {
		if item.PlaceID == "" {
			t.Fatalf("an item is not grounded in a resolved place: %+v", item)
		}
		if item.StartAt.Before(tripStart) || item.EndAt.After(tripStart.AddDate(0, 0, 7)) {
			t.Fatalf("an item falls outside the trip: %+v", item)
		}
	}
}

func TestTravelTimesAreRespectedInTheStoredPlan(t *testing.T) {
	repository := &store{}
	dependencies := madeiraDependencies(t, repository)
	subject := newHarness(t, dependencies)
	result, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	for _, day := range result.Itinerary.Days() {
		for index := 1; index < len(day.Items); index++ {
			gap := day.Items[index].StartAt.Sub(day.Items[index-1].EndAt)
			if gap < 10*time.Minute {
				t.Fatalf("only %v between consecutive items", gap)
			}
		}
	}
}

func TestRegenerationPreservesLockedItemsAndBumpsVersion(t *testing.T) {
	repository := &store{}
	subject := newHarness(t, madeiraDependencies(t, repository))
	first, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	locked := first.Itinerary.Items[0]
	locked.UserLocked = true
	repository.plan.Items[0] = locked
	second, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range second.Itinerary.Items {
		if item.ID == locked.ID {
			found = true
			if !item.StartAt.Equal(locked.StartAt) {
				t.Fatal("a locked item was moved during regeneration")
			}
		}
	}
	if !found {
		t.Fatal("the locked item was dropped on regeneration")
	}
	if second.Itinerary.ID != first.Itinerary.ID {
		t.Fatal("regeneration created a second itinerary for one trip")
	}
	if second.Itinerary.Version <= first.Itinerary.Version {
		t.Fatal("regeneration did not advance the version")
	}
}

func TestPaceOverrideChangesDensity(t *testing.T) {
	relaxed, err := newHarness(t, madeiraDependencies(t, &store{})).service.
		Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID, Pace: "relaxed"})
	if err != nil {
		t.Fatal(err)
	}
	packed, err := newHarness(t, madeiraDependencies(t, &store{})).service.
		Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID, Pace: "packed"})
	if err != nil {
		t.Fatal(err)
	}
	if len(relaxed.Itinerary.Days()) <= len(packed.Itinerary.Days()) {
		t.Fatalf("relaxed spread over %d days, packed over %d", len(relaxed.Itinerary.Days()), len(packed.Itinerary.Days()))
	}
	for _, day := range packed.Itinerary.Days() {
		if len(day.Items) > 4 {
			t.Fatalf("packed scheduled %d items in one day", len(day.Items))
		}
	}
	for _, day := range relaxed.Itinerary.Days() {
		if len(day.Items) > 2 {
			t.Fatalf("relaxed scheduled %d items in one day", len(day.Items))
		}
	}
}

func TestATripWithoutDestinationOrDatesCannotBePlanned(t *testing.T) {
	dependencies := madeiraDependencies(t, &store{})
	incomplete := tripRecord(t, 7)
	incomplete.DestinationID = nil
	dependencies.Trips = trips{record: incomplete, role: trip.Owner}
	_, err := newHarness(t, dependencies).service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatalf("error = %v", err)
	}
}

func TestViewersCannotGenerateOrEdit(t *testing.T) {
	repository := &store{}
	dependencies := madeiraDependencies(t, repository)
	dependencies.Trips = trips{record: tripRecord(t, 7), role: trip.Viewer}
	subject := newHarness(t, dependencies)
	if _, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID}); apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("a viewer generated an itinerary: %v", err)
	}
}

func TestUnauthenticatedRequestsAreRefused(t *testing.T) {
	subject := newHarness(t, madeiraDependencies(t, &store{}))
	if _, err := subject.service.Generate(t.Context(), user.Principal{}, usecase.GenerateRequest{TripID: tripID}); err == nil {
		t.Fatal("an anonymous caller generated an itinerary")
	}
	if _, err := subject.service.Read(t.Context(), user.Principal{}, tripID); err == nil {
		t.Fatal("an anonymous caller read an itinerary")
	}
}

func TestReadingBeforeGenerationReportsNotFound(t *testing.T) {
	subject := newHarness(t, madeiraDependencies(t, &store{}))
	if _, err := subject.service.Read(t.Context(), principal(), tripID); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("error = %v", err)
	}
}

func TestNoPlaceEvidenceFailsHonestlyRatherThanInventingAPlan(t *testing.T) {
	dependencies := madeiraDependencies(t, &store{})
	dependencies.Planner = failingPlanner{}
	_, err := newHarness(t, dependencies).service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatalf("error = %v", err)
	}
}

func TestItemEditsRespectStatusRulesAndOwnership(t *testing.T) {
	repository := &store{}
	subject := newHarness(t, madeiraDependencies(t, repository))
	result, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	target := result.Itinerary.Items[0]
	locked := true
	reserved := string(itinerary.Reserved)
	updated, err := subject.service.UpdateItem(t.Context(), principal(), usecase.ItemUpdate{
		TripID: tripID, ItemID: target.ID, Status: &reserved, UserLocked: &locked,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != itinerary.Reserved || !updated.UserLocked {
		t.Fatalf("update not applied: %+v", updated)
	}
	planned := string(itinerary.Planned)
	if _, err := subject.service.UpdateItem(t.Context(), principal(), usecase.ItemUpdate{
		TripID: tripID, ItemID: target.ID, Status: &planned,
	}); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatalf("an illegal status change was accepted: %v", err)
	}
	if _, err := subject.service.UpdateItem(t.Context(), principal(), usecase.ItemUpdate{
		TripID: tripID, ItemID: "1c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99",
	}); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("an unknown item was updated")
	}
}

func TestInvalidTimeEditsAreRejected(t *testing.T) {
	repository := &store{}
	subject := newHarness(t, madeiraDependencies(t, repository))
	result, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	target := result.Itinerary.Items[0]
	end := target.StartAt.Add(-time.Hour)
	if _, err := subject.service.UpdateItem(t.Context(), principal(), usecase.ItemUpdate{
		TripID: tripID, ItemID: target.ID, EndAt: &end,
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("an item ending before it starts was accepted: %v", err)
	}
}

func TestBudgetUsesTheTripCurrencyAndItineraryActivities(t *testing.T) {
	repository := &store{}
	dependencies := madeiraDependencies(t, repository)
	dependencies.Accountant = budgetengine.New(costs{values: []budgetengine.Estimate{{
		Category: budgetdomain.Stay, Minimum: amount(t, 90_000, "EUR"), Expected: amount(t, 100_000, "EUR"),
		Maximum: amount(t, 120_000, "EUR"), Confidence: budgetdomain.Medium, Source: "fixture", ObservedAt: now,
	}}}, nil, nil, clock.NewFixed(now))
	subject := newHarness(t, dependencies)
	if _, err := subject.service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID}); err != nil {
		t.Fatal(err)
	}
	result, err := subject.service.Budget(t.Context(), principal(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Currency.String() != "EUR" {
		t.Fatalf("currency = %q", result.Currency)
	}
	if !result.HasEstimate || result.Estimated.Expected.Minor() != 100_000 {
		t.Fatalf("estimate = %+v", result.Estimated)
	}
	if result.Lines[0].Category != budgetdomain.Stay {
		t.Fatalf("lines = %+v", result.Lines)
	}
}

func TestBudgetWithoutAnAccountantIsHonest(t *testing.T) {
	dependencies := madeiraDependencies(t, &store{})
	dependencies.Accountant = nil
	if _, err := newHarness(t, dependencies).service.Budget(t.Context(), principal(), tripID); apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatal("an unconfigured budget engine reported success")
	}
}

func TestNearbyExcludesPlacesThatCannotFitBeforeTheNextItem(t *testing.T) {
	repository := &store{}
	dependencies := madeiraDependencies(t, repository)
	dinner := itinerary.Item{
		ID: "dinner", TripID: tripID, Day: 1,
		StartAt: time.Date(2026, 10, 14, 19, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 10, 14, 21, 0, 0, 0, time.UTC),
		Kind:    itinerary.Meal, Title: "Dinner in Old Town", PlaceID: "old-town-dinner",
		Status: itinerary.Reserved, Source: itinerary.Traveller, ReservationLocked: true,
	}
	repository.plan = itinerary.Itinerary{TripID: tripID, Timezone: "Atlantic/Madeira", Items: []itinerary.Item{dinner}}
	repository.present = true
	dependencies.Finder = finder{places: []place.Place{
		madeiraPlace("quick-cafe", place.Cafe, 45),
		madeiraPlace("long-levada", place.Experience, 180),
	}}
	result, err := newHarness(t, dependencies).service.Nearby(t.Context(), principal(), usecase.NearbyRequest{
		TripID: tripID, Location: funchal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.NextItem == nil || result.NextItem.ID != "dinner" {
		t.Fatalf("next item = %+v", result.NextItem)
	}
	if result.AvailableMinutes != 140 {
		t.Fatalf("available minutes = %d, want 140", result.AvailableMinutes)
	}
	if result.Places[0].Place.ID != "quick-cafe" || !result.Places[0].Fits {
		t.Fatalf("ranking = %+v", result.Places[0])
	}
	for _, ranked := range result.Places {
		if ranked.Place.ID == "long-levada" && ranked.Fits {
			t.Fatal("a three hour walk was said to fit two hours twenty")
		}
	}
}

func TestNearbyRequiresAValidLocationAndConfiguredProvider(t *testing.T) {
	dependencies := madeiraDependencies(t, &store{})
	subject := newHarness(t, dependencies)
	if _, err := subject.service.Nearby(t.Context(), principal(), usecase.NearbyRequest{
		TripID: tripID, Location: geoPointOutOfRange(),
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an impossible coordinate was accepted")
	}
	dependencies.Finder = nil
	if _, err := newHarness(t, dependencies).service.Nearby(t.Context(), principal(), usecase.NearbyRequest{
		TripID: tripID, Location: funchal,
	}); apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatal("an unconfigured place provider reported success")
	}
	dependencies.Finder = finder{err: errors.New("provider down")}
	if _, err := newHarness(t, dependencies).service.Nearby(t.Context(), principal(), usecase.NearbyRequest{
		TripID: tripID, Location: funchal,
	}); apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatal("a failing place provider reported success")
	}
}

func TestWeatherBeyondTheHorizonStillProducesAPlan(t *testing.T) {
	dependencies := madeiraDependencies(t, &store{})
	dependencies.Forecaster = forecaster{window: weather.Unavailable(weather.BeyondForecast)}
	result, err := newHarness(t, dependencies).service.Generate(t.Context(), principal(), usecase.GenerateRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Itinerary.Items) == 0 {
		t.Fatal("a trip beyond the forecast horizon produced no plan")
	}
	for _, item := range result.Itinerary.Items {
		if item.WeatherImpact != weather.UnknownImpact && item.WeatherImpact != weather.NoImpact {
			t.Fatalf("weather impact was claimed without a forecast: %q", item.WeatherImpact)
		}
	}
}

type failingPlanner struct{}

func (failingPlanner) Generate(context.Context, itineraryengine.Request) (itineraryengine.Result, error) {
	return itineraryengine.Result{}, itineraryengine.ErrPlacesUnavailable
}

func geoPointOutOfRange() geo.Point { return geo.Point{Latitude: 200, Longitude: 0} }
