package recommendation

import (
	"context"
	"strings"
)

type Explanation struct {
	DestinationID string
	Summary       string
}

type ExplanationGenerator interface {
	Explain(context.Context, []Recommendation) ([]Explanation, error)
}

func fallbackSummary(recommendation Recommendation) string {
	parts := make([]string, 0, 3)
	for _, component := range recommendation.Match.Components {
		if component.Assessment == "strong" {
			parts = append(parts, strings.ToLower(component.Label))
		}
		if len(parts) == 2 {
			break
		}
	}
	if len(parts) == 0 {
		return recommendation.Destination.Name
	}
	return recommendation.Destination.Name + ": " + strings.Join(parts, " and ") + " fit this trip"
}

func applyExplanations(recommendations []Recommendation, explanations []Explanation) []Recommendation {
	summaries := make(map[string]string, len(explanations))
	for _, explanation := range explanations {
		if summary := strings.TrimSpace(explanation.Summary); summary != "" && len(summary) <= 400 {
			summaries[explanation.DestinationID] = summary
		}
	}
	for index := range recommendations {
		if summary, found := summaries[recommendations[index].Destination.ID]; found {
			recommendations[index].Summary = summary
		}
	}
	return recommendations
}
