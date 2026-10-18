package app

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
)

type placeReader struct{ engine *places.Engine }

func (p placeReader) Nearby(ctx context.Context, centre geo.Point, categories []place.Category, language string, limit int) ([]place.Place, error) {
	return p.engine.Nearby(ctx, places.NearbyQuery{
		Centre: centre, RadiusMeters: 10000, Categories: categories, Language: language, Limit: limit,
	})
}

type placeFinder struct{ engine *places.Engine }

func (p placeFinder) Nearby(ctx context.Context, request itineraryengine.NearbyRequest) ([]place.Place, error) {
	return p.engine.Nearby(ctx, places.NearbyQuery{
		Centre: request.Centre, RadiusMeters: request.RadiusMeters, Categories: request.Categories,
		Language: request.Language, Limit: request.Limit,
	})
}

func (p placeFinder) Resolve(ctx context.Context, name string, centre geo.Point, language string) (place.Place, bool) {
	return p.engine.Resolve(ctx, name, places.TextQuery{Centre: centre, Language: language})
}

type routeFinder struct{ engine *routing.Engine }

func (r routeFinder) Matrix(ctx context.Context, points []geo.Point, mode route.Mode, departAt time.Time) route.Matrix {
	return r.engine.Matrix(ctx, routing.MatrixQuery{Points: points, Mode: mode, DepartAt: departAt})
}
