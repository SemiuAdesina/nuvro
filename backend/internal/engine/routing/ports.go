package routing

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
)

type Query struct {
	Origin      geo.Point
	Destination geo.Point
	Mode        route.Mode
	DepartAt    time.Time
	Language    string
}

type MatrixQuery struct {
	Points   []geo.Point
	Mode     route.Mode
	DepartAt time.Time
	Language string
}

type Provider interface {
	Compute(context.Context, Query) (route.Evidence, error)
	Matrix(context.Context, MatrixQuery) (route.Matrix, error)
}

type Cache interface {
	Route(ctx context.Context, key string) (route.Evidence, bool)
	StoreRoute(ctx context.Context, key string, evidence route.Evidence, ttl time.Duration)
}
