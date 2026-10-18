package alerting

import (
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
)

type Engine struct{ evaluators []Evaluator }

func New(evaluators ...Evaluator) *Engine {
	if len(evaluators) == 0 {
		evaluators = []Evaluator{WeatherImpact{}, LeaveTime{}, ScheduleConflict{}}
	}
	return &Engine{evaluators: evaluators}
}

func (e *Engine) Names() []string {
	names := make([]string, 0, len(e.evaluators))
	for _, evaluator := range e.evaluators {
		names = append(names, evaluator.Name())
	}
	return names
}

func (e *Engine) Evaluate(input Input) []alert.Alert {
	produced := make([]alert.Alert, 0)
	seen := make(map[string]struct{})
	for _, evaluator := range e.evaluators {
		for _, candidate := range evaluator.Evaluate(input) {
			if candidate.Validate() != nil {
				continue
			}
			if _, duplicate := seen[candidate.Fingerprint]; duplicate {
				continue
			}
			seen[candidate.Fingerprint] = struct{}{}
			produced = append(produced, candidate)
		}
	}
	slices.SortStableFunc(produced, func(left, right alert.Alert) int {
		if order := rank(right.Severity) - rank(left.Severity); order != 0 {
			return order
		}
		return left.EffectiveAt.Compare(right.EffectiveAt)
	})
	return produced
}

func rank(severity alert.Severity) int {
	switch severity {
	case alert.Important:
		return 3
	case alert.Attention:
		return 2
	default:
		return 1
	}
}
