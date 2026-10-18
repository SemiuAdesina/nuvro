package journey_test

import (
	"context"
	"errors"
	"testing"
	"time"

	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
)

const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"

var now = time.Date(2026, 10, 14, 16, 40, 0, 0, time.UTC)
var tripStart = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

type trips struct {
	record trip.Trip
	role   trip.Role
	err    error
}

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if t.err != nil {
		return trip.Trip{}, t.err
	}
	if principal.UserID == "" {
		return trip.Trip{}, errors.New("unauthenticated")
	}
	return t.record, nil
}

func (t trips) Role(context.Context, string, string) (trip.Role, error) { return t.role, t.err }

type catalog struct{ err error }

func (c catalog) Destination(context.Context, string) (destination.Destination, error) {
	if c.err != nil {
		return destination.Destination{}, c.err
	}
	return destination.Destination{
		ID: "madeira-pt", Name: "Madeira", CountryCode: "PT", Timezone: "Atlantic/Madeira",
		Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
	}, nil
}

type store struct {
	plan    itinerary.Itinerary
	present bool
	saved   itinerary.Itinerary
	updated itinerary.Item
	saveErr error
}

func (s *store) SaveItinerary(_ context.Context, plan itinerary.Itinerary, _ []place.Place) (itinerary.Itinerary, error) {
	if s.saveErr != nil {
		return itinerary.Itinerary{}, s.saveErr
	}
	plan.Version = s.plan.Version + 1
	s.saved = plan
	s.plan = plan
	s.present = true
	return plan, nil
}

func (s *store) Itinerary(context.Context, string) (itinerary.Itinerary, error) {
	if !s.present {
		return itinerary.Itinerary{}, errors.New("no itinerary")
	}
	return s.plan, nil
}

func (s *store) ItineraryItem(_ context.Context, itemID string) (itinerary.Item, error) {
	for _, item := range s.plan.Items {
		if item.ID == itemID {
			return item, nil
		}
	}
	return itinerary.Item{}, errors.New("not found")
}

func (s *store) UpdateItineraryItem(_ context.Context, item itinerary.Item) error {
	s.updated = item
	for index := range s.plan.Items {
		if s.plan.Items[index].ID == item.ID {
			s.plan.Items[index] = item
		}
	}
	return nil
}

type preferences struct{ value user.Preferences }

func (p preferences) Preferences(context.Context, string) (user.Preferences, error) {
	return p.value, nil
}

type finder struct {
	places []place.Place
	err    error
}

func (f finder) Nearby(_ context.Context, query places.NearbyQuery) ([]place.Place, error) {
	return f.places, f.err
}

type forecaster struct{ window weather.Window }

func (f forecaster) For(context.Context, geo.Point, time.Time, time.Time, string) weather.Window {
	return f.window
}

type itineraryFinder struct{ places []place.Place }

func (i itineraryFinder) Nearby(context.Context, itineraryengine.NearbyRequest) ([]place.Place, error) {
	return i.places, nil
}

func (i itineraryFinder) Resolve(context.Context, string, geo.Point, string) (place.Place, bool) {
	return place.Place{}, false
}

type router struct{ seconds int }

func (r router) Matrix(_ context.Context, points []geo.Point, mode route.Mode, _ time.Time) route.Matrix {
	matrix := route.Matrix{Points: points, Mode: mode}
	for from := range points {
		for to := range points {
			if from == to {
				continue
			}
			matrix.Legs = append(matrix.Legs, route.Leg{FromIndex: from, ToIndex: to, Evidence: route.Evidence{
				Status: route.Found, Mode: mode, DurationSeconds: r.seconds, DistanceMeters: r.seconds, Confidence: route.High,
			}})
		}
	}
	return matrix
}

type impacts struct{ value weather.Impact }

func (i impacts) Impact(weather.Window, time.Time, weather.Sensitivity) weather.Impact {
	return i.value
}

type costs struct {
	values []budgetengine.Estimate
	err    error
}

func (c costs) Estimates(context.Context, budgetengine.Query) ([]budgetengine.Estimate, error) {
	return c.values, c.err
}

func madeiraPlace(id string, category place.Category, minutes int) place.Place {
	return place.Place{
		ID: id, Name: id, Category: category, Location: funchal, Status: place.Open,
		Rating: 4.5, ReviewCount: 400, TypicalMinutes: minutes,
	}
}

func madeiraPlaces() []place.Place {
	return []place.Place{
		madeiraPlace("cable-car", place.Attraction, 90),
		madeiraPlace("market", place.Market, 75),
		madeiraPlace("botanical-garden", place.Park, 120),
		madeiraPlace("old-town-dinner", place.Restaurant, 90),
		madeiraPlace("art-museum", place.Museum, 90),
		madeiraPlace("praia-formosa", place.Beach, 150),
		madeiraPlace("miradouro", place.Viewpoint, 45),
		madeiraPlace("cafe-central", place.Cafe, 45),
		madeiraPlace("levada-walk", place.Experience, 180),
	}
}

func tripRecord(t *testing.T, days int) trip.Trip {
	t.Helper()
	destinationID := "madeira-pt"
	start := tripStart
	end := tripStart.AddDate(0, 0, days-1)
	return trip.Trip{
		ID: tripID, OwnerID: owner, DestinationID: &destinationID, StartDate: &start, EndDate: &end,
		Adults: 2, Currency: "EUR", Status: trip.Planning,
	}
}

type harness struct {
	service *usecase.Service
	store   *store
}

func newHarness(t *testing.T, dependencies usecase.Dependencies) harness {
	t.Helper()
	if dependencies.Identifiers == nil {
		dependencies.Identifiers = id.NewUUID()
	}
	repository, _ := dependencies.Store.(*store)
	return harness{service: usecase.NewService(dependencies, clock.NewFixed(now)), store: repository}
}

func madeiraDependencies(t *testing.T, repository *store) usecase.Dependencies {
	t.Helper()
	return usecase.Dependencies{
		Trips:   trips{record: tripRecord(t, 7), role: trip.Owner},
		Catalog: catalog{},
		Store:   repository,
		Preferences: preferences{value: user.Preferences{
			Interests: []string{"beach", "food", "culture"}, Pace: "balanced",
		}},
		Planner: itineraryengine.New(itineraryFinder{places: madeiraPlaces()}, router{seconds: 600},
			impacts{value: weather.NoImpact}, id.NewUUID(), clock.NewFixed(now)),
		Forecaster: forecaster{window: weather.Window{Availability: weather.Available}},
		Finder:     finder{places: madeiraPlaces()},
	}
}

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func principal() user.Principal { return user.Principal{UserID: owner} }

var _ = budgetdomain.Flight
