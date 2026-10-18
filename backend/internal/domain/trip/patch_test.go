package trip_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
)

func base() trip.Trip {
	destination := "lisbon"
	origin := "LHR"
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	return trip.Trip{
		ID: "11112233-4455-6677-8899-aabbccddeeff", OwnerID: "00112233-4455-6677-8899-aabbccddeeff",
		DestinationID: &destination, OriginAirport: &origin, StartDate: &start, EndDate: &end,
		Adults: 2, Children: 1, Currency: "EUR", Status: trip.Draft,
		CreatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestOmittedFieldsAreUnchanged(t *testing.T) {
	now := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	unchanged, err := base().Apply(trip.Patch{}, now)
	if err != nil || *unchanged.DestinationID != "lisbon" || *unchanged.OriginAirport != "LHR" || unchanged.StartDate == nil {
		t.Fatalf("omitted fields changed: %+v %v", unchanged, err)
	}
	if unchanged.Adults != 2 || unchanged.UpdatedAt != now {
		t.Fatalf("counts or timestamps drifted: %+v", unchanged)
	}
}

func TestExplicitNullClearsAndValuesAssign(t *testing.T) {
	now := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	cleared, err := base().Apply(trip.Patch{
		DestinationID: change.Clear[string](), OriginAirport: change.Clear[string](), Dates: change.Clear[trip.Dates](),
	}, now)
	if err != nil || cleared.DestinationID != nil || cleared.OriginAirport != nil || cleared.StartDate != nil || cleared.EndDate != nil {
		t.Fatalf("explicit null did not clear: %+v %v", cleared, err)
	}
	assigned, err := base().Apply(trip.Patch{
		DestinationID: change.Set("azores"), Adults: change.Set(int32(4)), Currency: change.Set("USD"),
		Dates: change.Set(trip.Dates{Start: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)}),
	}, now)
	if err != nil || *assigned.DestinationID != "azores" || assigned.Adults != 4 || assigned.Currency != "USD" || assigned.StartDate.Day() != 2 {
		t.Fatalf("assignment failed: %+v %v", assigned, err)
	}
}

func TestPatchRejectsInvalidResultsAndUnclearableFields(t *testing.T) {
	now := time.Now().UTC()
	for name, patch := range map[string]trip.Patch{
		"cleared currency":  {Currency: change.Clear[string]()},
		"cleared adults":    {Adults: change.Clear[int32]()},
		"zero adults":       {Adults: change.Set(int32(0))},
		"lowercase":         {Currency: change.Set("usd")},
		"bad airport":       {OriginAirport: change.Set("Heathrow")},
		"reversed calendar": {Dates: change.Set(trip.Dates{Start: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)})},
	} {
		if _, err := base().Apply(patch, now); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}
