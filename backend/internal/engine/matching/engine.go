package matching

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

type Engine struct{ config Config }

func New(config Config) (*Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Engine{config: config}, nil
}

func (e *Engine) Version() string { return e.config.Version }

func (e *Engine) Score(intent discovery.Intent, target destination.Destination, result feasibility.Result, evidenceVersion string) (destination.MatchScore, error) {
	components := make([]destination.MatchComponent, 0, len(e.config.Weights))
	total := 0
	for _, weight := range e.config.Weights {
		value := e.fractionFor(weight.Key, intent, target, result)
		awarded := award(value, weight.PossiblePoints)
		total += awarded
		components = append(components, destination.MatchComponent{
			Key: string(weight.Key), Label: weight.Label, AwardedPoints: awarded,
			PossiblePoints: weight.PossiblePoints, Assessment: assess(awarded, weight.PossiblePoints),
			Explanation: value.explanation,
		})
	}
	score := destination.MatchScore{
		Score: total, Components: components, WeightsVersion: e.config.Version, EvidenceVersion: evidenceVersion,
	}
	if err := score.Validate(); err != nil {
		return destination.MatchScore{}, err
	}
	return score, nil
}

func (e *Engine) fractionFor(key ComponentKey, intent discovery.Intent, target destination.Destination, result feasibility.Result) fraction {
	switch key {
	case Interests:
		return interestFit(intent, target)
	case Weather:
		return weatherFit(intent, target)
	case Crowds:
		return profileFit(intent.Preferences.Crowds, target.CrowdProfile, "any")
	case Pace:
		return profileFit(intent.Preferences.Pace, target.PaceProfile, "balanced")
	case FlightFit:
		return flightFit(intent, result)
	case BudgetFit:
		return budgetFit(result)
	case AccessFit:
		return accessFit(intent, result)
	default:
		return fraction{numerator: 0, denominator: 1, explanation: "unknown_component"}
	}
}

func award(value fraction, possible int) int {
	if value.denominator <= 0 || value.numerator <= 0 {
		return 0
	}
	if value.numerator >= value.denominator {
		return possible
	}
	return (value.numerator * possible) / value.denominator
}
