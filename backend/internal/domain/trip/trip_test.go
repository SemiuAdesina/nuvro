package trip_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
)

func TestTripValidation(t *testing.T) {
	valid := trip.Trip{ID: "id", OwnerID: "owner", Adults: 1, Currency: "USD", Status: trip.Draft}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	cases := []func(*trip.Trip){
		func(v *trip.Trip) { v.Adults = 0 }, func(v *trip.Trip) { v.Children = -1 },
		func(v *trip.Trip) { v.Currency = "usd" }, func(v *trip.Trip) { v.Status = "invented" },
		func(v *trip.Trip) { d := time.Now(); v.StartDate = &d },
		func(v *trip.Trip) {
			start := time.Now()
			end := start.Add(-24 * time.Hour)
			v.StartDate = &start
			v.EndDate = &end
		},
		func(v *trip.Trip) { empty := ""; v.DestinationID = &empty },
	}
	for _, mutate := range cases {
		value := valid
		mutate(&value)
		if value.Validate() == nil {
			t.Fatalf("accepted %+v", value)
		}
	}
}

func TestLifecycleTransitions(t *testing.T) {
	states := []trip.Status{trip.Draft, trip.Planning, trip.Upcoming, trip.Active, trip.Completed, trip.Cancelled, trip.Archived}
	allowed := map[[2]trip.Status]bool{
		{trip.Draft, trip.Planning}: true, {trip.Draft, trip.Cancelled}: true, {trip.Draft, trip.Archived}: true,
		{trip.Planning, trip.Upcoming}: true, {trip.Planning, trip.Cancelled}: true, {trip.Planning, trip.Archived}: true,
		{trip.Upcoming, trip.Planning}: true, {trip.Upcoming, trip.Active}: true, {trip.Upcoming, trip.Cancelled}: true,
		{trip.Active, trip.Completed}: true, {trip.Active, trip.Cancelled}: true, {trip.Completed, trip.Archived}: true, {trip.Cancelled, trip.Archived}: true,
	}
	for _, from := range states {
		for _, to := range states {
			result, err := from.Transition(to)
			if (err == nil) != allowed[[2]trip.Status{from, to}] {
				t.Errorf("%s -> %s: %v", from, to, err)
			}
			if err != nil && result != from {
				t.Fatal("failed transition changed state")
			}
		}
	}
}
