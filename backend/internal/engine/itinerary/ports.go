package itinerary

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type PlaceFinder interface {
	Nearby(context.Context, NearbyRequest) ([]place.Place, error)
	Resolve(context.Context, string, geo.Point, string) (place.Place, bool)
}

type NearbyRequest struct {
	Centre       geo.Point
	RadiusMeters int
	Categories   []place.Category
	Language     string
	Limit        int
}

type RouteFinder interface {
	Matrix(context.Context, []geo.Point, route.Mode, time.Time) route.Matrix
}

type WeatherReader interface {
	Impact(weather.Window, time.Time, weather.Sensitivity) weather.Impact
}

type Suggester interface {
	Themes(context.Context, Request) ([]string, error)
}

type Request struct {
	TripID      string
	Destination geo.Point
	Timezone    string
	Start       time.Time
	End         time.Time
	Pace        Pace
	Interests   []string
	Language    string
	Currency    string
	Weather     weather.Window
	Locked      []itinerary.Item
	Now         time.Time
}
