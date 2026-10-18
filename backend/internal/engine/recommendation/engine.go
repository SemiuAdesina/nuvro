package recommendation

import "context"

type Engine struct {
	generator ExplanationGenerator
	limit     int
}

const defaultLimit = 20

func New(generator ExplanationGenerator, limit int) *Engine {
	if limit < 1 || limit > defaultLimit {
		limit = defaultLimit
	}
	return &Engine{generator: generator, limit: limit}
}

func (e *Engine) Rank(ctx context.Context, candidates []Candidate) Result {
	viable := make([]Candidate, 0, len(candidates))
	removed := 0
	for _, candidate := range candidates {
		if excluded(candidate) {
			removed++
			continue
		}
		viable = append(viable, candidate)
	}
	ordered := rank(viable)
	if len(ordered) > e.limit {
		ordered = ordered[:e.limit]
	}
	recommendations := make([]Recommendation, 0, len(ordered))
	for index, candidate := range ordered {
		recommendation := Recommendation{
			Destination: candidate.Destination, Match: candidate.Match, Feasibility: candidate.Feasibility,
			Reasons: reasonsFor(candidate), Rank: index + 1,
		}
		recommendation.Summary = fallbackSummary(recommendation)
		recommendations = append(recommendations, recommendation)
	}
	recommendations = e.explain(ctx, recommendations)
	return Result{Recommendations: recommendations, RankingVersion: RankingVersion, Excluded: removed}
}

func (e *Engine) explain(ctx context.Context, recommendations []Recommendation) []Recommendation {
	if e.generator == nil || len(recommendations) == 0 {
		return recommendations
	}
	explanations, err := e.generator.Explain(ctx, recommendations)
	if err != nil {
		return recommendations
	}
	return applyExplanations(recommendations, explanations)
}
