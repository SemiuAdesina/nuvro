//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

func newIdentifier(t *testing.T) string {
	t.Helper()
	value, err := id.NewUUID().Generate()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func seedTrip(t *testing.T, pool *postgres.Pool, store *postgres.Store) string {
	t.Helper()
	ownerID := newIdentifier(t)
	tripID := newIdentifier(t)
	if err := store.CreateAccount(t.Context(), user.Account{
		UserID: ownerID, IdentityID: newIdentifier(t), Provider: "fixture", Subject: newIdentifier(t),
		Profile: user.DefaultProfile(ownerID), Preferences: user.DefaultPreferences(),
	}); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)
	destinationID := "itinerary-test-destination"
	if err := store.SaveDestination(t.Context(), destination.Destination{
		ID: destinationID, Name: "Itinerary Test Destination", CountryCode: "PT",
		Timezone: "Atlantic/Madeira", Latitude: 32.76, Longitude: -16.96, PrimaryAirport: "FNC",
		CrowdProfile: "any", PaceProfile: "balanced",
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(t.Context(), "DELETE FROM destinations WHERE id = $1", destinationID) })
	now := time.Now().UTC()
	record := trip.Trip{
		ID: tripID, OwnerID: ownerID, DestinationID: &destinationID, StartDate: &start, EndDate: &end,
		Adults: 2, Currency: "EUR", Status: trip.Planning, CreatedAt: now, UpdatedAt: now,
	}
	if err := store.Insert(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	return tripID
}

func itineraryItem(t *testing.T, tripID string, day int, startHour, endHour int, locked bool) itinerary.Item {
	t.Helper()
	base := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC).AddDate(0, 0, day-1)
	return itinerary.Item{
		ID: newIdentifier(t), TripID: tripID, Day: day,
		StartAt: base.Add(time.Duration(startHour) * time.Hour),
		EndAt:   base.Add(time.Duration(endHour) * time.Hour),
		Kind:    itinerary.Activity, Title: "Visit", PlaceID: "fixture:place-1",
		Category: place.Museum, Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048},
		RouteFromPrevious: &route.Evidence{
			Status: route.Found, Mode: route.Walk, DurationSeconds: 600, DistanceMeters: 800, Confidence: route.High,
		},
		WeatherImpact: weather.LowImpact, EstimatedCostMinor: 1500, CostCurrency: "EUR",
		Status: itinerary.Planned, Source: itinerary.Generated, UserLocked: locked,
	}
}

func referencedPlace() place.Place {
	return place.Place{
		ID: "fixture:place-1", References: []place.Reference{{Provider: "fixture", ExternalID: "place-1"}},
		Name: "Museu", Category: place.Museum, Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048},
		Address: "Funchal", Timezone: "Atlantic/Madeira", Status: place.Open,
	}
}

func TestItineraryRoundTripsThroughPostgres(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "Atlantic/Madeira",
		AlgorithmVersion: "itinerary-v1", GeneratedAt: time.Now().UTC().Truncate(time.Second),
		Items: []itinerary.Item{itineraryItem(t, tripID, 1, 9, 11, false), itineraryItem(t, tripID, 2, 10, 12, false)},
	}
	saved, err := store.SaveItinerary(t.Context(), plan, []place.Place{referencedPlace()})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Version != 1 {
		t.Fatalf("version = %d", saved.Version)
	}
	loaded, err := store.Itinerary(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Items) != 2 || loaded.Timezone != "Atlantic/Madeira" {
		t.Fatalf("loaded = %+v", loaded)
	}
	first := loaded.Items[0]
	if first.PlaceID != "fixture:place-1" || first.Category != place.Museum {
		t.Fatalf("place reference lost: %+v", first)
	}
	if first.RouteFromPrevious == nil || first.RouteFromPrevious.DurationSeconds != 600 {
		t.Fatalf("route evidence lost: %+v", first.RouteFromPrevious)
	}
	if first.EstimatedCostMinor != 1500 || first.CostCurrency != "EUR" {
		t.Fatalf("cost lost: %+v", first)
	}
	if first.WeatherImpact != weather.LowImpact {
		t.Fatalf("weather impact lost: %q", first.WeatherImpact)
	}
	if first.Location.Latitude == 0 {
		t.Fatal("coordinates lost")
	}
	if !first.StartAt.Equal(plan.Items[0].StartAt.UTC()) {
		t.Fatalf("times drifted: %v vs %v", first.StartAt, plan.Items[0].StartAt)
	}
}

