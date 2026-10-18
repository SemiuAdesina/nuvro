package destination

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

type Media struct {
	URL           string
	Width         int
	Height        int
	AspectRatio   float64
	DominantColor string
	AltText       string
	Attribution   []evidence.Attribution
	ExpiresAt     *time.Time
	Freshness     evidence.Freshness
}
