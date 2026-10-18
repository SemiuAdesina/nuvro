package itinerary_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

var now = time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
var tripStart = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

type finder struct {
	places []place.Place
	err    error
	seen   engine.NearbyRequest
}

func (f *finder) Nearby(_ context.Context, request engine.NearbyRequest) ([]place.Place, error) {
	f.seen = request
	return f.places, f.err
}

func (f *finder) Resolve(context.Context, string, geo.Point, string) (place.Place, bool) {
	return place.Place{}, false
}

type router struct {
	seconds int
	points  int
}

func (r *router) Matrix(_ context.Context, points []geo.Point, mode route.Mode, _ time.Time) route.Matrix {
	r.points = len(points)
	matrix := route.Matrix{Points: points, Mode: mode}
	for from := range points {
		for to := range points {
			if from == to {
				continue
			}
			matrix.Legs = append(matrix.Legs, route.Leg{FromIndex: from, ToIndex: to, Evidence: route.Evidence{
				Status: route.Found, Mode: mode, DurationSeconds: r.seconds, DistanceMeters: r.seconds * 2,
				Confidence: route.High,
			}})
		}
	}
	return matrix
}

type forecaster struct{ impact weather.Impact }

func (f forecaster) Impact(weather.Window, time.Time, weather.Sensitivity) weather.Impact {
	return f.impact
}

func candidate(id string, category place.Category, minutes int) place.Place {
	return place.Place{
		ID: id, Name: id, Category: category, Location: funchal, Status: place.Open,
		Rating: 4.5, ReviewCount: 500, TypicalMinutes: minutes,
	}
}

func candidates(count int) []place.Place {
	values := make([]place.Place, 0, count)
	for index := range count {
		values = append(values, candidate(string(rune('a'+index)), place.Attraction, 90))
	}
	return values
}

func newEngine(t *testing.T, source *finder, routes engine.RouteFinder, impact weather.Impact) *engine.Engine {
	t.Helper()
	return engine.New(source, routes, forecaster{impact: impact}, id.NewUUID(), clock.NewFixed(now))
}

func request(pace engine.Pace, days int) engine.Request {
	return engine.Request{
		TripID: "trip", Destination: funchal, Timezone: "Atlantic/Madeira",
		Start: tripStart, End: tripStart.AddDate(0, 0, days-1), Pace: pace,
		Interests: []string{"culture", "food"}, Currency: "EUR", Now: now,
		Weather: weather.Window{Availability: weather.Available},
	}
}

func TestGeneratedItineraryIsValid(t *testing.T) {
	source := &finder{places: candidates(9)}
	result, err := newEngine(t, source, &router{seconds: 600}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 3))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Violations) != 0 {
		t.Fatalf("the engine produced an invalid plan: %v", result.Violations)
	}
	if len(result.Itinerary.Items) == 0 {
		t.Fatal("no items scheduled")
	}
	if result.Itinerary.AlgorithmVersion != engine.AlgorithmVersion {
		t.Fatal("algorithm version missing")
	}
	for _, item := range result.Itinerary.Items {
		if item.PlaceID == "" || item.ID == "" || item.Title == "" {
			t.Fatalf("item is not grounded in a resolved place: %+v", item)
		}
		if item.Source != domain.Generated || item.Status != domain.Planned {
			t.Fatalf("item provenance wrong: %+v", item)
		}
	}
}

func TestPaceChangesDensityWithoutBreakingValidity(t *testing.T) {
	counts := map[engine.Pace]int{}
	for _, pace := range []engine.Pace{engine.Relaxed, engine.Balanced, engine.Packed} {
		result, err := newEngine(t, &finder{places: candidates(20)}, &router{seconds: 600}, weather.NoImpact).
			Generate(t.Context(), request(pace, 3))
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Violations) != 0 {
			t.Fatalf("%s pace produced violations: %v", pace, result.Violations)
		}
		counts[pace] = len(result.Itinerary.Items)
	}
	if counts[engine.Relaxed] >= counts[engine.Balanced] || counts[engine.Balanced] >= counts[engine.Packed] {
		t.Fatalf("pace did not change density: %v", counts)
	}
}

func TestTravelTimeIsRespectedBetweenConsecutiveItems(t *testing.T) {
	result, err := newEngine(t, &finder{places: candidates(9)}, &router{seconds: 1800}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Violations) != 0 {
		t.Fatalf("violations: %v", result.Violations)
	}
	for _, day := range result.Itinerary.Days() {
		for index := 1; index < len(day.Items); index++ {
			gap := day.Items[index].StartAt.Sub(day.Items[index-1].EndAt)
			if gap < 30*time.Minute {
				t.Fatalf("only %v between items despite a 30 minute walk", gap)
			}
		}
	}
}

func TestClosedPlacesAreNeverScheduled(t *testing.T) {
	closed := candidate("closed", place.Museum, 90)
	closed.Status = place.Closed
	source := &finder{places: append([]place.Place{closed}, candidates(4)...)}
	result, err := newEngine(t, source, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 2))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Itinerary.Items {
		if item.PlaceID == "closed" {
			t.Fatal("a closed place was scheduled")
		}
	}
}

func TestPlacesClosedAtTheProposedHourAreSkipped(t *testing.T) {
	evening := candidate("evening-only", place.Museum, 60)
	evening.Hours = place.OpeningHours{
		Known:     true,
		Intervals: []place.Interval{{Weekday: tripStart.Weekday(), Opens: 20 * 60, Closes: 23 * 60}},
	}
	source := &finder{places: []place.Place{evening}}
	result, err := newEngine(t, source, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 1))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Itinerary.Items {
		if item.PlaceID == "evening-only" && item.StartAt.Hour() < 19 {
			t.Fatalf("scheduled a place outside its opening hours: %v", item.StartAt)
		}
	}
}

