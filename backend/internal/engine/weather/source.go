package weather

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

func source(observedAt time.Time) evidence.Source {
	return evidence.Source{
		Provider: "weather",
		Freshness: evidence.Freshness{
			ObservedAt: observedAt.UTC(),
			ValidUntil: observedAt.UTC().Add(ForecastCacheTTL),
			Status:     evidence.Live,
		},
	}
}
