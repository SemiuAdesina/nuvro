package feasibility

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type EntryQuery struct {
	PassportCountry    string
	ResidenceCountry   string
	ResidencePermits   []string
	OriginCountry      string
	DestinationCountry string
	Start              time.Time
	End                time.Time
}

type EntryEvidence struct {
	EntryMode   feasibility.EntryMode
	Preparation feasibility.Preparation
	TransitRisk feasibility.TransitRisk
	Confidence  feasibility.Confidence
	Reasons     []string
	SourceLabel string
	ObservedAt  time.Time
}

type EntryRequirementProvider interface {
	Entry(context.Context, EntryQuery) (EntryEvidence, error)
}

type FlightQuery struct {
	OriginAirport      string
	DestinationAirport string
	Start              time.Time
	End                time.Time
	Travelers          int
}

type FlightEvidence struct {
	Available       bool
	Stops           int
	DurationMinutes int
	HasPrice        bool
	Price           money.Amount
	Confidence      feasibility.Confidence
	ObservedAt      time.Time
}

type FlightEvidenceProvider interface {
	Flights(context.Context, FlightQuery) (FlightEvidence, error)
}

type CostQuery struct {
	DestinationID string
	Nights        int
	Travelers     int
	Currency      string
	Flight        FlightEvidence
}

type CostEvidence struct {
	Minimum    money.Amount
	Expected   money.Amount
	Maximum    money.Amount
	Confidence feasibility.Confidence
	Components []string
	ObservedAt time.Time
}

type TripCostEstimator interface {
	Estimate(context.Context, CostQuery) (CostEvidence, error)
}

type Rate struct {
	Rate       money.Rate
	Source     string
	ObservedAt time.Time
}

type FXRateProvider interface {
	Rate(ctx context.Context, from, to money.Currency, on time.Time) (Rate, error)
}
