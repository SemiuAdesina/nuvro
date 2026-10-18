package places_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var now = time.Date(2026, 10, 14, 16, 40, 0, 0, time.UTC)
var centre = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

type provider struct {
	nearby []place.Place
	text   []place.Place
	err    error
	calls  int
}

func (p *provider) Nearby(context.Context, places.NearbyQuery) ([]place.Place, error) {
	p.calls++
	return p.nearby, p.err
}

func (p *provider) Text(context.Context, places.TextQuery) ([]place.Place, error) {
	p.calls++
	return p.text, p.err
}

type details struct {
	value place.Place
	err   error
	calls int
}

func (d *details) Details(context.Context, places.DetailsQuery) (place.Place, error) {
	d.calls++
	return d.value, d.err
}

type store struct {
	entries map[string][]place.Place
	ttl     time.Duration
}

func (s *store) Places(_ context.Context, key string) ([]place.Place, bool) {
	values, found := s.entries[key]
	return values, found
}

func (s *store) StorePlaces(_ context.Context, key string, values []place.Place, ttl time.Duration) {
	if s.entries == nil {
		s.entries = map[string][]place.Place{}
	}
	s.entries[key] = values
	s.ttl = ttl
}

func sample(id string, category place.Category, minutes int, status place.OpeningStatus) place.Place {
	return place.Place{
		ID: id, Name: id, Category: category, Location: centre, Status: status,
		Rating: 4.4, ReviewCount: 300, TypicalMinutes: minutes,
	}
}

