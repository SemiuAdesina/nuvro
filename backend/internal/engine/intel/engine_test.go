package intel_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var now = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

type placeReader struct {
	places []place.Place
	err    error
	seen   []place.Category
}

func (p *placeReader) Nearby(_ context.Context, _ geo.Point, categories []place.Category, _ string, limit int) ([]place.Place, error) {
	p.seen = categories
	return p.places, p.err
}

type weatherReader struct{ window weather.Window }

func (w weatherReader) For(context.Context, geo.Point, time.Time, time.Time, string) weather.Window {
	return w.window
}

type mediaSelector struct{ assets []destination.Media }

func (m mediaSelector) Select(context.Context, media.Context, int) []destination.Media {
	return m.assets
}

func madeira() destination.Destination {
	return destination.Destination{
		ID: "madeira-pt", Name: "Madeira", CountryCode: "PT", Timezone: "Atlantic/Madeira",
		Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
	}
}

func request() intel.Request {
	return intel.Request{
		Destination: madeira(), Start: now.AddDate(0, 0, 1), End: now.AddDate(0, 0, 5),
		Language: "en", Interests: []string{"food", "nature"},
	}
}

func samplePlace() place.Place {
	return place.Place{
		ID: "a", Name: "Mercado", Category: place.Market, Location: geo.Point{Latitude: 32.6, Longitude: -16.9},
		Status: place.Open, Rating: 4.4, ReviewCount: 100,
	}
}

func sampleMedia() destination.Media {
	return destination.Media{
		URL: "https://example.invalid/photo", Width: 1600, Height: 900, AspectRatio: 1.78,
		Attribution: []evidence.Attribution{{Text: "A Photographer", Provider: "google_places"}},
	}
}

func TestEverySectionAvailable(t *testing.T) {
	engine := intel.New(
		&placeReader{places: []place.Place{samplePlace()}},
		weatherReader{window: weather.Window{Availability: weather.Available, Daily: []weather.Day{{Date: now, Condition: weather.Clear}}}},
		mediaSelector{assets: []destination.Media{sampleMedia()}},
		clock.NewFixed(now),
	)
	result := engine.Describe(t.Context(), request())
	for name, section := range map[string]intel.Section{
		"weather": result.WeatherSection, "places": result.PlacesSection, "media": result.MediaSection,
	} {
		if section.Status != intel.Available {
			t.Fatalf("%s section = %+v", name, section)
		}
		if section.Freshness.ObservedAt.IsZero() || !section.Freshness.ValidUntil.After(section.Freshness.ObservedAt) {
			t.Fatalf("%s section has no usable freshness: %+v", name, section.Freshness)
		}
	}
	if len(result.ThingsToDo) != 1 || len(result.Media) != 1 || !result.Weather.Known() {
		t.Fatalf("content missing: %+v", result)
	}
	if result.Destination.ID != "madeira-pt" || result.GeneratedAt.IsZero() {
		t.Fatal("destination or generation time lost")
	}
}

func TestOneFailingSourceDoesNotFailThePage(t *testing.T) {
	engine := intel.New(
		&placeReader{err: errors.New("places down")},
		weatherReader{window: weather.Window{Availability: weather.Available, Daily: []weather.Day{{Date: now}}}},
		mediaSelector{assets: []destination.Media{sampleMedia()}},
		clock.NewFixed(now),
	)
	result := engine.Describe(t.Context(), request())
	if result.PlacesSection.Status != intel.Unavailable || result.PlacesSection.Reason == "" {
		t.Fatalf("places section = %+v", result.PlacesSection)
	}
	if result.WeatherSection.Status != intel.Available || result.MediaSection.Status != intel.Available {
		t.Fatal("a failing place provider took down other sections")
	}
	if len(result.Media) != 1 {
		t.Fatal("media was lost")
	}
}

func TestBeyondForecastIsNotAFailure(t *testing.T) {
	engine := intel.New(
		&placeReader{places: []place.Place{samplePlace()}},
		weatherReader{window: weather.Unavailable(weather.BeyondForecast)},
		mediaSelector{assets: []destination.Media{sampleMedia()}},
		clock.NewFixed(now),
	)
	result := engine.Describe(t.Context(), request())
	if result.WeatherSection.Status != intel.NotApplicable {
		t.Fatalf("weather section = %+v", result.WeatherSection)
	}
	if result.WeatherSection.Reason != "beyond_forecast_horizon" {
		t.Fatalf("reason = %q", result.WeatherSection.Reason)
	}
	if result.Weather.Known() {
		t.Fatal("weather beyond the horizon was reported as known")
	}
}

func TestNoProvidersConfiguredReportsEverySectionUnavailable(t *testing.T) {
	result := intel.New(nil, nil, nil, clock.NewFixed(now)).Describe(t.Context(), request())
	for name, section := range map[string]intel.Section{
		"weather": result.WeatherSection, "places": result.PlacesSection, "media": result.MediaSection,
	} {
		if section.Status != intel.Unavailable || section.Reason == "" {
			t.Fatalf("%s section = %+v", name, section)
		}
	}
	if len(result.ThingsToDo) != 0 || len(result.Media) != 0 {
		t.Fatal("content was invented without providers")
	}
	if result.Destination.ID == "" {
		t.Fatal("the static destination record should still be returned")
	}
}

func TestEmptyResultsAreReportedAsUnavailableNotAvailable(t *testing.T) {
	engine := intel.New(&placeReader{}, weatherReader{window: weather.Unavailable(weather.ProviderUnavailable)},
		mediaSelector{}, clock.NewFixed(now))
	result := engine.Describe(t.Context(), request())
	if result.PlacesSection.Status != intel.Unavailable || result.MediaSection.Status != intel.Unavailable {
		t.Fatal("empty results were reported as available")
	}
	if result.WeatherSection.Status != intel.Unavailable {
		t.Fatalf("weather section = %+v", result.WeatherSection)
	}
}

func TestInterestsSteerTheCategoriesRequested(t *testing.T) {
	reader := &placeReader{places: []place.Place{samplePlace()}}
	intel.New(reader, weatherReader{}, mediaSelector{}, clock.NewFixed(now)).Describe(t.Context(), request())
	wanted := map[place.Category]bool{}
	for _, category := range reader.seen {
		wanted[category] = true
	}
	if !wanted[place.Restaurant] || !wanted[place.Park] {
		t.Fatalf("interests were not translated into categories: %v", reader.seen)
	}
}
