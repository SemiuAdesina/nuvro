package travel_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

func dayThreePlan() *itineraries {
	cableCar := item("cable-car", 10, place.Attraction)
	lunch := item("lunch", 13, place.Restaurant)
	garden := item("garden", 16, place.Park)
	return &itineraries{
		present: true,
		plan: itinerary.Itinerary{
			ID: "plan", TripID: tripID, Timezone: "Atlantic/Madeira", Version: 4,
			AlgorithmVersion: "itinerary-v1", Items: []itinerary.Item{cableCar, lunch, garden},
		},
	}
}

func TestMadeiraDayThreeLeaveTime(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 9, 20)
	dependencies := baseDependencies(store, newProposals(), now)
	dependencies.Routes = router{seconds: 720}
	result, err := newService(t, now, dependencies).Active(t.Context(), principal(), usecase.ActiveRequest{
		TripID: tripID, Location: &funchal,
	})
	if err != nil {
		t.Fatal(err)
	}
	snapshot := result.Snapshot
	if snapshot.Timezone != "Atlantic/Madeira" {
		t.Fatalf("timezone = %q", snapshot.Timezone)
	}
	if snapshot.CurrentDay != 3 || snapshot.Phase != domain.DuringTrip {
		t.Fatalf("day = %d phase = %q", snapshot.CurrentDay, snapshot.Phase)
	}
	if snapshot.NextItem == nil || snapshot.NextItem.ID != "cable-car" {
		t.Fatalf("next = %+v", snapshot.NextItem)
	}
	if !snapshot.Departure.LeaveBy.Equal(at(2, 9, 38)) {
		t.Fatalf("leaveBy = %v, want 09:38", snapshot.Departure.LeaveBy)
	}
	if snapshot.Departure.MinutesUntilLeave != 18 {
		t.Fatalf("minutesUntilLeave = %d", snapshot.Departure.MinutesUntilLeave)
	}
	if snapshot.Departure.Status != domain.LeaveSoon {
		t.Fatalf("status = %q", snapshot.Departure.Status)
	}
	assertDayThree(t, snapshot)
}

func assertDayThree(t *testing.T, snapshot domain.Snapshot) {
	t.Helper()
	if len(snapshot.Today) != 3 {
		t.Fatalf("today = %d items", len(snapshot.Today))
	}
	if len(snapshot.Actions) == 0 || snapshot.GeneratedAt.IsZero() {
		t.Fatal("actions or generation time missing")
	}
}

func TestActiveSnapshotWithoutLocationDoesNotClaimALeaveTime(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 9, 20)
	result, err := newService(t, now, baseDependencies(store, newProposals(), now)).
		Active(t.Context(), principal(), usecase.ActiveRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.Departure.Status != domain.RouteUnavailable {
		t.Fatalf("status = %q", result.Snapshot.Departure.Status)
	}
	if result.Snapshot.Departure.HasLeaveBy {
		t.Fatal("a leave time was invented without a location")
	}
}

func TestPartialProviderFailureStillReturnsASnapshot(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 9, 20)
	dependencies := baseDependencies(store, newProposals(), now)
	dependencies.Forecaster = forecaster{window: weather.Unavailable(weather.ProviderUnavailable)}
	dependencies.Routes = router{seconds: -1}
	dependencies.Alerts = alerts{err: errAlerts}
	result, err := newService(t, now, dependencies).Active(t.Context(), principal(), usecase.ActiveRequest{
		TripID: tripID, Location: &funchal,
	})
	if err != nil {
		t.Fatalf("a provider outage failed the whole snapshot: %v", err)
	}
	if result.Snapshot.HasWeather {
		t.Fatal("weather was claimed while the provider was down")
	}
	if result.Snapshot.NextItem == nil {
		t.Fatal("the itinerary was lost")
	}
	if result.Snapshot.Departure.Status != domain.RouteUnavailable {
		t.Fatalf("status = %q", result.Snapshot.Departure.Status)
	}
}

