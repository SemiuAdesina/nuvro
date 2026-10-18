package matching_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/matching"
)

func engine(t *testing.T) *matching.Engine {
	t.Helper()
	value, err := matching.New(matching.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func target() destination.Destination {
	return destination.Destination{
		ID: "zanzibar-tz", Name: "Zanzibar", CountryCode: "TZ", Timezone: "Africa/Dar_es_Salaam",
		PrimaryAirport: "ZNZ", ClimateTags: []string{"warm"}, InterestTags: []string{"beach", "food", "relaxation"},
		CrowdProfile: "quiet", PaceProfile: "relaxed",
	}
}

func intent() discovery.Intent {
	return discovery.Intent{
		Preferences: discovery.Preferences{
			Interests: []string{"beach", "food"}, Weather: "warm", Crowds: "quiet",
			Pace: "relaxed", FlightTolerance: "fewer_stops",
		},
	}
}

func strongResult(t *testing.T) feasibility.Result {
	t.Helper()
	return feasibility.Result{
		Access: feasibility.Access{Status: feasibility.Easy, EntryMode: feasibility.VisaFree, Confidence: feasibility.High},
		Affordability: feasibility.Affordability{
			Status: feasibility.StrongFit, Budget: amount(t, 500_000_000, "NGN"), Confidence: feasibility.High, Ratio: 60,
		},
		Flight:  feasibility.Flight{Status: feasibility.FlightAvailable, Stops: 0, DurationMinutes: 300, Confidence: feasibility.High},
		Overall: feasibility.Strong,
	}
}

func TestDefaultConfigTotalsOneHundredPoints(t *testing.T) {
	if err := matching.DefaultConfig().Validate(); err != nil {
		t.Fatal(err)
	}
	if matching.DefaultConfig().Version != matching.Version {
		t.Fatal("configuration version drifted from the package version")
	}
	for name, config := range map[string]matching.Config{
		"short":     {Version: "x", Weights: []matching.Weight{{Key: matching.Interests, Label: "Interests", PossiblePoints: 50}}},
		"duplicate": {Version: "x", Weights: []matching.Weight{{Key: matching.Interests, Label: "A", PossiblePoints: 50}, {Key: matching.Interests, Label: "B", PossiblePoints: 50}}},
		"unnamed":   {Version: "", Weights: matching.DefaultConfig().Weights},
		"negative":  {Version: "x", Weights: []matching.Weight{{Key: matching.Interests, Label: "A", PossiblePoints: -10}, {Key: matching.BudgetFit, Label: "B", PossiblePoints: 110}}},
	} {
		if _, err := matching.New(config); err == nil {
			t.Fatalf("%s configuration was accepted", name)
		}
	}
}

func TestPerfectFitScoresOneHundredAndIsExplainable(t *testing.T) {
	score, err := engine(t).Score(intent(), target(), strongResult(t), "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if score.Score != 100 {
		t.Fatalf("score = %d, want 100", score.Score)
	}
	if score.WeightsVersion != matching.Version || score.EvidenceVersion != "evidence-v1" {
		t.Fatalf("provenance missing: %+v", score)
	}
	total := 0
	for _, component := range score.Components {
		if component.Key == "" || component.Label == "" || component.Explanation == "" {
			t.Fatalf("component is not explainable: %+v", component)
		}
		total += component.PossiblePoints
	}
	if total != 100 || len(score.Components) != 7 {
		t.Fatalf("components do not add up: %d over %d components", total, len(score.Components))
	}
}

func TestScoreIsDeterministicAcrossRepeatedCalls(t *testing.T) {
	subject := engine(t)
	first, err := subject.Score(intent(), target(), strongResult(t), "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	for range 25 {
		again, err := subject.Score(intent(), target(), strongResult(t), "evidence-v1")
		if err != nil || again.Score != first.Score || len(again.Components) != len(first.Components) {
			t.Fatalf("score is not deterministic: %d then %d", first.Score, again.Score)
		}
		for index := range again.Components {
			if again.Components[index] != first.Components[index] {
				t.Fatal("component output changed between identical calls")
			}
		}
	}
}

func TestScoreStaysWithinBoundsForHostileEvidence(t *testing.T) {
	subject := engine(t)
	worst := feasibility.Result{
		Access:        feasibility.Access{Status: feasibility.NotFeasible, EntryMode: feasibility.VisaRequired, Confidence: feasibility.High},
		Affordability: feasibility.Affordability{Status: feasibility.OverBudget, Budget: amount(t, 1, "NGN"), Confidence: feasibility.High, Ratio: 900},
		Flight:        feasibility.Flight{Status: feasibility.FlightUnavailable, Confidence: feasibility.Low},
		Overall:       feasibility.Infeasible,
	}
	mismatch := discovery.Intent{Preferences: discovery.Preferences{
		Interests: []string{"skiing"}, Weather: "cold", Crowds: "lively", Pace: "packed", FlightTolerance: "shortest",
	}}
	score, err := subject.Score(mismatch, target(), worst, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if score.Score != 0 {
		t.Fatalf("worst case score = %d, want 0", score.Score)
	}
	unknown := feasibility.Result{
		Access:        feasibility.UnknownAccessAssessment("entry_evidence_unavailable"),
		Affordability: feasibility.UnknownAffordability(amount(t, 500_000_000, "NGN"), "cost_evidence_unavailable"),
		Flight:        feasibility.UnknownFlight(),
		Overall:       feasibility.UnknownOverall,
	}
	partial, err := subject.Score(intent(), target(), unknown, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if partial.Score <= 0 || partial.Score >= 100 {
		t.Fatalf("unknown evidence produced %d, want a partial score", partial.Score)
	}
}

func TestPaperworkPreferencePenalisesPreparation(t *testing.T) {
	subject := engine(t)
	result := strongResult(t)
	result.Access = feasibility.Access{Status: feasibility.PreparationRequired, EntryMode: feasibility.VisaRequired, Confidence: feasibility.High}
	relaxed := intent()
	relaxed.Preferences.Paperwork = "any"
	strict := intent()
	strict.Preferences.Paperwork = "keep_it_simple"
	lenientScore, err := subject.Score(relaxed, target(), result, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	strictScore, err := subject.Score(strict, target(), result, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if strictScore.Score >= lenientScore.Score {
		t.Fatalf("keep it simple did not penalise preparation: %d vs %d", strictScore.Score, lenientScore.Score)
	}
}

func TestScoreIgnoresEverythingOutsideItsDeclaredInputs(t *testing.T) {
	subject := engine(t)
	baseline, err := subject.Score(intent(), target(), strongResult(t), "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	sponsored := target()
	sponsored.StyleTags = []string{"sponsored", "affiliate_partner", "commission_high"}
	sponsored.Region = "Paid placement"
	promoted, err := subject.Score(intent(), sponsored, strongResult(t), "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if promoted.Score != baseline.Score {
		t.Fatalf("commercial metadata changed the score: %d vs %d", promoted.Score, baseline.Score)
	}
}

func TestObservationTimeNeverEntersTheScore(t *testing.T) {
	subject := engine(t)
	first := strongResult(t)
	first.EvaluatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	second := strongResult(t)
	second.EvaluatedAt = time.Date(2027, 6, 30, 23, 59, 0, 0, time.UTC)
	early, err := subject.Score(intent(), target(), first, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	late, err := subject.Score(intent(), target(), second, "evidence-v1")
	if err != nil {
		t.Fatal(err)
	}
	if early.Score != late.Score {
		t.Fatal("the score depends on wall clock time")
	}
}
