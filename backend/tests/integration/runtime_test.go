//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
)

func alertFixture(t *testing.T, tripID, fingerprint string, expiresAt time.Time) alert.Alert {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	return alert.Alert{
		ID: newIdentifier(t), TripID: tripID, Type: alert.WeatherImpact, Severity: alert.Attention,
		TitleCode: "alert.weather_impact", MessageData: map[string]string{"activity": "Garden"},
		RelatedItemID: "item-1", Action: alert.AdjustPlan, Status: alert.Unread,
		Fingerprint: fingerprint, CreatedAt: now, EffectiveAt: now.Add(-time.Hour), ExpiresAt: expiresAt,
		Freshness: evidence.Freshness{ObservedAt: now, ValidUntil: now.Add(time.Hour), Status: evidence.Live},
	}
}

func TestAlertsRoundTripAndDeduplicateOnFingerprint(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	future := time.Now().UTC().Add(3 * time.Hour).Truncate(time.Second)

	first := alertFixture(t, tripID, "fingerprint-weather-1", future)
	saved, err := store.SaveAlert(t.Context(), first)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Type != alert.WeatherImpact || saved.MessageData["activity"] != "Garden" {
		t.Fatalf("saved = %+v", saved)
	}
	if saved.Freshness.Status != evidence.Live || saved.Freshness.ObservedAt.IsZero() {
		t.Fatalf("freshness lost: %+v", saved.Freshness)
	}

	repeat := alertFixture(t, tripID, "fingerprint-weather-1", future)
	repeat.Severity = alert.Important
	if _, err := store.SaveAlert(t.Context(), repeat); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Alerts(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Fatalf("the same fingerprint produced %d rows", len(stored))
	}
	if stored[0].Severity != alert.Important {
		t.Fatalf("the repeat evaluation did not refresh severity: %+v", stored[0])
	}
	if stored[0].ID != saved.ID {
		t.Fatal("deduplication created a new identifier")
	}
}