func TestRegenerationReplacesUnlockedItemsAndKeepsLockedOnes(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	locked := itineraryItem(t, tripID, 1, 19, 21, true)
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "Atlantic/Madeira",
		AlgorithmVersion: "itinerary-v1", GeneratedAt: time.Now().UTC(),
		Items: []itinerary.Item{itineraryItem(t, tripID, 1, 9, 11, false)},
	}
	if _, err := store.SaveItinerary(t.Context(), plan, []place.Place{referencedPlace()}); err != nil {
		t.Fatal(err)
	}
	plan.Items = append(plan.Items, locked)
	if _, err := store.SaveItinerary(t.Context(), plan, nil); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Itinerary(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Items) != 2 {
		t.Fatalf("stored %d items", len(stored.Items))
	}
	replacement := itinerary.Itinerary{
		ID: stored.ID, TripID: tripID, Timezone: "Atlantic/Madeira", AlgorithmVersion: "itinerary-v1",
		GeneratedAt: time.Now().UTC(),
		Items:       []itinerary.Item{itineraryItem(t, tripID, 3, 14, 16, false), locked},
	}
	saved, err := store.SaveItinerary(t.Context(), replacement, nil)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Version != 3 {
		t.Fatalf("version = %d, want 3 after three saves", saved.Version)
	}
	final, err := store.Itinerary(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Items) != 2 {
		t.Fatalf("final items = %d", len(final.Items))
	}
	foundLocked := false
	for _, item := range final.Items {
		if item.ID == locked.ID {
			foundLocked = true
			if !item.UserLocked {
				t.Fatal("the locked flag was lost")
			}
		}
	}
	if !foundLocked {
		t.Fatal("the locked item was deleted on regeneration")
	}
}

func TestOneItineraryPerTripAndItemUpdates(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	item := itineraryItem(t, tripID, 1, 9, 11, false)
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "UTC", AlgorithmVersion: "itinerary-v1",
		GeneratedAt: time.Now().UTC(), Items: []itinerary.Item{item},
	}
	first, err := store.SaveItinerary(t.Context(), plan, []place.Place{referencedPlace()})
	if err != nil {
		t.Fatal(err)
	}
	second := plan
	second.ID = newIdentifier(t)
	again, err := store.SaveItinerary(t.Context(), second, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != first.ID {
		t.Fatal("a second itinerary row was created for one trip")
	}
	stored, err := store.ItineraryItem(t.Context(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	stored.Status = itinerary.Reserved
	stored.UserLocked = true
	stored.Title = "Reserved visit"
	if err := store.UpdateItineraryItem(t.Context(), stored); err != nil {
		t.Fatal(err)
	}
	updated, err := store.ItineraryItem(t.Context(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != itinerary.Reserved || !updated.UserLocked || updated.Title != "Reserved visit" {
		t.Fatalf("update lost: %+v", updated)
	}
}

func TestDatabaseRejectsImpossibleItineraryItems(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	reversed := itineraryItem(t, tripID, 1, 11, 9, false)
	plan := itinerary.Itinerary{
		ID: newIdentifier(t), TripID: tripID, Timezone: "UTC", AlgorithmVersion: "itinerary-v1",
		GeneratedAt: time.Now().UTC(), Items: []itinerary.Item{reversed},
	}
	if _, err := store.SaveItinerary(t.Context(), plan, nil); err == nil {
		t.Fatal("an item ending before it starts was stored")
	}
}
