package recommendation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
)

func candidate(id string, score int, overall feasibility.Overall, access feasibility.AccessStatus, ratio int) recommendation.Candidate {
	return recommendation.Candidate{
		Destination: destination.Destination{ID: id, Name: id, CountryCode: "PT", PrimaryAirport: "LIS", Timezone: "Europe/Lisbon"},
		Match: destination.MatchScore{
			Score: score, WeightsVersion: "match-v1", EvidenceVersion: "match-v1",
			Components: []destination.MatchComponent{{Key: "interests", Label: "Interests", AwardedPoints: score / 4, PossiblePoints: 25, Assessment: destination.Strong, Explanation: "interest_overlap"}},
		},
		Feasibility: feasibility.Result{
			Overall: overall,
			Access:  feasibility.Access{Status: access, EntryMode: feasibility.VisaFree, Confidence: feasibility.High},
			Affordability: feasibility.Affordability{
				Status: feasibility.GoodFit, Confidence: feasibility.High, HasEstimate: ratio >= 0, Ratio: ratio,
			},
			Flight: feasibility.Flight{Status: feasibility.FlightAvailable, Stops: 0, Confidence: feasibility.High},
		},
	}
}

func identifiers(result recommendation.Result) []string {
	found := make([]string, 0, len(result.Recommendations))
	for _, value := range result.Recommendations {
		found = append(found, value.Destination.ID)
	}
	return found
}

func TestHigherMatchRanksFirst(t *testing.T) {
	engine := recommendation.New(nil, 10)
	result := engine.Rank(t.Context(), []recommendation.Candidate{
		candidate("b", 70, feasibility.Strong, feasibility.Easy, 50),
		candidate("a", 90, feasibility.Viable, feasibility.Manageable, 80),
	})
	if got := identifiers(result); got[0] != "a" || got[1] != "b" {
		t.Fatalf("order = %v", got)
	}
	if result.Recommendations[0].Rank != 1 || result.Recommendations[1].Rank != 2 {
		t.Fatal("ranks were not assigned in order")
	}
	if result.RankingVersion != recommendation.RankingVersion {
		t.Fatal("ranking version missing")
	}
}

func TestTieBreakingIsDeterministicAndDocumented(t *testing.T) {
	engine := recommendation.New(nil, 10)
	tied := []recommendation.Candidate{
		candidate("zulu", 80, feasibility.Viable, feasibility.Easy, 60),
		candidate("alpha", 80, feasibility.Viable, feasibility.Easy, 60),
		candidate("mike", 80, feasibility.Strong, feasibility.Easy, 60),
		candidate("kilo", 80, feasibility.Viable, feasibility.Easy, 40),
		candidate("papa", 80, feasibility.Viable, feasibility.Manageable, 60),
	}
	first := identifiers(engine.Rank(t.Context(), tied))
	want := []string{"mike", "kilo", "alpha", "zulu", "papa"}
	for index, expected := range want {
		if first[index] != expected {
			t.Fatalf("order = %v, want %v", first, want)
		}
	}
	for range 20 {
		if got := identifiers(engine.Rank(t.Context(), tied)); got[0] != first[0] || got[4] != first[4] {
			t.Fatalf("ordering is not stable: %v then %v", first, got)
		}
	}
}

func TestUnknownCostRanksBelowKnownCost(t *testing.T) {
	engine := recommendation.New(nil, 10)
	result := engine.Rank(t.Context(), []recommendation.Candidate{
		candidate("unknown-cost", 80, feasibility.Viable, feasibility.Easy, -1),
		candidate("known-cost", 80, feasibility.Viable, feasibility.Easy, 95),
	})
	if got := identifiers(result); got[0] != "known-cost" {
		t.Fatalf("order = %v", got)
	}
}