func TestInvalidAndDuplicatePlacesAreDropped(t *testing.T) {
	source := &provider{nearby: []place.Place{
		sample("good", place.Museum, 60, place.Open),
		sample("good", place.Museum, 60, place.Open),
		{ID: "broken", Name: "", Category: place.Museum},
		{ID: "off-globe", Name: "Off", Category: place.Museum, Location: geo.Point{Latitude: 200}},
	}}
	found, err := places.New(source, nil, nil, clock.NewFixed(now)).
		Nearby(t.Context(), places.NearbyQuery{Centre: centre, RadiusMeters: 2000, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].ID != "good" {
		t.Fatalf("accepted %v", found)
	}
}

func TestNearbyResultsAreCachedPerQuery(t *testing.T) {
	source := &provider{nearby: []place.Place{sample("a", place.Museum, 60, place.Open)}}
	cache := &store{}
	engine := places.New(source, nil, cache, clock.NewFixed(now))
	query := places.NearbyQuery{Centre: centre, RadiusMeters: 2000, Limit: 10}
	if _, err := engine.Nearby(t.Context(), query); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Nearby(t.Context(), query); err != nil {
		t.Fatal(err)
	}
	if source.calls != 1 {
		t.Fatalf("provider called %d times for an identical query", source.calls)
	}
	if cache.ttl != places.NearbyCacheTTL {
		t.Fatalf("nearby ttl = %v", cache.ttl)
	}
	different := query
	different.RadiusMeters = 8000
	if _, err := engine.Nearby(t.Context(), different); err != nil {
		t.Fatal(err)
	}
	if source.calls != 2 {
		t.Fatal("a different radius reused the cached result")
	}
}

func TestDetailsUseALongerLifetimeThanSearch(t *testing.T) {
	value := sample("a", place.Museum, 60, place.Open)
	source := &details{value: value}
	cache := &store{}
	engine := places.New(&provider{}, source, cache, clock.NewFixed(now))
	query := places.DetailsQuery{Reference: place.Reference{Provider: "google_places", ExternalID: "X"}}
	if _, err := engine.Details(t.Context(), query); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Details(t.Context(), query); err != nil {
		t.Fatal(err)
	}
	if source.calls != 1 {
		t.Fatalf("details provider called %d times", source.calls)
	}
	if cache.ttl <= places.NearbyCacheTTL {
		t.Fatalf("details ttl %v is not longer than search ttl %v", cache.ttl, places.NearbyCacheTTL)
	}
}

func TestUnconfiguredProviderIsHonest(t *testing.T) {
	engine := places.New(nil, nil, nil, clock.NewFixed(now))
	if engine.Configured() {
		t.Fatal("an unconfigured engine claimed to be configured")
	}
	if _, err := engine.Nearby(t.Context(), places.NearbyQuery{Centre: centre}); !errors.Is(err, places.ErrUnavailable) {
		t.Fatalf("nearby error = %v", err)
	}
	if _, err := engine.Details(t.Context(), places.DetailsQuery{}); !errors.Is(err, places.ErrUnavailable) {
		t.Fatalf("details error = %v", err)
	}
}

func TestResolveOnlyAcceptsAConfidentNameMatch(t *testing.T) {
	source := &provider{text: []place.Place{sample("Mercado dos Lavradores", place.Market, 60, place.Open)}}
	engine := places.New(source, nil, nil, clock.NewFixed(now))
	if _, found := engine.Resolve(t.Context(), "  mercado DOS lavradores ", places.TextQuery{Centre: centre}); !found {
		t.Fatal("an exact name did not resolve")
	}
	if _, found := engine.Resolve(t.Context(), "Sunset Terrace Cafe", places.TextQuery{Centre: centre}); found {
		t.Fatal("an invented establishment was resolved to an unrelated place")
	}
	empty := places.New(&provider{}, nil, nil, clock.NewFixed(now))
	if _, found := empty.Resolve(t.Context(), "Anything", places.TextQuery{Centre: centre}); found {
		t.Fatal("a name resolved against no results")
	}
}

func TestRankingPutsFittingOpenRelevantPlacesFirst(t *testing.T) {
	candidates := []place.Place{
		sample("long-museum", place.Museum, 240, place.Open),
		sample("quick-cafe", place.Cafe, 45, place.Open),
		sample("closed-market", place.Market, 60, place.Closed),
	}
	ranked := places.Rank(candidates, places.Context{
		Now: now, Origin: centre, AvailableMinutes: 140, Interests: []string{"food"},
		Routes: map[string]route.Evidence{
			"quick-cafe":    {Status: route.Found, Mode: route.Walk, DurationSeconds: 480},
			"long-museum":   {Status: route.Found, Mode: route.Walk, DurationSeconds: 600},
			"closed-market": {Status: route.Found, Mode: route.Walk, DurationSeconds: 300},
		},
	})
	if ranked[0].Place.ID != "quick-cafe" {
		t.Fatalf("ranking = %v", identifiers(ranked))
	}
	if !ranked[0].Fits {
		t.Fatal("a place that fits the free time was marked as not fitting")
	}
	for _, value := range ranked {
		if value.Place.ID == "long-museum" && value.Fits {
			t.Fatal("a four hour visit was said to fit two hours of free time")
		}
	}
	if ranked[len(ranked)-1].Fits {
		t.Fatal("places that do not fit should sort last")
	}
}

func TestRankingDoesNotFilterWhenNoFreeTimeIsKnown(t *testing.T) {
	ranked := places.Rank([]place.Place{sample("long", place.Museum, 300, place.Open)}, places.Context{Now: now})
	if !ranked[0].Fits {
		t.Fatal("with no known free time nothing should be excluded for time")
	}
}

func TestWeatherSuppressesUnsuitableOutdoorPlaces(t *testing.T) {
	stormy := weather.Window{
		Availability: weather.Available,
		Hourly:       []weather.Observation{{At: now, Condition: weather.Storm, TemperatureCelsius: 18}},
	}
	ranked := places.Rank([]place.Place{
		sample("beach", place.Beach, 60, place.Open),
		sample("museum", place.Museum, 60, place.Open),
	}, places.Context{Now: now, Weather: stormy, AvailableMinutes: 240})
	if ranked[0].Place.ID != "museum" {
		t.Fatalf("a storm did not push the beach down: %v", identifiers(ranked))
	}
	for _, value := range ranked {
		if value.Place.ID == "beach" && value.WeatherImpact != weather.HighImpact {
			t.Fatalf("beach weather impact = %q", value.WeatherImpact)
		}
	}
}

func TestRankingIsDeterministic(t *testing.T) {
	candidates := []place.Place{
		sample("a", place.Museum, 60, place.Open),
		sample("b", place.Museum, 60, place.Open),
		sample("c", place.Museum, 60, place.Open),
	}
	first := identifiers(places.Rank(candidates, places.Context{Now: now}))
	for range 20 {
		again := identifiers(places.Rank(candidates, places.Context{Now: now}))
		for index := range again {
			if again[index] != first[index] {
				t.Fatalf("ranking changed: %v then %v", first, again)
			}
		}
	}
}

func identifiers(ranked []places.Ranked) []string {
	values := make([]string, 0, len(ranked))
	for _, value := range ranked {
		values = append(values, value.Place.ID)
	}
	return values
}
