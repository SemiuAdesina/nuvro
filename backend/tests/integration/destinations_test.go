//go:build integration

package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
)

func applySeed(t *testing.T, pool *postgres.Pool) {
	t.Helper()
	content, err := os.ReadFile("../../db/seed/destinations.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range strings.Split(string(content), "-- +seed statement") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := pool.Exec(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(t.Context(), "DELETE FROM destinations WHERE id LIKE '%-__'")
	})
}

func TestDestinationCatalogPersistsStaticMetadataOnly(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	applySeed(t, pool)

	catalog, err := store.Destinations(t.Context(), 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) < 14 {
		t.Fatalf("seed produced %d destinations", len(catalog))
	}
	countries := map[string]bool{}
	for _, value := range catalog {
		if err := value.Validate(); err != nil {
			t.Fatalf("%s is invalid: %v", value.ID, err)
		}
		countries[value.CountryCode] = true
	}
	for _, required := range []string{"TZ", "PT", "TH", "MX", "BB", "NZ"} {
		if !countries[required] {
			t.Fatalf("the seed is not geographically diverse; %s is missing", required)
		}
	}
	zanzibar := find(t, catalog, "zanzibar-tz")
	if len(zanzibar.InterestTags) < 3 || len(zanzibar.ClimateTags) == 0 {
		t.Fatalf("tags were not loaded: %+v", zanzibar)
	}
	if zanzibar.PrimaryAirport != "ZNZ" || zanzibar.Timezone == "" || zanzibar.Latitude == 0 {
		t.Fatalf("static metadata missing: %+v", zanzibar)
	}
}

func TestDestinationUpsertReplacesTagsWithoutDuplicating(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	record := destination.Destination{
		ID: "integration-test-place", Name: "Test Place", Region: "Region", CountryCode: "PT",
		Timezone: "Europe/Lisbon", Latitude: 38.7, Longitude: -9.1, PrimaryAirport: "LIS",
		ClimateTags: []string{"mild"}, InterestTags: []string{"food", "culture"}, StyleTags: []string{"city"},
		CrowdProfile: "quiet", PaceProfile: "relaxed",
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(t.Context(), "DELETE FROM destinations WHERE id = $1", record.ID)
	})
	if err := store.SaveDestination(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	record.InterestTags = []string{"food"}
	record.Name = "Renamed Place"
	if err := store.SaveDestination(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	catalog, err := store.Destinations(t.Context(), 200)
	if err != nil {
		t.Fatal(err)
	}
	stored := find(t, catalog, record.ID)
	if stored.Name != "Renamed Place" {
		t.Fatalf("upsert did not update: %+v", stored)
	}
	if len(stored.InterestTags) != 1 || stored.InterestTags[0] != "food" {
		t.Fatalf("tags were not replaced: %v", stored.InterestTags)
	}
	invalid := record
	invalid.CountryCode = "PRT"
	if err := store.SaveDestination(t.Context(), invalid); err == nil {
		t.Fatal("an invalid record was written")
	}
}

func find(t *testing.T, catalog []destination.Destination, id string) destination.Destination {
	t.Helper()
	for _, value := range catalog {
		if value.ID == id {
			return value
		}
	}
	t.Fatalf("%s not found in the catalog", id)
	return destination.Destination{}
}