func TestSnapshotWorksBeforeAnyItineraryExists(t *testing.T) {
	now := at(2, 9, 20)
	result, err := newService(t, now, baseDependencies(&itineraries{}, newProposals(), now)).
		Active(t.Context(), principal(), usecase.ActiveRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.NextItem != nil || len(result.Snapshot.Today) != 0 {
		t.Fatal("items appeared without an itinerary")
	}
	if result.Snapshot.Departure.Status != domain.NoUpcomingItem {
		t.Fatalf("status = %q", result.Snapshot.Departure.Status)
	}
}

func TestPhaseBeforeAndAfterTheTrip(t *testing.T) {
	store := dayThreePlan()
	before := at(-3, 9, 0)
	early, err := newService(t, before, baseDependencies(store, newProposals(), before)).
		Active(t.Context(), principal(), usecase.ActiveRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if early.Snapshot.Phase != domain.BeforeTrip {
		t.Fatalf("phase = %q", early.Snapshot.Phase)
	}
	after := at(20, 9, 0)
	late, err := newService(t, after, baseDependencies(store, newProposals(), after)).
		Active(t.Context(), principal(), usecase.ActiveRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if late.Snapshot.Phase != domain.AfterTrip {
		t.Fatalf("phase = %q", late.Snapshot.Phase)
	}
}

func TestOnlyVisibleAlertsReachTheSnapshot(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 9, 20)
	dependencies := baseDependencies(store, newProposals(), now)
	dependencies.Alerts = alerts{stored: []alert.Alert{
		{ID: "visible", TripID: tripID, Status: alert.Unread, EffectiveAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)},
		{ID: "resolved", TripID: tripID, Status: alert.Resolved, EffectiveAt: now.Add(-time.Hour)},
		{ID: "expired", TripID: tripID, Status: alert.Unread, EffectiveAt: now.Add(-3 * time.Hour), ExpiresAt: now.Add(-time.Hour)},
		{ID: "future", TripID: tripID, Status: alert.Unread, EffectiveAt: now.Add(time.Hour)},
	}}
	result, err := newService(t, now, dependencies).Active(t.Context(), principal(), usecase.ActiveRequest{TripID: tripID})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Alerts) != 1 || result.Alerts[0].ID != "visible" {
		t.Fatalf("alerts = %v", result.Alerts)
	}
}

func TestActiveSnapshotRefusesAnonymousCallersAndTripsWithoutDestination(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 9, 20)
	dependencies := baseDependencies(store, newProposals(), now)
	if _, err := newService(t, now, dependencies).Active(t.Context(), user.Principal{},
		usecase.ActiveRequest{TripID: tripID}); err == nil {
		t.Fatal("an anonymous caller received a snapshot")
	}
	incomplete := madeiraTrip()
	incomplete.DestinationID = nil
	dependencies.Trips = trips{record: incomplete, role: trip.Owner}
	if _, err := newService(t, now, dependencies).Active(t.Context(), principal(),
		usecase.ActiveRequest{TripID: tripID}); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatal("a trip without a destination produced a snapshot")
	}
}

func TestWeatherReplanProposesButNeverApplies(t *testing.T) {
	store := dayThreePlan()
	dinner := item("dinner", 19, place.Restaurant)
	dinner.ReservationLocked = true
	dinner.Status = itinerary.Reserved
	store.plan.Items = append(store.plan.Items, dinner)
	now := at(2, 12, 0)
	proposalStore := newProposals()
	service := newService(t, now, baseDependencies(store, proposalStore, now))
	result, err := service.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: string(replan.Weather),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Proposal.Changes) != 1 {
		t.Fatalf("changes = %+v", result.Proposal.Changes)
	}
	change := result.Proposal.Changes[0]
	if change.Original.ID != "garden" || change.Proposed.PlaceID != "museum" {
		t.Fatalf("change = %+v", change)
	}
	if result.Proposal.ID == "" {
		t.Fatal("the proposal was not persisted with an identifier")
	}
	if store.replaced != nil {
		t.Fatal("the itinerary was modified before approval")
	}
	for _, item := range store.plan.Items {
		if item.ID == "dinner" && !item.StartAt.Equal(dinner.StartAt) {
			t.Fatal("the reservation moved")
		}
	}
	if len(result.Original) != 4 {
		t.Fatalf("the before state was not returned: %d items", len(result.Original))
	}
}

