package recommendation

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

func reasonsFor(candidate Candidate) []string {
	reasons := make([]string, 0, 6)
	reasons = append(reasons, accessReason(candidate.Feasibility.Access.Status)...)
	reasons = append(reasons, affordabilityReason(candidate.Feasibility.Affordability.Status)...)
	reasons = append(reasons, flightReason(candidate.Feasibility.Flight)...)
	for _, component := range candidate.Match.Components {
		if component.Assessment == destination.Strong && component.Explanation != "" {
			reasons = append(reasons, component.Key+"_fit")
		}
	}
	return reasons
}

func accessReason(status feasibility.AccessStatus) []string {
	switch status {
	case feasibility.Easy:
		return []string{"easy_entry"}
	case feasibility.Manageable:
		return []string{"manageable_entry"}
	case feasibility.PreparationRequired:
		return []string{"visa_preparation"}
	case feasibility.NotFeasible:
		return []string{"entry_not_feasible"}
	default:
		return []string{"entry_evidence_unknown"}
	}
}

func affordabilityReason(status feasibility.AffordabilityStatus) []string {
	switch status {
	case feasibility.StrongFit:
		return []string{"comfortably_within_budget"}
	case feasibility.GoodFit:
		return []string{"within_budget"}
	case feasibility.Tight:
		return []string{"budget_tight"}
	case feasibility.OverBudget:
		return []string{"over_budget"}
	default:
		return []string{"cost_evidence_unknown"}
	}
}

func flightReason(flight feasibility.Flight) []string {
	if flight.Status != feasibility.FlightAvailable {
		return []string{"flight_evidence_unknown"}
	}
	if flight.Stops == 0 {
		return []string{"direct_flight"}
	}
	return nil
}
