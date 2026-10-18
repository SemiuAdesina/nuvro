package personalization_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func newEngine(t *testing.T) *engine.Engine {
	t.Helper()
	value, err := engine.New(engine.DefaultConfig(), clock.NewFixed(now))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func signal(signalType domain.SignalType, subject domain.SubjectType, subjectID string, at time.Time) domain.Signal {
	value := domain.NewSignal("user", signalType, subject, subjectID, at)
	value.SourceEvent = subjectID + at.Format(time.RFC3339)
	return value
}

func keys(preferences []domain.LearnedPreference) map[domain.Key]domain.LearnedPreference {
	found := map[domain.Key]domain.LearnedPreference{}
	for _, preference := range preferences {
		found[preference.Key] = preference
	}
	return found
}

func TestOneWeakSignalNeverBecomesAPreference(t *testing.T) {
	learned := newEngine(t).Aggregate([]domain.Signal{
		signal(domain.DestinationSaved, domain.Interest, "coastal", now.Add(-24*time.Hour)),
	})
	if len(learned) != 0 {
		t.Fatalf("one save produced %v", learned)
	}
}

func TestAFewCasualSavesStayBelowTheThreshold(t *testing.T) {
	saves := make([]domain.Signal, 0, 3)
	for offset := 1; offset <= 3; offset++ {
		saves = append(saves, signal(domain.DestinationSaved, domain.Interest, "coastal",
			now.Add(-time.Duration(offset*24)*time.Hour)))
	}
	if learned := newEngine(t).Aggregate(saves); len(learned) != 0 {
		t.Fatalf("three casual saves produced %v", learned)
	}
}

func TestRepeatedMeaningfulChoicesBuildConfidence(t *testing.T) {
	subject := newEngine(t)
	strong := subject.Aggregate([]domain.Signal{
		signal(domain.DestinationSelected, domain.Interest, "coastal", now.Add(-72*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-48*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "coastal", now.Add(-24*time.Hour)),
	})
	if len(strong) != 1 {
		t.Fatalf("meaningful evidence produced %v", strong)
	}
	if strong[0].Explanation != domain.PostTripFeedback {
		t.Fatalf("explanation = %q", strong[0].Explanation)
	}
	if strong[0].RulesVersion != domain.RulesVersion {
		t.Fatal("the rules version was not recorded")
	}
	stronger := subject.Aggregate([]domain.Signal{
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-96*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-72*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-48*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-24*time.Hour)),
	})
	if stronger[0].Score <= strong[0].Score {
		t.Fatalf("more evidence did not score higher: %d vs %d", stronger[0].Score, strong[0].Score)
	}
	if stronger[0].Confidence.Rank() < strong[0].Confidence.Rank() {
		t.Fatal("more evidence produced lower confidence")
	}
}

func TestManyWeakSavesEventuallyCrossTheThreshold(t *testing.T) {
	saves := make([]domain.Signal, 0, 10)
	for offset := 1; offset <= 10; offset++ {
		saves = append(saves, signal(domain.DestinationSaved, domain.Interest, "coastal",
			now.Add(-time.Duration(offset*24)*time.Hour)))
	}
	learned := newEngine(t).Aggregate(saves)
	if len(learned) != 1 || learned[0].Key != domain.InterestCoastal {
		t.Fatalf("sustained saving produced %v", learned)
	}
	if learned[0].Explanation != domain.RepeatedSaves {
		t.Fatalf("explanation = %q", learned[0].Explanation)
	}
}

