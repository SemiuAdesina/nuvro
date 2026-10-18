package requirements

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

type TravelerContext struct {
	PassportCountries    []string
	NationalityCountries []string
	ResidenceCountry     string
}

type Query struct {
	Traveler           TravelerContext
	DestinationCountry string
	StartDate          time.Time
	EndDate            time.Time
}

type Evidence struct {
	Query        Query
	Source       evidence.Source
	Requirements []string
}

type EvidenceReader interface {
	Read(context.Context, Query) (Evidence, error)
}