func TestHighWeatherImpactActivitiesAreNotScheduled(t *testing.T) {
	source := &finder{places: []place.Place{candidate("beach", place.Beach, 120)}}
	result, err := newEngine(t, source, &router{seconds: 300}, weather.HighImpact).
		Generate(t.Context(), request(engine.Balanced, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Itinerary.Items) != 0 {
		t.Fatalf("a weather-blocked activity was scheduled: %+v", result.Itinerary.Items)
	}
}

func TestLockedItemsArePreservedExactly(t *testing.T) {
	locked := domain.Item{
		ID: "locked-item", TripID: "trip", Day: 1,
		StartAt: tripStart.Add(11 * time.Hour), EndAt: tripStart.Add(13 * time.Hour),
		Kind: domain.Meal, Title: "Reserved lunch", PlaceID: "reserved-place",
		Status: domain.Reserved, Source: domain.Traveller, ReservationLocked: true,
	}
	input := request(engine.Balanced, 2)
	input.Locked = []domain.Item{locked}
	result, err := newEngine(t, &finder{places: candidates(6)}, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range result.Itinerary.Items {
		if item.ID != locked.ID {
			continue
		}
		found = true
		if !item.StartAt.Equal(locked.StartAt) || !item.EndAt.Equal(locked.EndAt) {
			t.Fatalf("a locked item was moved: %+v", item)
		}
	}
	if !found {
		t.Fatal("the locked item disappeared")
	}
	for _, violation := range result.Violations {
		if violation.Code == domain.LockedItemMoved {
			t.Fatal("the validator reports the locked item was moved")
		}
	}
}

func TestNoPlaceEvidenceFailsHonestly(t *testing.T) {
	if _, err := newEngine(t, &finder{}, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 2)); !errors.Is(err, engine.ErrPlacesUnavailable) {
		t.Fatalf("empty catalog produced %v", err)
	}
	failing := &finder{err: errors.New("provider unavailable")}
	if _, err := newEngine(t, failing, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 2)); err == nil {
		t.Fatal("a failing place provider produced an itinerary")
	}
	unconfigured := engine.New(nil, nil, nil, id.NewUUID(), clock.NewFixed(now))
	if _, err := unconfigured.Generate(t.Context(), request(engine.Balanced, 2)); !errors.Is(err, engine.ErrPlacesUnavailable) {
		t.Fatal("an unconfigured engine invented a plan")
	}
}

func TestMissingRouteEvidenceStillProducesAValidPlan(t *testing.T) {
	result, err := newEngine(t, &finder{places: candidates(6)}, nil, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Itinerary.Items) == 0 {
		t.Fatal("no plan without route evidence")
	}
	if len(result.Violations) != 0 {
		t.Fatalf("unknown routes were treated as failures: %v", result.Violations)
	}
}

func TestCandidatePoolIsCappedBeforeRouting(t *testing.T) {
	source := &finder{places: candidates(20)}
	routes := &router{seconds: 300}
	if _, err := newEngine(t, source, routes, weather.NoImpact).
		Generate(t.Context(), request(engine.Balanced, 3)); err != nil {
		t.Fatal(err)
	}
	if source.seen.Limit > engine.CandidateCap {
		t.Fatalf("requested %d candidates, above the cap", source.seen.Limit)
	}
	if routes.points > engine.CandidateCap {
		t.Fatalf("sent %d points into the route matrix", routes.points)
	}
	if len(source.seen.Categories) == 0 {
		t.Fatal("interests were not translated into place categories")
	}
}

func TestScheduleOrderingIsStable(t *testing.T) {
	subject := newEngine(t, &finder{places: candidates(9)}, &router{seconds: 600}, weather.NoImpact)
	first, err := subject.Generate(t.Context(), request(engine.Balanced, 3))
	if err != nil {
		t.Fatal(err)
	}
	for range 5 {
		again, err := subject.Generate(t.Context(), request(engine.Balanced, 3))
		if err != nil {
			t.Fatal(err)
		}
		if len(again.Itinerary.Items) != len(first.Itinerary.Items) {
			t.Fatal("item count changed between identical runs")
		}
		for index := range again.Itinerary.Items {
			if again.Itinerary.Items[index].PlaceID != first.Itinerary.Items[index].PlaceID {
				t.Fatal("schedule ordering is not stable")
			}
			if !again.Itinerary.Items[index].StartAt.Equal(first.Itinerary.Items[index].StartAt) {
				t.Fatal("scheduled times are not stable")
			}
		}
	}
}

func TestItemsStayInsideDestinationDaytimeHours(t *testing.T) {
	result, err := newEngine(t, &finder{places: candidates(12)}, &router{seconds: 300}, weather.NoImpact).
		Generate(t.Context(), request(engine.Packed, 2))
	if err != nil {
		t.Fatal(err)
	}
	location, err := time.LoadLocation("Atlantic/Madeira")
	if err != nil {
		t.Fatal(err)
	}
	shape := engine.ShapeFor(engine.Packed)
	for _, item := range result.Itinerary.Items {
		local := item.StartAt.In(location)
		minutes := local.Hour()*60 + local.Minute()
		if minutes < shape.DayStartMinutes {
			t.Fatalf("item scheduled at %v, before the day starts", local)
		}
		end := item.EndAt.In(location)
		if end.Hour()*60+end.Minute() > shape.DayEndMinutes {
			t.Fatalf("item ends at %v, after the day ends", end)
		}
	}
}
