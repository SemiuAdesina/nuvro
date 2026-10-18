package feasibility

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Engine struct {
	entry  EntryRequirementProvider
	flight FlightEvidenceProvider
	costs  TripCostEstimator
	rates  FXRateProvider
	timer  clock.Clock
}

type Providers struct {
	Entry  EntryRequirementProvider
	Flight FlightEvidenceProvider
	Costs  TripCostEstimator
	Rates  FXRateProvider
}

func New(providers Providers, timer clock.Clock) *Engine {
	return &Engine{entry: providers.Entry, flight: providers.Flight, costs: providers.Costs, rates: providers.Rates, timer: timer}
}

func (e *Engine) Evaluate(ctx context.Context, intent discovery.Intent, target destination.Destination) feasibility.Result {
	access := e.access(ctx, intent, target)
	flightEvidence, flightAvailable := e.flightEvidence(ctx, intent, target)
	cost, costAvailable := e.costEvidence(ctx, intent, target, flightEvidence)
	affordability := e.affordability(ctx, intent, cost, costAvailable)
	result := feasibility.Result{
		Access:        access,
		Affordability: affordability,
		Flight:        summarize(flightEvidence, flightAvailable),
		EvaluatedAt:   e.timer.Now().UTC(),
	}
	result.Overall = overallOf(result)
	result.Cautions = cautionsFor(result, intent)
	return result
}

func overallOf(result feasibility.Result) feasibility.Overall {
	if result.Access.Blocks() || result.Affordability.Blocks() || result.Flight.Status == feasibility.FlightUnavailable {
		return feasibility.Infeasible
	}
	if result.Access.Status == feasibility.UnknownAccess || result.Affordability.Status == feasibility.InsufficientEvidence {
		return feasibility.UnknownOverall
	}
	strength := result.Access.Rank() + result.Affordability.Rank()
	switch {
	case strength >= 7 && result.Flight.Status == feasibility.FlightAvailable:
		return feasibility.Strong
	case strength >= 5:
		return feasibility.Viable
	default:
		return feasibility.Challenging
	}
}

func cautionsFor(result feasibility.Result, intent discovery.Intent) []string {
	var cautions []string
	if result.Access.Status == feasibility.PreparationRequired {
		cautions = append(cautions, "visa_preparation")
	}
	if result.Access.Status == feasibility.UnknownAccess {
		cautions = append(cautions, "entry_evidence_unknown")
	}
	if result.Access.TransitRisk == feasibility.TransitReviewNeeded {
		cautions = append(cautions, "transit_review_required")
	}
	if result.Affordability.Status == feasibility.Tight {
		cautions = append(cautions, "budget_tight")
	}
	if result.Affordability.Status == feasibility.InsufficientEvidence {
		cautions = append(cautions, "cost_evidence_unknown")
	}
	if result.Flight.Status != feasibility.FlightAvailable {
		cautions = append(cautions, "flight_evidence_unknown")
	}
	if intent.Preferences.MaxFlightMinutes > 0 && result.Flight.DurationMinutes > intent.Preferences.MaxFlightMinutes {
		cautions = append(cautions, "long_flight")
	}
	return cautions
}
