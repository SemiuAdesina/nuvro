package discover_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/matching"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

var evaluatedAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

type profiles struct {
	profile     user.Profile
	preferences user.Preferences
	err         error
}

func (p profiles) Profile(context.Context, string) (user.Profile, error) {
	return p.profile, p.err
}

func (p profiles) Preferences(context.Context, string) (user.Preferences, error) {
	return p.preferences, p.err
}

type catalog struct {
	destinations []destination.Destination
	err          error
	limit        int32
}

func (c *catalog) Destinations(_ context.Context, limit int32) ([]destination.Destination, error) {
	c.limit = limit
	return c.destinations, c.err
}

type entryRule struct {
	mode        domain.EntryMode
	preparation domain.Preparation
}

type entryProvider struct {
	rules map[string]map[string]entryRule
}

func (e entryProvider) Entry(_ context.Context, query engine.EntryQuery) (engine.EntryEvidence, error) {
	byDestination, found := e.rules[query.PassportCountry]
	if !found {
		return engine.EntryEvidence{}, errors.New("no fixture for this passport")
	}
	rule, found := byDestination[query.DestinationCountry]
	if !found {
		return engine.EntryEvidence{}, errors.New("no fixture for this destination")
	}
	return engine.EntryEvidence{
		EntryMode: rule.mode, Preparation: rule.preparation, TransitRisk: domain.NoTransitRisk,
		Confidence: domain.High, SourceLabel: "fixture entry reference", ObservedAt: evaluatedAt,
	}, nil
}

type flightProvider struct {
	routes map[string]engine.FlightEvidence
	failed map[string]bool
}

func (f flightProvider) Flights(_ context.Context, query engine.FlightQuery) (engine.FlightEvidence, error) {
	key := query.OriginAirport + "-" + query.DestinationAirport
	if f.failed[key] {
		return engine.FlightEvidence{}, errors.New("flight provider unavailable")
	}
	evidence, found := f.routes[key]
	if !found {
		return engine.FlightEvidence{Available: false, Confidence: domain.High, ObservedAt: evaluatedAt}, nil
	}
	evidence.Confidence = domain.High
	evidence.ObservedAt = evaluatedAt
	return evidence, nil
}

type costProvider struct {
	perNight map[string]int64
	currency string
}

func (c costProvider) Estimate(_ context.Context, query engine.CostQuery) (engine.CostEvidence, error) {
	nightly, found := c.perNight[query.DestinationID]
	if !found {
		return engine.CostEvidence{}, errors.New("no cost fixture")
	}
	expected := nightly * int64(query.Nights) * int64(query.Travelers)
	build := func(minor int64) money.Amount {
		amount, _ := money.New(minor, money.Currency(c.currency))
		return amount
	}
	return engine.CostEvidence{
		Minimum: build(expected * 85 / 100), Expected: build(expected), Maximum: build(expected * 120 / 100),
		Confidence: domain.Medium, Components: []string{"flight", "stay", "food"}, ObservedAt: evaluatedAt,
	}, nil
}

type rateProvider struct{ micros map[string]int64 }

func (r rateProvider) Rate(_ context.Context, from, to money.Currency, _ time.Time) (engine.Rate, error) {
	micros, found := r.micros[from.String()+"-"+to.String()]
	if !found {
		return engine.Rate{}, errors.New("no rate fixture")
	}
	rate, err := money.NewRate(micros)
	if err != nil {
		return engine.Rate{}, err
	}
	return engine.Rate{Rate: rate, Source: "fixture rate reference", ObservedAt: evaluatedAt}, nil
}

func newService(t *testing.T, reader discover.ProfileReader, destinations *catalog, providers engine.Providers) *discover.Service {
	t.Helper()
	scoring, err := matching.New(matching.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	timer := clock.NewFixed(evaluatedAt)
	return discover.NewService(reader, destinations, discover.Engines{
		Understanding: understanding.New(nil),
		Feasibility:   engine.New(providers, timer),
		Matching:      scoring,
		Ranking:       recommendation.New(nil, 10),
	}, timer)
}

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func place(id, country, airport string, climate, interests []string, crowd string) destination.Destination {
	return destination.Destination{
		ID: id, Name: id, CountryCode: country, PrimaryAirport: airport, Timezone: "UTC",
		ClimateTags: climate, InterestTags: interests, CrowdProfile: crowd, PaceProfile: "relaxed",
	}
}
