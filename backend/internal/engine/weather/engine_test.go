package weather_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var now = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

type provider struct {
	current  domain.Observation
	hourly   []domain.Observation
	daily    []domain.Day
	err      error
	requests int
	lastDays int
}

func (p *provider) Current(context.Context, engine.Query) (domain.Observation, error) {
	p.requests++
	return p.current, p.err
}

func (p *provider) Hourly(_ context.Context, query engine.Query) ([]domain.Observation, error) {
	p.requests++
	return p.hourly, p.err
}

func (p *provider) Daily(_ context.Context, query engine.Query) ([]domain.Day, error) {
	p.requests++
	p.lastDays = query.Days
	return p.daily, p.err
}

type cache struct {
	stored map[string]domain.Window
	ttl    time.Duration
	reads  int
}

func (c *cache) Window(_ context.Context, key string) (domain.Window, bool) {
	c.reads++
	window, found := c.stored[key]
	return window, found
}

func (c *cache) StoreWindow(_ context.Context, key string, window domain.Window, ttl time.Duration) {
	if c.stored == nil {
		c.stored = map[string]domain.Window{}
	}
	c.stored[key] = window
	c.ttl = ttl
}

func full() *provider {
	return &provider{
		current: domain.Observation{At: now, Condition: domain.Clear, TemperatureCelsius: 22},
		hourly: []domain.Observation{
			{At: now.Add(time.Hour), Condition: domain.Cloudy, TemperatureCelsius: 21},
			{At: now.Add(400 * time.Hour), Condition: domain.Clear, TemperatureCelsius: 25},
		},
		daily: []domain.Day{
			{Date: now.Truncate(24 * time.Hour), Condition: domain.Clear, HighCelsius: 24, LowCelsius: 18},
			{Date: now.AddDate(0, 0, 40), Condition: domain.Clear, HighCelsius: 30, LowCelsius: 20},
		},
	}
}

func TestNearTermForecastIsReturnedAndFiltered(t *testing.T) {
	source := full()
	window := engine.New(source, nil, clock.NewFixed(now)).
		For(t.Context(), funchal, now.AddDate(0, 0, 1), now.AddDate(0, 0, 4), "en")
	if !window.Known() {
		t.Fatalf("availability = %q", window.Availability)
	}
	if len(window.Hourly) != 1 {
		t.Fatalf("kept %d hours, want only the one inside the horizon", len(window.Hourly))
	}
	if len(window.Daily) != 1 {
		t.Fatalf("kept %d days, want only the one inside the horizon", len(window.Daily))
	}
	if window.Current == nil || window.Source.Provider == "" {
		t.Fatal("current conditions or provenance missing")
	}
}

func TestTripsBeyondTheForecastHorizonGetNoInventedWeather(t *testing.T) {
	source := full()
	window := engine.New(source, nil, clock.NewFixed(now)).
		For(t.Context(), funchal, now.AddDate(0, 6, 0), now.AddDate(0, 6, 7), "en")
	if window.Availability != domain.BeyondForecast {
		t.Fatalf("availability = %q, want beyond_forecast_horizon", window.Availability)
	}
	if window.Current != nil || len(window.Hourly) != 0 || len(window.Daily) != 0 {
		t.Fatal("a long range trip received forecast data")
	}
	if source.requests != 0 {
		t.Fatal("the provider was called for a date it cannot forecast")
	}
}

func TestUnconfiguredOrFailingProviderReportsUnavailable(t *testing.T) {
	unconfigured := engine.New(nil, nil, clock.NewFixed(now))
	if window := unconfigured.For(t.Context(), funchal, now, now.AddDate(0, 0, 2), "en"); window.Availability != domain.ProviderUnavailable {
		t.Fatalf("unconfigured availability = %q", window.Availability)
	}
	if unconfigured.Configured() {
		t.Fatal("an unconfigured engine reported itself configured")
	}
	failing := &provider{err: errors.New("provider down")}
	window := engine.New(failing, nil, clock.NewFixed(now)).For(t.Context(), funchal, now, now.AddDate(0, 0, 2), "en")
	if window.Availability != domain.ProviderUnavailable || window.Known() {
		t.Fatalf("a failing provider produced %+v", window)
	}
}

func TestPartialProviderSuccessStillReturnsWhatIsKnown(t *testing.T) {
	source := full()
	source.current = domain.Observation{}
	window := engine.New(source, nil, clock.NewFixed(now)).For(t.Context(), funchal, now, now.AddDate(0, 0, 2), "en")
	if !window.Known() {
		t.Fatal("a missing current observation discarded the forecast")
	}
	if window.Current != nil {
		t.Fatal("an invalid current observation was kept")
	}
	if len(window.Daily) == 0 {
		t.Fatal("daily forecast lost")
	}
}

func TestWindowsAreCachedWithADocumentedLifetime(t *testing.T) {
	store := &cache{}
	source := full()
	subject := engine.New(source, store, clock.NewFixed(now))
	first := subject.For(t.Context(), funchal, now, now.AddDate(0, 0, 2), "en")
	calls := source.requests
	second := subject.For(t.Context(), funchal, now, now.AddDate(0, 0, 2), "en")
	if source.requests != calls {
		t.Fatal("the provider was called again for an identical window")
	}
	if len(second.Daily) != len(first.Daily) {
		t.Fatal("the cached window differs from the live one")
	}
	if store.ttl != engine.ForecastCacheTTL || store.ttl <= 0 {
		t.Fatalf("forecast cache ttl = %v", store.ttl)
	}
	subject.For(t.Context(), funchal, now, now.AddDate(0, 0, 5), "en")
	if source.requests == calls {
		t.Fatal("a different window reused the cached entry")
	}
}

func TestRequestedCountsRespectProviderHorizons(t *testing.T) {
	source := full()
	engine.New(source, nil, clock.NewFixed(now)).For(t.Context(), funchal, now, now.AddDate(0, 0, 9), "en")
	if source.lastDays < 1 || source.lastDays > domain.DailyHorizonDays {
		t.Fatalf("requested %d days", source.lastDays)
	}
}

func TestImpactUsesHourlyThenDailyThenUnknown(t *testing.T) {
	subject := engine.New(full(), nil, clock.NewFixed(now))
	window := subject.For(t.Context(), funchal, now, now.AddDate(0, 0, 3), "en")
	if impact := subject.Impact(window, now.Add(time.Hour), domain.Outdoor); impact == domain.UnknownImpact {
		t.Fatal("an hour with a forecast produced unknown impact")
	}
	if impact := subject.Impact(window, now.AddDate(0, 0, 8), domain.Outdoor); impact != domain.UnknownImpact {
		t.Fatalf("a day with no forecast produced %q", impact)
	}
	unavailable := domain.Unavailable(domain.ProviderUnavailable)
	if impact := subject.Impact(unavailable, now, domain.Outdoor); impact != domain.UnknownImpact {
		t.Fatal("an unavailable window produced an impact")
	}
}
