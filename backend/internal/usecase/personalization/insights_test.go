package personalization_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

func insightFor(t *testing.T, service *usecase.Service, key domain.Key) domain.Insight {
	t.Helper()
	insights, err := service.Insights(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	for _, insight := range insights {
		if insight.Key == key {
			return insight
		}
	}
	t.Fatalf("%q was not learned: %v", key, insights)
	return domain.Insight{}
}

func holds(t *testing.T, service *usecase.Service, key domain.Key) bool {
	t.Helper()
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	return value.Holds(key)
}

func submitFeedback(t *testing.T, service *usecase.Service) {
	t.Helper()
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
}

func TestDismissingAnInsightSuppressesItsInfluence(t *testing.T) {
	service := newService(t, newStore(true), newFeedbackStore(), newCache(), user.Preferences{})
	recordHistory(t, service)
	submitFeedback(t, service)
	target := insightFor(t, service, domain.StayQuietArea)
	updated, err := service.UpdateInsight(t.Context(), principal(), target.ID, domain.Dismissed)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.Dismissed {
		t.Fatalf("status = %q", updated.Status)
	}
	if holds(t, service, domain.StayQuietArea) {
		t.Fatal("a dismissed insight still influences recommendations")
	}
	submitFeedback(t, service)
	if holds(t, service, domain.StayQuietArea) {
		t.Fatal("more evidence resurrected a dismissed insight")
	}
}

func TestConfirmingAnInsightStrengthensIt(t *testing.T) {
	service := newService(t, newStore(true), newFeedbackStore(), newCache(), user.Preferences{})
	recordHistory(t, service)
	submitFeedback(t, service)
	target := insightFor(t, service, domain.InterestCoastal)
	before, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if _, updateErr := service.UpdateInsight(t.Context(), principal(), target.ID, domain.Confirmed); updateErr != nil {
		t.Fatal(updateErr)
	}
	after, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if after.Influence(target.Key) <= before.Influence(target.Key) {
		t.Fatalf("confirmation did not strengthen influence: %d then %d",
			before.Influence(target.Key), after.Influence(target.Key))
	}
	if after.Influence(target.Key) >= domain.ExplicitWeight {
		t.Fatal("a confirmed inference reached explicit weight")
	}
}

func TestInsightStatusRulesAndOwnership(t *testing.T) {
	repository := newStore(true)
	service := newService(t, repository, newFeedbackStore(), newCache(), user.Preferences{})
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	insights, err := service.Insights(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	target := insights[0]
	if _, err := service.UpdateInsight(t.Context(), principal(), target.ID, domain.Dismissed); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateInsight(t.Context(), principal(), target.ID, domain.Confirmed); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatal("a dismissed insight was confirmed directly")
	}
	stranger := user.Principal{UserID: "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"}
	if _, err := service.UpdateInsight(t.Context(), stranger, target.ID, domain.Confirmed); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("another traveller updated an insight")
	}
}

func TestClearingLearningKeepsExplicitPreferences(t *testing.T) {
	repository := newStore(true)
	explicit := user.Preferences{Interests: []string{"culture"}}
	service := newService(t, repository, newFeedbackStore(), newCache(), explicit)
	recordHistory(t, service)
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	if err := service.ClearLearning(t.Context(), principal()); err != nil {
		t.Fatal(err)
	}
	if repository.cleared != 1 || repository.signalsCleared != 1 {
		t.Fatal("clearing did not remove both preferences and signals")
	}
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Learned) != 0 {
		t.Fatalf("learned preferences survived the clear: %v", value.Learned)
	}
	if !value.Holds(domain.InterestCulture) {
		t.Fatal("clearing learning removed an explicit preference")
	}
	if !value.HistoryLearningEnabled {
		t.Fatal("clearing learning also switched the toggle off")
	}
}
