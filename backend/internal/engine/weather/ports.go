package weather

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type Query struct {
	Location geo.Point
	Language string
	Days     int
	Hours    int
}

type Provider interface {
	Current(context.Context, Query) (weather.Observation, error)
	Hourly(context.Context, Query) ([]weather.Observation, error)
	Daily(context.Context, Query) ([]weather.Day, error)
}

type Cache interface {
	Window(ctx context.Context, key string) (weather.Window, bool)
	StoreWindow(ctx context.Context, key string, window weather.Window, ttl time.Duration)
}
