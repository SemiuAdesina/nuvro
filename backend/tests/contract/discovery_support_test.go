package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/matching"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

func (s *store) Destinations(context.Context, int32) ([]destination.Destination, error) {
	return []destination.Destination{
		{
			ID: "zanzibar-tz", Name: "Zanzibar", Region: "Unguja", CountryCode: "TZ", Timezone: "Africa/Dar_es_Salaam",
			Latitude: -6.1659, Longitude: 39.2026, PrimaryAirport: "ZNZ",
			ClimateTags: []string{"warm"}, InterestTags: []string{"beach", "food", "relaxation"},
			CrowdProfile: "quiet", PaceProfile: "relaxed",
		},
		{
			ID: "reykjavik-is", Name: "Reykjavik", CountryCode: "IS", Timezone: "Atlantic/Reykjavik",
			Latitude: 64.1466, Longitude: -21.9426, PrimaryAirport: "KEF",
			ClimateTags: []string{"cold"}, InterestTags: []string{"nature"},
			CrowdProfile: "quiet", PaceProfile: "balanced",
		},
	}, nil
}

type contractEntry struct{}

func (contractEntry) Entry(_ context.Context, query feasibility.EntryQuery) (feasibility.EntryEvidence, error) {
	if query.PassportCountry == "" {
		return feasibility.EntryEvidence{}, errors.New("no passport supplied")
	}
	evidence := feasibility.EntryEvidence{
		EntryMode: domain.VisaFree, Preparation: domain.NoPreparation, TransitRisk: domain.NoTransitRisk,
		Confidence: domain.High, SourceLabel: "fixture entry reference", ObservedAt: contractNow,
	}
	if query.DestinationCountry == "IS" {
		evidence.EntryMode = domain.EVisa
		evidence.Preparation = domain.MediumPreparation
	}
	return evidence, nil
}

type contractFlights struct{}

func (contractFlights) Flights(_ context.Context, query feasibility.FlightQuery) (feasibility.FlightEvidence, error) {
	price, err := money.New(45_000, money.Currency("USD"))
	if err != nil {
		return feasibility.FlightEvidence{}, err
	}
	return feasibility.FlightEvidence{
		Available: true, Stops: 1, DurationMinutes: 480, HasPrice: true, Price: price,
		Confidence: domain.High, ObservedAt: contractNow,
	}, nil
}

type contractCosts struct{}

func (contractCosts) Estimate(_ context.Context, query feasibility.CostQuery) (feasibility.CostEvidence, error) {
	build := func(minor int64) money.Amount {
		amount, _ := money.New(minor, money.Currency("USD"))
		return amount
	}
	nightly := int64(12_000)
	if query.DestinationID == "reykjavik-is" {
		nightly = 30_000
	}
	expected := nightly * int64(query.Nights) * int64(query.Travelers)
	return feasibility.CostEvidence{
		Minimum: build(expected * 85 / 100), Expected: build(expected), Maximum: build(expected * 120 / 100),
		Confidence: domain.Medium, Components: []string{"flight", "stay", "food"}, ObservedAt: contractNow,
	}, nil
}

type contractRates struct{}

func (contractRates) Rate(_ context.Context, from, to money.Currency, _ time.Time) (feasibility.Rate, error) {
	micros := int64(1_000_000)
	if from.String() == "USD" && to.String() == "NGN" {
		micros = 1_500_000_000
	}
	rate, err := money.NewRate(micros)
	if err != nil {
		return feasibility.Rate{}, err
	}
	return feasibility.Rate{Rate: rate, Source: "fixture rate reference", ObservedAt: contractNow}, nil
}

var contractNow = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

func discoveryService(t *testing.T, repository *store, timer clock.Clock) *discover.Service {
	t.Helper()
	scoring, err := matching.New(matching.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	providers := feasibility.Providers{Entry: contractEntry{}, Flight: contractFlights{}, Costs: contractCosts{}, Rates: contractRates{}}
	return discover.NewService(repository, repository, discover.Engines{
		Understanding: understanding.New(nil),
		Feasibility:   feasibility.New(providers, timer),
		Matching:      scoring,
		Ranking:       recommendation.New(nil, 10),
	}, timer)
}
