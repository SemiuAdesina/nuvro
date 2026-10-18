package personalization_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func contextWith(explicit, learned []personalization.Preference, suppressed []personalization.Key, enabled bool) personalization.Context {
	value := personalization.EmptyContext("user", enabled, now)
	value.Explicit = explicit
	value.Learned = learned
	value.Suppressed = suppressed
	return value
}

func preference(key personalization.Key, origin personalization.Origin, confidence personalization.Confidence) personalization.Preference {
	return personalization.Preference{
		Key: key, Origin: origin, Confidence: confidence,
		Weight: personalization.WeightFor(origin, confidence),
	}
}

func TestExplicitPreferenceOutranksEveryInference(t *testing.T) {
	value := contextWith(
		[]personalization.Preference{preference(personalization.InterestCoastal, personalization.ExplicitProfile, personalization.High)},
		[]personalization.Preference{preference(personalization.InterestCoastal, personalization.InferredLearned, personalization.High)},
		nil, true,
	)
	if value.Influence(personalization.InterestCoastal) != personalization.ExplicitWeight {
		t.Fatalf("influence = %d", value.Influence(personalization.InterestCoastal))
	}
	origin, found := value.OriginOf(personalization.InterestCoastal)
	if !found || origin != personalization.ExplicitProfile {
		t.Fatalf("origin = %q", origin)
	}
}

func TestConfirmedLearningOutranksInferred(t *testing.T) {
	confirmed := contextWith(nil,
		[]personalization.Preference{preference(personalization.StayQuietArea, personalization.ConfirmedLearned, personalization.MediumConfidence)},
		nil, true)
	inferred := contextWith(nil,
		[]personalization.Preference{preference(personalization.StayQuietArea, personalization.InferredLearned, personalization.High)},
		nil, true)
	if confirmed.Influence(personalization.StayQuietArea) <= inferred.Influence(personalization.StayQuietArea) {
		t.Fatal("a confirmed preference did not outweigh an inferred one")
	}
	if personalization.ExplicitProfile.Rank() <= personalization.ConfirmedLearned.Rank() {
		t.Fatal("origin ranking is wrong")
	}
	if personalization.ConfirmedLearned.Rank() <= personalization.InferredLearned.Rank() {
		t.Fatal("origin ranking is wrong")
	}
}

func TestLowConfidenceInferenceCannotDominate(t *testing.T) {
	low := personalization.WeightFor(personalization.InferredLearned, personalization.Low)
	high := personalization.WeightFor(personalization.InferredLearned, personalization.High)
	if low >= high {
		t.Fatal("confidence does not scale influence")
	}
	if high >= personalization.WeightFor(personalization.ConfirmedLearned, personalization.Low) {
		t.Fatal("a high-confidence inference outweighed a confirmed preference")
	}
}

func TestDisabledLearningRemovesEveryLearnedPreference(t *testing.T) {
	value := contextWith(
		[]personalization.Preference{preference(personalization.InterestFood, personalization.ExplicitProfile, personalization.High)},
		[]personalization.Preference{preference(personalization.StayQuietArea, personalization.ConfirmedLearned, personalization.High)},
		nil, false,
	)
	if value.Holds(personalization.StayQuietArea) {
		t.Fatal("learned preferences influenced a context with learning disabled")
	}
	if !value.Holds(personalization.InterestFood) {
		t.Fatal("an explicit preference was lost when learning was disabled")
	}
}

func TestSuppressedPreferencesHaveNoInfluence(t *testing.T) {
	value := contextWith(nil,
		[]personalization.Preference{preference(personalization.StayQuietArea, personalization.InferredLearned, personalization.High)},
		[]personalization.Key{personalization.StayQuietArea}, true)
	if value.Holds(personalization.StayQuietArea) {
		t.Fatal("a suppressed preference still influenced the context")
	}
	if _, found := value.OriginOf(personalization.StayQuietArea); found {
		t.Fatal("a suppressed preference reported an origin")
	}
}

func TestEmptyContextIsUsableForColdStart(t *testing.T) {
	value := personalization.EmptyContext("user", true, now)
	if value.Holds(personalization.InterestCoastal) {
		t.Fatal("a new user already holds preferences")
	}
	if value.Explicit == nil || value.Learned == nil || value.Suppressed == nil {
		t.Fatal("a cold start context has nil collections")
	}
	if value.RulesVersion != personalization.RulesVersion {
		t.Fatalf("rules version = %q", value.RulesVersion)
	}
}

func TestOpposingKeysAreRecognised(t *testing.T) {
	if !personalization.StayQuietArea.Opposes(personalization.StayCentralArea) {
		t.Fatal("quiet and central are not recognised as opposites")
	}
	if !personalization.PaceRelaxed.Opposes(personalization.PacePacked) {
		t.Fatal("relaxed and packed are not recognised as opposites")
	}
	if personalization.InterestFood.Opposes(personalization.InterestCoastal) {
		t.Fatal("unrelated interests were treated as opposites")
	}
}

func TestKeyCategories(t *testing.T) {
	if personalization.InterestCoastal.Category() != "interest" {
		t.Fatalf("category = %q", personalization.InterestCoastal.Category())
	}
	if personalization.StayQuietArea.Category() != "stay" {
		t.Fatalf("category = %q", personalization.StayQuietArea.Category())
	}
	if personalization.Key("invented").Valid() {
		t.Fatal("an invented key was accepted")
	}
}

func TestInsightStatusLifecycle(t *testing.T) {
	if !personalization.Active.CanTransitionTo(personalization.Confirmed) {
		t.Fatal("an active insight could not be confirmed")
	}
	if !personalization.Active.CanTransitionTo(personalization.Dismissed) {
		t.Fatal("an active insight could not be dismissed")
	}
	if !personalization.Confirmed.CanTransitionTo(personalization.Dismissed) {
		t.Fatal("a confirmed insight could not be dismissed")
	}
	if personalization.Dismissed.CanTransitionTo(personalization.Confirmed) {
		t.Fatal("a dismissed insight was confirmed without becoming active again")
	}
	if personalization.Active.CanTransitionTo("invented") {
		t.Fatal("an invented status was accepted")
	}
	if !personalization.Active.Influences() || !personalization.Confirmed.Influences() {
		t.Fatal("an influencing status was ignored")
	}
	if personalization.Dismissed.Influences() || personalization.Suppressed.Influences() {
		t.Fatal("a dismissed or suppressed status still influences")
	}
}
