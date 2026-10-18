package matching

import (
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

type fraction struct {
	numerator   int
	denominator int
	explanation string
}

func interestFit(intent discovery.Intent, target destination.Destination) fraction {
	wanted := discovery.NormalizeTags(intent.Preferences.Interests)
	if len(wanted) == 0 {
		return fraction{numerator: 1, denominator: 2, explanation: "no_interests_supplied"}
	}
	offered := discovery.NormalizeTags(target.InterestTags)
	matched := 0
	for _, interest := range wanted {
		if slices.Contains(offered, interest) {
			matched++
		}
	}
	return fraction{numerator: matched, denominator: len(wanted), explanation: "interest_overlap"}
}

func weatherFit(intent discovery.Intent, target destination.Destination) fraction {
	preferred := intent.Preferences.Weather
	if preferred == "" || preferred == "any" {
		return fraction{numerator: 1, denominator: 1, explanation: "no_weather_preference"}
	}
	if slices.Contains(discovery.NormalizeTags(target.ClimateTags), preferred) {
		return fraction{numerator: 1, denominator: 1, explanation: "weather_match"}
	}
	return fraction{numerator: 0, denominator: 1, explanation: "weather_mismatch"}
}

func profileFit(preference, profile, neutral string) fraction {
	if preference == "" || preference == neutral {
		return fraction{numerator: 1, denominator: 1, explanation: "no_preference"}
	}
	if preference == profile {
		return fraction{numerator: 1, denominator: 1, explanation: "profile_match"}
	}
	if profile == "" || profile == neutral {
		return fraction{numerator: 1, denominator: 2, explanation: "profile_neutral"}
	}
	return fraction{numerator: 0, denominator: 1, explanation: "profile_mismatch"}
}

func flightFit(intent discovery.Intent, result feasibility.Result) fraction {
	flight := result.Flight
	if flight.Status == feasibility.FlightUnavailable {
		return fraction{numerator: 0, denominator: 1, explanation: "no_route_found"}
	}
	if flight.Status != feasibility.FlightAvailable {
		return fraction{numerator: 1, denominator: 2, explanation: "flight_evidence_unknown"}
	}
	if intent.Preferences.MaxFlightMinutes > 0 && flight.DurationMinutes > intent.Preferences.MaxFlightMinutes {
		return fraction{numerator: 0, denominator: 1, explanation: "exceeds_flight_tolerance"}
	}
	switch intent.Preferences.FlightTolerance {
	case "fewer_stops":
		return stopsFit(flight.Stops)
	case "shortest":
		return durationFit(flight.DurationMinutes)
	default:
		return fraction{numerator: 1, denominator: 1, explanation: "no_flight_preference"}
	}
}

func stopsFit(stops int) fraction {
	switch {
	case stops <= 0:
		return fraction{numerator: 1, denominator: 1, explanation: "direct_flight"}
	case stops == 1:
		return fraction{numerator: 2, denominator: 3, explanation: "one_stop"}
	default:
		return fraction{numerator: 0, denominator: 1, explanation: "multiple_stops"}
	}
}

func durationFit(minutes int) fraction {
	switch {
	case minutes <= 240:
		return fraction{numerator: 1, denominator: 1, explanation: "short_flight"}
	case minutes <= 600:
		return fraction{numerator: 2, denominator: 3, explanation: "medium_flight"}
	default:
		return fraction{numerator: 1, denominator: 3, explanation: "long_flight"}
	}
}

func budgetFit(result feasibility.Result) fraction {
	switch result.Affordability.Status {
	case feasibility.StrongFit:
		return fraction{numerator: 1, denominator: 1, explanation: "comfortably_within_budget"}
	case feasibility.GoodFit:
		return fraction{numerator: 4, denominator: 5, explanation: "within_budget"}
	case feasibility.Tight:
		return fraction{numerator: 2, denominator: 5, explanation: "budget_tight"}
	case feasibility.OverBudget:
		return fraction{numerator: 0, denominator: 1, explanation: "over_budget"}
	default:
		return fraction{numerator: 1, denominator: 2, explanation: "cost_evidence_unknown"}
	}
}

func accessFit(intent discovery.Intent, result feasibility.Result) fraction {
	simple := intent.Preferences.Paperwork == "keep_it_simple"
	switch result.Access.Status {
	case feasibility.Easy:
		return fraction{numerator: 1, denominator: 1, explanation: "easy_entry"}
	case feasibility.Manageable:
		if simple {
			return fraction{numerator: 1, denominator: 2, explanation: "some_paperwork"}
		}
		return fraction{numerator: 4, denominator: 5, explanation: "some_paperwork"}
	case feasibility.PreparationRequired:
		if simple {
			return fraction{numerator: 0, denominator: 1, explanation: "visa_preparation"}
		}
		return fraction{numerator: 2, denominator: 5, explanation: "visa_preparation"}
	case feasibility.NotFeasible:
		return fraction{numerator: 0, denominator: 1, explanation: "entry_not_feasible"}
	default:
		return fraction{numerator: 1, denominator: 2, explanation: "entry_evidence_unknown"}
	}
}

func assess(awarded, possible int) destination.Assessment {
	if possible == 0 {
		return destination.Neutral
	}
	share := (awarded * 100) / possible
	switch {
	case share >= 85:
		return destination.Strong
	case share >= 60:
		return destination.Good
	case share >= 35:
		return destination.Neutral
	default:
		return destination.Caution
	}
}