func TestNegativeSignalsOffsetPositiveEvidence(t *testing.T) {
	subject := newEngine(t)
	positive := []domain.Signal{
		signal(domain.DestinationSelected, domain.Interest, "nightlife", now.Add(-72*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "nightlife", now.Add(-48*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "nightlife", now.Add(-24*time.Hour)),
	}
	before := subject.Aggregate(positive)
	after := subject.Aggregate(append(positive,
		signal(domain.FeedbackNegative, domain.Interest, "nightlife", now.Add(-12*time.Hour))))
	if len(before) == 0 {
		t.Fatal("positive evidence produced nothing")
	}
	if len(after) > 0 && after[0].Score >= before[0].Score {
		t.Fatal("negative feedback did not reduce the score")
	}
}

func TestOlderSignalsInfluenceLessThanRecentOnes(t *testing.T) {
	subject := newEngine(t)
	recent := subject.Aggregate([]domain.Signal{
		signal(domain.DestinationSelected, domain.TripPace, "relaxed", now.Add(-24*time.Hour)),
		signal(domain.DestinationSelected, domain.TripPace, "relaxed", now.Add(-48*time.Hour)),
		signal(domain.FeedbackPositive, domain.TripPace, "relaxed", now.Add(-72*time.Hour)),
	})
	old := subject.Aggregate([]domain.Signal{
		signal(domain.DestinationSelected, domain.TripPace, "relaxed", now.AddDate(-2, 0, 0)),
		signal(domain.DestinationSelected, domain.TripPace, "relaxed", now.AddDate(-2, -1, 0)),
		signal(domain.FeedbackPositive, domain.TripPace, "relaxed", now.AddDate(-2, -2, 0)),
	})
	if len(recent) == 0 {
		t.Fatal("recent evidence produced nothing")
	}
	if len(old) > 0 && old[0].Score >= recent[0].Score {
		t.Fatalf("old evidence scored %d against recent %d", old[0].Score, recent[0].Score)
	}
}

func TestDecayIsDeterministicAndHalves(t *testing.T) {
	decay := engine.DefaultDecay()
	full := decay.Factor(engine.SlowDecay, 0)
	oneHalfLife := decay.Factor(engine.SlowDecay, decay.SlowHalfLife)
	twoHalfLives := decay.Factor(engine.SlowDecay, 2*decay.SlowHalfLife)
	if full != 100 || oneHalfLife != 50 || twoHalfLives != 25 {
		t.Fatalf("decay = %d %d %d", full, oneHalfLife, twoHalfLives)
	}
	if decay.Factor(engine.FastDecay, decay.FastHalfLife) != 50 {
		t.Fatal("fast decay does not halve at its own half life")
	}
	if decay.HalfLife(engine.FastDecay) >= decay.HalfLife(engine.SlowDecay) {
		t.Fatal("fast decay is not faster than slow decay")
	}
	if engine.ClassOf(domain.PaceRelaxed) != engine.FastDecay {
		t.Fatal("a temporary pace preference does not decay quickly")
	}
	if engine.ClassOf(domain.InterestCoastal) != engine.SlowDecay {
		t.Fatal("a durable interest decays quickly")
	}
}

func TestOpposingPreferencesDoNotBothSurvive(t *testing.T) {
	learned := newEngine(t).Aggregate([]domain.Signal{
		signal(domain.FeedbackPositive, domain.StayArea, "quiet", now.Add(-24*time.Hour)),
		signal(domain.FeedbackPositive, domain.StayArea, "quiet", now.Add(-48*time.Hour)),
		signal(domain.FeedbackPositive, domain.StayArea, "quiet", now.Add(-72*time.Hour)),
		signal(domain.DestinationSelected, domain.StayArea, "central", now.Add(-30*time.Hour)),
		signal(domain.DestinationSelected, domain.StayArea, "central", now.Add(-54*time.Hour)),
	})
	found := keys(learned)
	if _, quiet := found[domain.StayQuietArea]; !quiet {
		t.Fatalf("the stronger preference was dropped: %v", learned)
	}
	if _, central := found[domain.StayCentralArea]; central {
		t.Fatal("two opposing preferences were both kept")
	}
}

func TestUnmappedSubjectsAreIgnored(t *testing.T) {
	learned := newEngine(t).Aggregate([]domain.Signal{
		signal(domain.DestinationSelected, domain.Interest, "spaceflight", now.Add(-24*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "spaceflight", now.Add(-48*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "spaceflight", now.Add(-72*time.Hour)),
	})
	if len(learned) != 0 {
		t.Fatalf("an unmapped subject became a preference: %v", learned)
	}
}

func TestAggregationIsDeterministic(t *testing.T) {
	subject := newEngine(t)
	signals := []domain.Signal{
		signal(domain.FeedbackPositive, domain.Interest, "food", now.Add(-24*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "food", now.Add(-48*time.Hour)),
		signal(domain.DestinationSelected, domain.Interest, "coastal", now.Add(-24*time.Hour)),
		signal(domain.FeedbackPositive, domain.Interest, "coastal", now.Add(-48*time.Hour)),
	}
	first := subject.Aggregate(signals)
	for range 25 {
		again := subject.Aggregate(signals)
		if len(again) != len(first) {
			t.Fatal("aggregation is not deterministic")
		}
		for index := range again {
			if again[index].Key != first[index].Key || again[index].Score != first[index].Score {
				t.Fatal("aggregation ordering or scoring is not stable")
			}
		}
	}
}

func TestContextSeparatesExplicitFromLearned(t *testing.T) {
	preferences := user.Preferences{Interests: []string{"beach", "food"}, Pace: "relaxed", Crowds: "quiet"}
	learned := []domain.LearnedPreference{{
		UserID: "user", Key: domain.InterestCulture, Confidence: domain.High, Score: 20,
		EvidenceCount: 4, Status: domain.Active, RulesVersion: domain.RulesVersion,
		FirstObservedAt: now.Add(-72 * time.Hour), LastObservedAt: now,
	}}
	value := newEngine(t).Context("user", preferences, learned, true)
	if len(value.Explicit) != 4 {
		t.Fatalf("explicit = %v", value.Explicit)
	}
	for _, preference := range value.Explicit {
		if preference.Origin != domain.ExplicitProfile || preference.Weight != domain.ExplicitWeight {
			t.Fatalf("explicit preference = %+v", preference)
		}
	}
	if len(value.Learned) != 1 || value.Learned[0].Key != domain.InterestCulture {
		t.Fatalf("learned = %v", value.Learned)
	}
	if value.Learned[0].Origin != domain.InferredLearned {
		t.Fatalf("origin = %q", value.Learned[0].Origin)
	}
}

func TestDismissedPreferencesBecomeSuppressed(t *testing.T) {
	learned := []domain.LearnedPreference{{
		UserID: "user", Key: domain.StayQuietArea, Confidence: domain.High, Score: 20,
		EvidenceCount: 4, Status: domain.Dismissed, RulesVersion: domain.RulesVersion,
		FirstObservedAt: now.Add(-72 * time.Hour), LastObservedAt: now,
	}}
	value := newEngine(t).Context("user", user.Preferences{}, learned, true)
	if len(value.Learned) != 0 {
		t.Fatal("a dismissed preference still influences")
	}
	if len(value.Suppressed) != 1 || value.Suppressed[0] != domain.StayQuietArea {
		t.Fatalf("suppressed = %v", value.Suppressed)
	}
	if value.Holds(domain.StayQuietArea) {
		t.Fatal("a dismissed preference holds influence")
	}
}

func TestConfirmedPreferencesCarryMoreWeight(t *testing.T) {
	confirmed := []domain.LearnedPreference{{
		UserID: "user", Key: domain.StayQuietArea, Confidence: domain.MediumConfidence, Score: 12,
		EvidenceCount: 3, Status: domain.Confirmed, RulesVersion: domain.RulesVersion,
		FirstObservedAt: now.Add(-72 * time.Hour), LastObservedAt: now,
	}}
	value := newEngine(t).Context("user", user.Preferences{}, confirmed, true)
	if value.Learned[0].Origin != domain.ConfirmedLearned {
		t.Fatalf("origin = %q", value.Learned[0].Origin)
	}
	if value.Influence(domain.StayQuietArea) != domain.ConfirmedWeight {
		t.Fatalf("weight = %d", value.Influence(domain.StayQuietArea))
	}
}

func TestFeedbackBecomesStrongSignals(t *testing.T) {
	submission := feedback.TripFeedback{
		TripID: "trip", UserID: "user",
		Positives: []feedback.Aspect{feedback.LocalFood, feedback.QuietArea},
		Negatives: []feedback.Aspect{feedback.LongTransfers},
		Pace:      feedback.TooBusy, TravelBurden: feedback.TooMuchTravel,
		SubmittedAt: now,
	}
	signals := engine.SignalsFromFeedback(submission)
	if len(signals) < 4 {
		t.Fatalf("signals = %v", signals)
	}
	subjects := map[string]domain.SignalType{}
	for _, produced := range signals {
		if produced.Strength != domain.Strong {
			t.Fatalf("feedback produced a weak signal: %+v", produced)
		}
		if produced.Validate() != nil {
			t.Fatalf("an invalid signal was produced: %+v", produced)
		}
		if produced.SourceEvent == "" {
			t.Fatal("a feedback signal carries no source event for deduplication")
		}
		subjects[produced.SubjectID] = produced.Type
	}
	if subjects["food"] != domain.FeedbackPositive || subjects["quiet"] != domain.FeedbackPositive {
		t.Fatalf("positives = %v", subjects)
	}
	if _, found := subjects["long_transfers"]; !found {
		t.Fatal("too much driving did not become travel burden evidence")
	}
	if subjects["relaxed"] != domain.FeedbackPositive {
		t.Fatal("a too-busy trip did not become a relaxed pace signal")
	}
}

func TestConfigurationIsValidated(t *testing.T) {
	if err := engine.DefaultConfig().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*engine.Config){
		"no version":      func(c *engine.Config) { c.Version = "" },
		"no evidence":     func(c *engine.Config) { c.Thresholds.MinimumEvidence = 0 },
		"inverted scores": func(c *engine.Config) { c.Thresholds.HighScore = 1 },
		"no decay":        func(c *engine.Config) { c.Decay.SlowHalfLife = 0 },
	} {
		config := engine.DefaultConfig()
		mutate(&config)
		if _, err := engine.New(config, clock.NewFixed(now)); err == nil {
			t.Fatalf("%s configuration was accepted", name)
		}
	}
}