func TestAlertStatusAndExpiryPersist(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	now := time.Now().UTC()

	live := alertFixture(t, tripID, "fingerprint-live", now.Add(3*time.Hour))
	stale := alertFixture(t, tripID, "fingerprint-stale", now.Add(-time.Minute))
	stale.EffectiveAt = now.Add(-2 * time.Hour)
	for _, value := range []alert.Alert{live, stale} {
		if _, err := store.SaveAlert(t.Context(), value); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.UpdateAlertStatus(t.Context(), live.ID, alert.Read); err != nil {
		t.Fatal(err)
	}
	reread, err := store.Alert(t.Context(), live.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reread.Status != alert.Read {
		t.Fatalf("status = %q", reread.Status)
	}
	if err := store.ExpireAlerts(t.Context(), tripID, now); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Alerts(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range stored {
		switch value.Fingerprint {
		case "fingerprint-stale":
			if value.Status != alert.Expired {
				t.Fatalf("a past alert was not expired: %+v", value)
			}
		case "fingerprint-live":
			if value.Status != alert.Read {
				t.Fatalf("a live alert was expired: %+v", value)
			}
		}
	}
}

func TestDatabaseRejectsInventedAlertVocabulary(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	broken := alertFixture(t, tripID, "fingerprint-broken", time.Now().UTC().Add(time.Hour))
	broken.Severity = "critical"
	if _, err := store.SaveAlert(t.Context(), broken); err == nil {
		t.Fatal("an invented severity was stored")
	}
}

func TestReplanProposalsPersistWithAuditFields(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	now := time.Now().UTC().Truncate(time.Second)
	original := itineraryItem(t, tripID, 1, 15, 17, false)
	proposed := original
	proposed.Title = "Madeira Story Centre"
	proposed.PlaceID = ""
	delta, err := money.New(-1500, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	proposal := replan.Proposal{
		ID: newIdentifier(t), TripID: tripID, Trigger: replan.Weather, ItineraryVersion: 1,
		Changes: []replan.Change{{
			Kind: replan.Replace, Original: &original, Proposed: &proposed,
			Reasons: []replan.Reason{replan.WeatherConflict},
		}},
		Reasons: []replan.Reason{replan.WeatherConflict}, Preserved: []replan.Constraint{replan.LockedItems},
		BudgetDelta: delta, HasBudgetDelta: true, Summary: "Replaces the garden",
		CreatedAt: now, ExpiresAt: now.Add(2 * time.Hour),
	}
	if err := store.SaveProposal(t.Context(), proposal); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Proposal(t.Context(), proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Trigger != replan.Weather || stored.ItineraryVersion != 1 {
		t.Fatalf("stored = %+v", stored)
	}
	if len(stored.Changes) != 1 || stored.Changes[0].Proposed.Title != "Madeira Story Centre" {
		t.Fatalf("changes lost: %+v", stored.Changes)
	}
	if !stored.HasBudgetDelta || stored.BudgetDelta.Minor() != -1500 {
		t.Fatalf("budget delta lost: %+v", stored.BudgetDelta)
	}
	if stored.BudgetDelta.Currency().String() != "EUR" {
		t.Fatal("the currency was lost")
	}
	if stored.Applied() {
		t.Fatal("a fresh proposal was marked applied")
	}
	owner := ownerOf(t, pool, tripID)
	if err := store.MarkProposalApplied(t.Context(), proposal.ID, owner, now); err != nil {
		t.Fatal(err)
	}
	applied, err := store.Proposal(t.Context(), proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !applied.Applied() {
		t.Fatal("the applied timestamp was not recorded")
	}
}

func TestReplacingItemsIsVersionCheckedAndAtomic(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "Atlantic/Madeira",
		AlgorithmVersion: "itinerary-v1", GeneratedAt: time.Now().UTC(),
		Items: []itinerary.Item{itineraryItem(t, tripID, 1, 9, 11, false)},
	}
	if _, err := store.SaveItinerary(t.Context(), plan, []place.Place{referencedPlace()}); err != nil {
		t.Fatal(err)
	}
	replacement := []itinerary.Item{itineraryItem(t, tripID, 2, 14, 16, false)}
	updated, err := store.ReplaceItineraryItems(t.Context(), tripID, 1, replacement)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 {
		t.Fatalf("version = %d", updated.Version)
	}
	stored, err := store.Itinerary(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Items) != 1 || stored.Items[0].Day != 2 {
		t.Fatalf("items = %+v", stored.Items)
	}
	if _, err := store.ReplaceItineraryItems(t.Context(), tripID, 1, replacement); err == nil {
		t.Fatal("a stale version was accepted")
	}
	unchanged, err := store.Itinerary(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Version != 2 || len(unchanged.Items) != 1 {
		t.Fatalf("the rejected write left state behind: %+v", unchanged)
	}
}

func TestTripsNeedingEvaluationOnlyReturnsPlannedActiveTrips(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	until := from.Add(24 * time.Hour)

	found, err := store.TripsNeedingEvaluation(t.Context(), from, until, 50)
	if err != nil {
		t.Fatal(err)
	}
	if containsTrip(found, tripID) {
		t.Fatal("a trip with no itinerary was queued for evaluation")
	}
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "Atlantic/Madeira",
		AlgorithmVersion: "itinerary-v1", GeneratedAt: time.Now().UTC(),
		Items: []itinerary.Item{itineraryItem(t, tripID, 1, 9, 11, false)},
	}
	if _, err := store.SaveItinerary(t.Context(), plan, []place.Place{referencedPlace()}); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), "UPDATE trips SET status = 'active' WHERE id = $1", tripID); err != nil {
		t.Fatal(err)
	}
	found, err = store.TripsNeedingEvaluation(t.Context(), from, until, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !containsTrip(found, tripID) {
		t.Fatal("an active trip with an itinerary was not queued")
	}
	if _, err := pool.Exec(t.Context(), "UPDATE trips SET status = 'completed' WHERE id = $1", tripID); err != nil {
		t.Fatal(err)
	}
	found, err = store.TripsNeedingEvaluation(t.Context(), from, until, 50)
	if err != nil {
		t.Fatal(err)
	}
	if containsTrip(found, tripID) {
		t.Fatal("a completed trip was queued for evaluation")
	}
}

func containsTrip(trips []trip.Trip, tripID string) bool {
	for _, record := range trips {
		if record.ID == tripID {
			return true
		}
	}
	return false
}

func ownerOf(t *testing.T, pool *postgres.Pool, tripID string) string {
	t.Helper()
	var owner string
	row := pool.QueryRow(t.Context(), "SELECT owner_id::text FROM trips WHERE id = $1", tripID)
	if err := row.Scan(&owner); err != nil {
		t.Fatal(err)
	}
	return owner
}
