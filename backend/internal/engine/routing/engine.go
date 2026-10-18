package routing

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const StaticCacheTTL = 24 * time.Hour
const TrafficCacheTTL = 10 * time.Minute
const MaximumMatrixPoints = 12

type Engine struct {
	provider Provider
	cache    Cache
	timer    clock.Clock
}

func New(provider Provider, cache Cache, timer clock.Clock) *Engine {
	return &Engine{provider: provider, cache: cache, timer: timer}
}

func (e *Engine) Configured() bool { return e.provider != nil }

func (e *Engine) Compute(ctx context.Context, query Query) route.Evidence {
	if !query.Mode.Valid() {
		return route.UnavailableEvidence(query.Origin, query.Destination, query.Mode)
	}
	if e.provider == nil {
		return route.UnavailableEvidence(query.Origin, query.Destination, query.Mode)
	}
	key := routeKey(query)
	if cached, found := e.cached(ctx, key); found {
		return cached
	}
	evidence, err := e.provider.Compute(ctx, query)
	if err != nil || evidence.Validate() != nil {
		return route.UnavailableEvidence(query.Origin, query.Destination, query.Mode)
	}
	e.store(ctx, key, evidence, TimeToLive(query.Mode))
	return evidence
}

func (e *Engine) Matrix(ctx context.Context, query MatrixQuery) route.Matrix {
	empty := route.Matrix{Points: query.Points, Mode: query.Mode}
	if e.provider == nil || !query.Mode.Valid() || len(query.Points) < 2 {
		return empty
	}
	if len(query.Points) > MaximumMatrixPoints {
		query.Points = query.Points[:MaximumMatrixPoints]
	}
	matrix, err := e.provider.Matrix(ctx, query)
	if err != nil {
		return empty
	}
	matrix.Points = query.Points
	matrix.Mode = query.Mode
	return matrix
}

func (e *Engine) LeaveBy(arriveBy time.Time, evidence route.Evidence, buffer time.Duration) (time.Time, bool) {
	if !evidence.Known() {
		return time.Time{}, false
	}
	return arriveBy.Add(-evidence.Duration() - buffer), true
}

func TimeToLive(mode route.Mode) time.Duration {
	switch mode {
	case route.Walk, route.Bike:
		return StaticCacheTTL
	default:
		return TrafficCacheTTL
	}
}

func (e *Engine) cached(ctx context.Context, key string) (route.Evidence, bool) {
	if e.cache == nil {
		return route.Evidence{}, false
	}
	return e.cache.Route(ctx, key)
}

func (e *Engine) store(ctx context.Context, key string, evidence route.Evidence, ttl time.Duration) {
	if e.cache == nil {
		return
	}
	e.cache.StoreRoute(ctx, key, evidence, ttl)
}

func routeKey(query Query) string {
	parts := []string{
		string(query.Mode), coordinate(query.Origin), coordinate(query.Destination),
	}
	if query.Mode == route.Drive || query.Mode == route.Transit {
		parts = append(parts, query.DepartAt.UTC().Format("2006010215"))
	}
	return strings.Join(parts, "_")
}

func coordinate(point geo.Point) string {
	return strconv.FormatFloat(point.Latitude, 'f', 4, 64) + "," + strconv.FormatFloat(point.Longitude, 'f', 4, 64)
}
