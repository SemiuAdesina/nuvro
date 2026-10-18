package replanning

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type PlaceFinder interface {
	Nearby(ctx context.Context, centre geo.Point, categories []place.Category, language string, limit int) ([]place.Place, error)
}

type RouteFinder interface {
	Compute(ctx context.Context, origin, destination geo.Point, mode route.Mode, departAt time.Time) route.Evidence
}

type WeatherReader interface {
	Impact(weather.Window, time.Time, weather.Sensitivity) weather.Impact
}

type Interpreter interface {
	Interpret(ctx context.Context, message string) (Intent, error)
}

type Intent struct {
	Trigger    replan.Trigger
	LowerPace  bool
	PreferNear bool
	Categories []place.Category
}

type Request struct {
	TripID            string
	Trigger           replan.Trigger
	Itinerary         itinerary.Itinerary
	Weather           weather.Window
	Now               time.Time
	CurrentLocation   *geo.Point
	WindowStart       time.Time
	WindowEnd         time.Time
	Message           string
	Language          string
	Currency          string
	BudgetCeiling     *money.Amount
	ReducedWalking    bool
	MaximumWalkMeters int
	Pace              string
	Timezone          string
}