func TestInfeasibleCandidatesAreExcluded(t *testing.T) {
	engine := recommendation.New(nil, 10)
	result := engine.Rank(t.Context(), []recommendation.Candidate{
		candidate("blocked", 99, feasibility.Infeasible, feasibility.NotFeasible, 20),
		candidate("open", 60, feasibility.Viable, feasibility.Easy, 60),
	})
	if len(result.Recommendations) != 1 || result.Recommendations[0].Destination.ID != "open" {
		t.Fatalf("infeasible destination survived: %v", identifiers(result))
	}
	if result.Excluded != 1 {
		t.Fatalf("excluded = %d, want 1", result.Excluded)
	}
}

func TestReasonsAreStableCodesAndLimitIsRespected(t *testing.T) {
	engine := recommendation.New(nil, 2)
	result := engine.Rank(t.Context(), []recommendation.Candidate{
		candidate("a", 90, feasibility.Strong, feasibility.Easy, 40),
		candidate("b", 80, feasibility.Viable, feasibility.PreparationRequired, 60),
		candidate("c", 70, feasibility.Viable, feasibility.Easy, 60),
	})
	if len(result.Recommendations) != 2 {
		t.Fatalf("limit ignored: %d recommendations", len(result.Recommendations))
	}
	reasons := result.Recommendations[0].Reasons
	if len(reasons) == 0 {
		t.Fatal("no reasons returned")
	}
	for _, reason := range reasons {
		for _, character := range reason {
			if character >= 'A' && character <= 'Z' || character == ' ' {
				t.Fatalf("reason %q is prose, not a stable code", reason)
			}
		}
	}
}

type generator struct {
	explanations []recommendation.Explanation
	err          error
}

func (g generator) Explain(context.Context, []recommendation.Recommendation) ([]recommendation.Explanation, error) {
	return g.explanations, g.err
}

func TestExplanationNeverAltersFactsAndFailsSafely(t *testing.T) {
	candidates := []recommendation.Candidate{
		candidate("a", 90, feasibility.Strong, feasibility.Easy, 40),
		candidate("b", 60, feasibility.Viable, feasibility.Easy, 60),
	}
	deterministic := recommendation.New(nil, 10).Rank(t.Context(), candidates)
	if deterministic.Recommendations[0].Summary == "" {
		t.Fatal("no deterministic fallback summary")
	}
	explained := recommendation.New(generator{explanations: []recommendation.Explanation{{DestinationID: "a", Summary: "Warm beaches and short flights."}}}, 10).
		Rank(t.Context(), candidates)
	if explained.Recommendations[0].Summary != "Warm beaches and short flights." {
		t.Fatalf("explanation was not applied: %q", explained.Recommendations[0].Summary)
	}
	if explained.Recommendations[0].Match.Score != deterministic.Recommendations[0].Match.Score {
		t.Fatal("explanation changed the score")
	}
	if explained.Recommendations[0].Destination.ID != deterministic.Recommendations[0].Destination.ID {
		t.Fatal("explanation changed the ranking")
	}
	if explained.Recommendations[1].Summary != deterministic.Recommendations[1].Summary {
		t.Fatal("an unexplained recommendation lost its fallback")
	}
	failed := recommendation.New(generator{err: errors.New("model unavailable")}, 10).Rank(t.Context(), candidates)
	if len(failed.Recommendations) != 2 || failed.Recommendations[0].Summary == "" {
		t.Fatal("a failed generator broke the response")
	}
}

func TestOversizedOrUnknownExplanationsAreIgnored(t *testing.T) {
	candidates := []recommendation.Candidate{candidate("a", 90, feasibility.Strong, feasibility.Easy, 40)}
	oversized := make([]byte, 401)
	for index := range oversized {
		oversized[index] = 'x'
	}
	result := recommendation.New(generator{explanations: []recommendation.Explanation{
		{DestinationID: "a", Summary: string(oversized)},
		{DestinationID: "does-not-exist", Summary: "ignored"},
	}}, 10).Rank(t.Context(), candidates)
	if len(result.Recommendations[0].Summary) > 400 {
		t.Fatal("an oversized summary was accepted")
	}
}