func TestApplyingAProposalIsAtomicAndAdvancesTheVersion(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 12, 0)
	proposalStore := newProposals()
	service := newService(t, now, baseDependencies(store, proposalStore, now))
	proposed, err := service.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: string(replan.Weather),
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.ApplyReplan(t.Context(), principal(), tripID, proposed.Proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 5 {
		t.Fatalf("version = %d, want 5", updated.Version)
	}
	if store.replaced == nil {
		t.Fatal("the itinerary was not replaced")
	}
	for _, item := range store.replaced {
		if item.ID == "garden" && item.PlaceID == "place-garden" {
			t.Fatal("the replaced activity survived")
		}
	}
	if _, err := service.ApplyReplan(t.Context(), principal(), tripID, proposed.Proposal.ID); apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatal("a proposal was applied twice")
	}
}

func TestStaleAndMismatchedProposalsAreRefused(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 12, 0)
	proposalStore := newProposals()
	service := newService(t, now, baseDependencies(store, proposalStore, now))
	proposed, err := service.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: string(replan.Weather),
	})
	if err != nil {
		t.Fatal(err)
	}
	later := newService(t, now.Add(4*time.Hour), baseDependencies(store, proposalStore, now))
	if _, err := later.ApplyReplan(t.Context(), principal(), tripID, proposed.Proposal.ID); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatalf("an expired proposal was applied: %v", err)
	}
	store.plan.Version = 9
	if _, err := service.ApplyReplan(t.Context(), principal(), tripID, proposed.Proposal.ID); apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("a proposal built on an older plan was applied: %v", err)
	}
	if _, err := service.ApplyReplan(t.Context(), principal(), tripID, "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("an unknown proposal was applied")
	}
}

func TestViewersCannotReplanOrApply(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 12, 0)
	dependencies := baseDependencies(store, newProposals(), now)
	dependencies.Trips = trips{record: madeiraTrip(), role: trip.Viewer}
	service := newService(t, now, dependencies)
	if _, err := service.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: string(replan.Weather),
	}); apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatal("a viewer proposed a replan")
	}
}

func TestReplanRejectsUnknownTriggersAndEmptyItineraries(t *testing.T) {
	now := at(2, 12, 0)
	store := dayThreePlan()
	service := newService(t, now, baseDependencies(store, newProposals(), now))
	if _, err := service.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: "vibes",
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an invented trigger was accepted")
	}
	empty := newService(t, now, baseDependencies(&itineraries{}, newProposals(), now))
	if _, err := empty.Replan(t.Context(), principal(), usecase.ReplanRequest{
		TripID: tripID, Trigger: string(replan.Weather),
	}); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatal("a trip with no itinerary was replanned")
	}
}

func TestOfflinePackIsVersionedAndMarkedStale(t *testing.T) {
	store := dayThreePlan()
	now := at(2, 12, 0)
	pack, err := newService(t, now, baseDependencies(store, newProposals(), now)).
		OfflinePack(t.Context(), principal(), tripID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if pack.Version == "" || pack.TripVersion != 4 {
		t.Fatalf("pack = %+v", pack)
	}
	if !pack.ExpiresAt.After(pack.GeneratedAt) {
		t.Fatal("the pack has no expiry")
	}
	if pack.Freshness.Status != "cached" {
		t.Fatalf("freshness = %q, want cached", pack.Freshness.Status)
	}
	if len(pack.Days) == 0 {
		t.Fatal("no itinerary in the pack")
	}
	if len(pack.Excluded) == 0 {
		t.Fatal("licensing exclusions were not declared")
	}
	for _, excluded := range pack.Excluded {
		if excluded == "" {
			t.Fatal("an empty exclusion was declared")
		}
	}
	if pack.Timezone != "Atlantic/Madeira" {
		t.Fatalf("timezone = %q", pack.Timezone)
	}
}

var errAlerts = alertError{}

type alertError struct{}

func (alertError) Error() string { return "alert store unavailable" }
