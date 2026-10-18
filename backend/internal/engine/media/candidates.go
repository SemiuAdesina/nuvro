package media

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

type Context struct {
	DestinationID string
	PlaceID       string
	StartDate     time.Time
	EndDate       time.Time
	Landscape     bool
	Language      string
}

type Candidate struct {
	Media             destination.Media
	Source            evidence.Source
	ProviderReference string
	Season            string
}

type CandidateReader interface {
	Candidates(context.Context, Context) ([]Candidate, error)
}
