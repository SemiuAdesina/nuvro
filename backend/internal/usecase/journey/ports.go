package journey

import (
	"context"
	"time"

	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
	Role(context.Context, string, string) (trip.Role, error)
}

type Catalog interface {
	Destination(context.Context, string) (destination.Destination, error)
}

type Store interface {
	SaveItinerary(context.Context, itinerary.Itinerary, []place.Place) (itinerary.Itinerary, error)
	Itinerary(context.Context, string) (itinerary.Itinerary, error)
	ItineraryItem(context.Context, string) (itinerary.Item, error)
	UpdateItineraryItem(context.Context, itinerary.Item) error
}

type Preferences interface {
	Preferences(context.Context, string) (user.Preferences, error)
}

type Planner interface {
	Generate(context.Context, itineraryengine.Request) (itineraryengine.Result, error)
}

type Forecaster interface {
	For(ctx context.Context, location geo.Point, start, end time.Time, language string) weather.Window
}

type Accountant interface {
	Calculate(context.Context, budgetengine.Input) (budgetdomain.Trip, error)
}

type Finder interface {
	Nearby(context.Context, places.NearbyQuery) ([]place.Place, error)
}
