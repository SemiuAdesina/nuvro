package places

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
)

type NearbyQuery struct {
	Centre       geo.Point
	RadiusMeters int
	Categories   []place.Category
	OpenNow      bool
	Language     string
	Limit        int
}

type TextQuery struct {
	Text         string
	Centre       geo.Point
	RadiusMeters int
	Language     string
	Limit        int
}

type DetailsQuery struct {
	Reference place.Reference
	Language  string
	At        time.Time
}

type SearchProvider interface {
	Nearby(context.Context, NearbyQuery) ([]place.Place, error)
	Text(context.Context, TextQuery) ([]place.Place, error)
}

type DetailsProvider interface {
	Details(context.Context, DetailsQuery) (place.Place, error)
}

type Cache interface {
	Places(ctx context.Context, key string) ([]place.Place, bool)
	StorePlaces(ctx context.Context, key string, values []place.Place, ttl time.Duration)
}
