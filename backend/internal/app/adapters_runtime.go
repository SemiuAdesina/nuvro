package app

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
)

type routeComputer struct{ engine *routing.Engine }

func (r routeComputer) Compute(ctx context.Context, origin, destination geo.Point, mode route.Mode, departAt time.Time) route.Evidence {
	return r.engine.Compute(ctx, routing.Query{
		Origin: origin, Destination: destination, Mode: mode, DepartAt: departAt,
	})
}
