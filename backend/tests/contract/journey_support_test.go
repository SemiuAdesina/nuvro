package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	destinationusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/destination"
	journeyusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

var madeira = destination.Destination{
	ID: "madeira-pt", Name: "Madeira", Region: "Madeira", CountryCode: "PT",
	Timezone: "Atlantic/Madeira", Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
}

func (s *store) Destination(_ context.Context, destinationID string) (destination.Destination, error) {
	if destinationID != madeira.ID {
		return destination.Destination{}, errors.New("destination not found")
	}
	return madeira, nil
}

func (s *store) SaveItinerary(_ context.Context, plan itinerary.Itinerary, _ []place.Place) (itinerary.Itinerary, error) {
	plan.Version = s.itinerary.Version + 1
	s.itinerary = plan
	s.hasItinerary = true
	return plan, nil
}

func (s *store) Itinerary(context.Context, string) (itinerary.Itinerary, error) {
	if !s.hasItinerary {
		return itinerary.Itinerary{}, errors.New("no itinerary")
	}
	return s.itinerary, nil
}

func (s *store) ItineraryItem(_ context.Context, itemID string) (itinerary.Item, error) {
	for _, item := range s.itinerary.Items {
		if item.ID == itemID {
			return item, nil
		}
	}
	return itinerary.Item{}, errors.New("item not found")
}

func (s *store) UpdateItineraryItem(_ context.Context, item itinerary.Item) error {
	for index := range s.itinerary.Items {
		if s.itinerary.Items[index].ID == item.ID {
			s.itinerary.Items[index] = item
		}
	}
	return nil
}

func contractPlace(id string, category place.Category, minutes int) place.Place {
	return place.Place{
		ID: id, References: []place.Reference{{Provider: "fixture", ExternalID: id}},
		Name: id, Category: category, Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048},
		Address: "Funchal", Status: place.Open, Rating: 4.5, ReviewCount: 300, TypicalMinutes: minutes,
		Media: []destination.Media{{
			URL: "https://example.invalid/" + id, Width: 1600, Height: 900, AspectRatio: 1.78,
			AltText:     "A view of " + id,
			Attribution: []evidence.Attribution{{Text: "A Photographer", Provider: "fixture"}},
		}},
		Attribution: []evidence.Attribution{{Text: "A Photographer", Provider: "fixture"}},
	}
}

func contractPlaces() []place.Place {
	return []place.Place{
		contractPlace("cable-car", place.Attraction, 90),
		contractPlace("market", place.Market, 75),
		contractPlace("museum", place.Museum, 90),
		contractPlace("garden", place.Park, 120),
		contractPlace("dinner", place.Restaurant, 90),
		contractPlace("cafe", place.Cafe, 45),
	}
}

type contractPlaceReader struct{}

func (contractPlaceReader) Nearby(context.Context, geo.Point, []place.Category, string, int) ([]place.Place, error) {
	return contractPlaces(), nil
}

func (contractPlaceReader) NearbyQuery(context.Context, places.NearbyQuery) ([]place.Place, error) {
	return contractPlaces(), nil
}

type contractFinder struct{}

func (contractFinder) Nearby(context.Context, places.NearbyQuery) ([]place.Place, error) {
	return contractPlaces(), nil
}

type contractPlanFinder struct{}

func (contractPlanFinder) Nearby(context.Context, itineraryengine.NearbyRequest) ([]place.Place, error) {
	return contractPlaces(), nil
}

func (contractPlanFinder) Resolve(context.Context, string, geo.Point, string) (place.Place, bool) {
	return place.Place{}, false
}

type contractRouter struct{}

func (contractRouter) Matrix(_ context.Context, points []geo.Point, mode route.Mode, _ time.Time) route.Matrix {
	matrix := route.Matrix{Points: points, Mode: mode}
	for from := range points {
		for to := range points {
			if from == to {
				continue
			}
			matrix.Legs = append(matrix.Legs, route.Leg{FromIndex: from, ToIndex: to, Evidence: route.Evidence{
				Status: route.Found, Mode: mode, DurationSeconds: 600, DistanceMeters: 900, Confidence: route.High,
			}})
		}
	}
	return matrix
}

type contractWeather struct{}

func (contractWeather) For(_ context.Context, _ geo.Point, start, _ time.Time, _ string) weather.Window {
	return weather.Window{
		Availability: weather.Available,
		Current: &weather.Observation{
			At: contractNow, Condition: weather.Clear, TemperatureCelsius: 23,
			PrecipitationChance: 10, HumidityPercent: 60, UVIndex: 5,
		},
		Daily: []weather.Day{{
			Date: start.UTC().Truncate(24 * time.Hour), Condition: weather.Clear,
			HighCelsius: 24, LowCelsius: 18, PrecipitationChance: 10,
		}},
		Source: evidence.Source{Provider: "fixture"},
	}
}

func (contractWeather) Impact(weather.Window, time.Time, weather.Sensitivity) weather.Impact {
	return weather.NoImpact
}

type contractMedia struct{}

func (contractMedia) Select(context.Context, media.Context, int) []destination.Media {
	return contractPlaces()[0].Media
}

type journeyCosts struct{}

func (journeyCosts) Estimates(context.Context, budgetengine.Query) ([]budgetengine.Estimate, error) {
	build := func(minor int64) money.Amount {
		amount, _ := money.New(minor, money.Currency("EUR"))
		return amount
	}
	return []budgetengine.Estimate{
		{Category: budgetdomain.Stay, Minimum: build(90_000), Expected: build(100_000), Maximum: build(120_000),
			Confidence: budgetdomain.Medium, Source: "fixture", ObservedAt: contractNow},
		{Category: budgetdomain.Food, Minimum: build(20_000), Expected: build(28_000), Maximum: build(35_000),
			Confidence: budgetdomain.Low, Source: "fixture", ObservedAt: contractNow},
	}, nil
}

func destinationService(t *testing.T, repository *store, timer clock.Clock) *destinationusecase.Service {
	t.Helper()
	return destinationusecase.NewService(repository, repository,
		intel.New(contractPlaceReader{}, contractWeather{}, contractMedia{}, timer), timer)
}

func journeyService(t *testing.T, repository *store, timer clock.Clock) *journeyusecase.Service {
	t.Helper()
	return journeyusecase.NewService(journeyusecase.Dependencies{
		Trips:   tripAccess{store: repository},
		Catalog: repository, Store: repository, Preferences: repository,
		Planner:    itineraryengine.New(contractPlanFinder{}, contractRouter{}, contractWeather{}, id.NewUUID(), timer),
		Forecaster: contractWeather{}, Accountant: budgetengine.New(journeyCosts{}, nil, nil, timer),
		Finder: contractFinder{}, Identifiers: id.NewUUID(),
	}, timer)
}

type tripAccess struct{ store *store }

func (t tripAccess) Read(ctx context.Context, principal user.Principal, tripID string) (trip.Trip, error) {
	return plan.NewReader(t.store).Read(ctx, principal, tripID)
}

func (t tripAccess) Role(ctx context.Context, tripID, userID string) (trip.Role, error) {
	return t.store.Role(ctx, tripID, userID)
}
