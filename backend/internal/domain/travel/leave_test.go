package travel_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
)

var day = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)

func at(hour, minute int) time.Time {
	return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

func activity(startHour int) itinerary.Item {
	return itinerary.Item{
		ID: "next", TripID: "trip", Day: 3, StartAt: at(startHour, 0),
		EndAt: at(startHour+2, 0), Kind: itinerary.Activity, Title: "Cable car",
		Category: place.Attraction, Location: geo.Point{Latitude: 32.66, Longitude: -16.9},
		Status: itinerary.Planned, Source: itinerary.Generated,
	}
}

func walking(minutes int) route.Evidence {
	return route.Evidence{
		Status: route.Found, Mode: route.Walk, DurationSeconds: minutes * 60,
		DistanceMeters: minutes * 80, Confidence: route.High,
	}
}

func TestLeaveBySubtractsRouteAndArrivalBuffer(t *testing.T) {
	item := activity(10)
	departure := travel.EvaluateDeparture(at(9, 20), &item, walking(12), travel.DefaultBufferPolicy())
	if !departure.HasLeaveBy {
		t.Fatal("no leave time produced")
	}
	if !departure.LeaveBy.Equal(at(9, 38)) {
		t.Fatalf("leaveBy = %v, want 09:38 after a 12 minute walk and a 10 minute arrival buffer", departure.LeaveBy)
	}
	if departure.MinutesUntilLeave != 18 {
		t.Fatalf("minutesUntilLeave = %d", departure.MinutesUntilLeave)
	}
	if departure.Status != travel.LeaveSoon {
		t.Fatalf("status = %q", departure.Status)
	}
}

func TestDepartureStatusesFollowTheClock(t *testing.T) {
	item := activity(10)
	policy := travel.DefaultBufferPolicy()
	for name, test := range map[string]struct {
		now  time.Time
		want travel.LeaveStatus
	}{
		"well ahead":   {at(8, 0), travel.OnTime},
		"leave soon":   {at(9, 20), travel.LeaveSoon},
		"leave now":    {at(9, 36), travel.LeaveNow},
		"already late": {at(9, 50), travel.LikelyLate},
	} {
		departure := travel.EvaluateDeparture(test.now, &item, walking(12), policy)
		if departure.Status != test.want {
			t.Fatalf("%s: status = %q, want %q", name, departure.Status, test.want)
		}
	}
}

func TestNoRouteEvidenceNeverProducesALeaveTime(t *testing.T) {
	item := activity(10)
	unavailable := route.UnavailableEvidence(geo.Point{}, item.Location, route.Walk)
	departure := travel.EvaluateDeparture(at(9, 20), &item, unavailable, travel.DefaultBufferPolicy())
	if departure.Status != travel.RouteUnavailable {
		t.Fatalf("status = %q", departure.Status)
	}
	if departure.HasLeaveBy || departure.MinutesUntilLeave != -1 {
		t.Fatalf("a leave time was invented without route evidence: %+v", departure)
	}
}

func TestNoUpcomingItemIsReportedHonestly(t *testing.T) {
	departure := travel.EvaluateDeparture(at(9, 20), nil, walking(10), travel.DefaultBufferPolicy())
	if departure.Status != travel.NoUpcomingItem || departure.HasLeaveBy {
		t.Fatalf("departure = %+v", departure)
	}
}

func TestArrivalBufferVariesByItemKind(t *testing.T) {
	policy := travel.DefaultBufferPolicy()
	meal := activity(10)
	meal.Kind = itinerary.Meal
	reserved := activity(10)
	reserved.ReservationLocked = true
	museum := activity(10)
	museum.Category = place.Museum
	beach := activity(10)
	beach.Category = place.Beach
	if policy.For(reserved) <= policy.For(meal) {
		t.Fatal("a reservation should carry the largest arrival buffer")
	}
	if policy.For(meal) <= policy.For(beach) {
		t.Fatal("a meal should carry a larger buffer than a beach visit")
	}
	if policy.For(museum) <= policy.For(beach) {
		t.Fatal("a ticketed venue should carry a larger buffer than an open space")
	}
}

func TestLeaveTimeIsDeterministic(t *testing.T) {
	item := activity(10)
	first := travel.EvaluateDeparture(at(9, 20), &item, walking(12), travel.DefaultBufferPolicy())
	for range 50 {
		again := travel.EvaluateDeparture(at(9, 20), &item, walking(12), travel.DefaultBufferPolicy())
		if again != first {
			t.Fatal("leave time evaluation is not deterministic")
		}
	}
}
