package personalization_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

func coastalFoodTrip() usecase.FeedbackRequest {
	return usecase.FeedbackRequest{
		TripID:       tripID,
		Positives:    []string{string(feedbackdomain.CoastalPlaces), string(feedbackdomain.LocalFood), string(feedbackdomain.QuietArea)},
		Negatives:    []string{string(feedbackdomain.LongTransfers)},
		Pace:         string(feedbackdomain.TooBusy),
		TravelBurden: string(feedbackdomain.TooMuchTravel),
		ReturnIntent: string(feedbackdomain.WouldReturn),
	}
}

func saveSignal(subject domain.SubjectType, subjectID string, offsetHours int) domain.Signal {
	signal := domain.NewSignal(owner, domain.DestinationSaved, subject, subjectID,
		now.Add(-time.Duration(offsetHours)*time.Hour))
	signal.SourceEvent = "save:" + subjectID + ":" + strconv.Itoa(offsetHours)
	return signal
}

func priorEvidence() []domain.Signal {
	signals := make([]domain.Signal, 0, 8)
	for _, offset := range []int{240, 216, 192} {
		signals = append(signals, saveSignal(domain.Interest, "coastal", offset))
	}
	for _, offset := range []int{168, 144} {
		signals = append(signals, saveSignal(domain.Interest, "food", offset))
	}
	for _, offset := range []int{120, 96} {
		signals = append(signals, saveSignal(domain.StayArea, "quiet", offset))
	}
	return signals
}

func recordHistory(t *testing.T, service *usecase.Service) {
	t.Helper()
	if err := service.Record(t.Context(), principal(), priorEvidence()); err != nil {
		t.Fatal(err)
	}
}

func TestFullLearningLoopFromFeedback(t *testing.T) {
	repository := newStore(true)
	service := newService(t, repository, newFeedbackStore(), newCache(), user.Preferences{})
	recordHistory(t, service)
	stored, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip())
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Positives) != 3 || stored.Pace != feedbackdomain.TooBusy {
		t.Fatalf("stored = %+v", stored)
	}
	if len(repository.signals) == 0 {
		t.Fatal("feedback produced no signals")
	}
	insights, err := service.Insights(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	found := map[domain.Key]domain.Insight{}
	for _, insight := range insights {
		found[insight.Key] = insight
	}
	for _, expected := range []domain.Key{domain.InterestCoastal, domain.InterestFood, domain.StayQuietArea, domain.TransferTolerance} {
		if _, present := found[expected]; !present {
			t.Fatalf("%q was not learned: %v", expected, insights)
		}
	}
	if found[domain.InterestCoastal].SummaryCode == "" || found[domain.InterestCoastal].EvidenceCount == 0 {
		t.Fatalf("insight = %+v", found[domain.InterestCoastal])
	}
	if found[domain.InterestCoastal].Parameters["explanation"] != string(domain.PostTripFeedback) {
		t.Fatalf("explanation = %v", found[domain.InterestCoastal].Parameters)
	}
}

func TestResubmittingTheSameFeedbackDoesNotDoubleTheEvidence(t *testing.T) {
	repository := newStore(true)
	service := newService(t, repository, newFeedbackStore(), newCache(), user.Preferences{})
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	first := len(repository.signals)
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	if len(repository.signals) != first {
		t.Fatalf("resubmission added %d signals", len(repository.signals)-first)
	}
}

func TestExplicitPreferencesAlwaysReachTheContext(t *testing.T) {
	repository := newStore(true)
	explicit := user.Preferences{Interests: []string{"culture"}, Pace: "packed", Crowds: "lively"}
	service := newService(t, repository, newFeedbackStore(), newCache(), explicit)
	recordHistory(t, service)
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	origin, found := value.OriginOf(domain.InterestCulture)
	if !found || origin != domain.ExplicitProfile {
		t.Fatalf("explicit culture interest = %q %v", origin, found)
	}
	if value.Influence(domain.InterestCulture) != domain.ExplicitWeight {
		t.Fatal("an explicit preference did not carry full weight")
	}
	learnedOrigin, learnedFound := value.OriginOf(domain.InterestCoastal)
	if !learnedFound || learnedOrigin != domain.InferredLearned {
		t.Fatalf("learned coastal interest = %q %v", learnedOrigin, learnedFound)
	}
	if value.Influence(domain.InterestCoastal) >= value.Influence(domain.InterestCulture) {
		t.Fatal("an inference outweighed an explicit preference")
	}
}

func TestDisablingLearningRemovesHistoricalInfluence(t *testing.T) {
	repository := newStore(true)
	contexts := newCache()
	explicit := user.Preferences{Interests: []string{"culture"}}
	service := newService(t, repository, newFeedbackStore(), contexts, explicit)
	recordHistory(t, service)
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	enabled, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if !enabled.Holds(domain.InterestCoastal) {
		t.Fatal("learning was enabled but produced no influence")
	}
	disabled, err := service.SetHistoryLearning(t.Context(), principal(), false)
	if err != nil {
		t.Fatal(err)
	}
	if disabled.HistoryLearningEnabled || len(disabled.Learned) != 0 {
		t.Fatalf("disabled context = %+v", disabled)
	}
	if disabled.Holds(domain.InterestCoastal) {
		t.Fatal("a learned preference survived the toggle")
	}
	if !disabled.Holds(domain.InterestCulture) {
		t.Fatal("an explicit preference was lost when learning was disabled")
	}
	if len(repository.preferences) == 0 {
		t.Fatal("disabling learning deleted stored preferences without being asked")
	}
}

func TestDisabledLearningStopsRecordingNewSignals(t *testing.T) {
	repository := newStore(false)
	service := newService(t, repository, newFeedbackStore(), newCache(), user.Preferences{})
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	if len(repository.signals) != 0 {
		t.Fatalf("signals were recorded while learning was off: %v", repository.signals)
	}
	if len(repository.preferences) != 0 {
		t.Fatal("preferences were learned while learning was off")
	}
}

func TestContextIsCachedAndInvalidatedOnChange(t *testing.T) {
	repository := newStore(true)
	contexts := newCache()
	service := newService(t, repository, newFeedbackStore(), contexts, user.Preferences{})
	if _, err := service.Context(t.Context(), principal()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Context(t.Context(), principal()); err != nil {
		t.Fatal(err)
	}
	if contexts.hits == 0 {
		t.Fatal("the context was never served from cache")
	}
	if _, err := service.SetHistoryLearning(t.Context(), principal(), false); err != nil {
		t.Fatal(err)
	}
	if contexts.invalidated == 0 {
		t.Fatal("changing the toggle did not invalidate the cache")
	}
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if value.HistoryLearningEnabled {
		t.Fatal("a stale context was served after the toggle changed")
	}
}

func TestColdStartProducesAUsableContext(t *testing.T) {
	service := newService(t, newStore(true), newFeedbackStore(), newCache(),
		user.Preferences{Interests: []string{"beach"}})
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Learned) != 0 {
		t.Fatal("a new user already has learned preferences")
	}
	if !value.Holds(domain.InterestCoastal) {
		t.Fatal("explicit setup preferences were not usable at cold start")
	}
	if value.RulesVersion == "" || value.GeneratedAt.IsZero() {
		t.Fatalf("context = %+v", value)
	}
}

func TestFeedbackValidationAndOwnership(t *testing.T) {
	service := newService(t, newStore(true), newFeedbackStore(), newCache(), user.Preferences{})
	if _, err := service.SubmitFeedback(t.Context(), principal(), usecase.FeedbackRequest{TripID: tripID}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("empty feedback was accepted")
	}
	contradictory := usecase.FeedbackRequest{
		TripID:    tripID,
		Positives: []string{string(feedbackdomain.QuietArea)},
		Negatives: []string{string(feedbackdomain.QuietArea)},
	}
	if _, err := service.SubmitFeedback(t.Context(), principal(), contradictory); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("contradictory feedback was accepted")
	}
	invented := usecase.FeedbackRequest{TripID: tripID, Positives: []string{"space_walk"}}
	if _, err := service.SubmitFeedback(t.Context(), principal(), invented); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an invented aspect was accepted")
	}
	if _, err := service.SubmitFeedback(t.Context(), user.Principal{}, coastalFoodTrip()); err == nil {
		t.Fatal("an anonymous caller submitted feedback")
	}
	if _, err := service.Feedback(t.Context(), principal(), tripID); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("missing feedback did not report not found")
	}
}

func TestPassportAndCurrencyNeverBecomeSignals(t *testing.T) {
	repository := newStore(true)
	service := newService(t, repository, newFeedbackStore(), newCache(), user.Preferences{})
	if _, err := service.SubmitFeedback(t.Context(), principal(), coastalFoodTrip()); err != nil {
		t.Fatal(err)
	}
	for _, signal := range repository.signals {
		switch signal.Subject {
		case domain.Interest, domain.ActivityKind, domain.StayArea, domain.TripPace, domain.TravelBurden, domain.Destination:
		default:
			t.Fatalf("an unexpected subject was recorded: %+v", signal)
		}
		for _, forbidden := range []string{"NG", "NGN", "passport", "residence", "visa", "currency"} {
			if signal.SubjectID == forbidden {
				t.Fatalf("a logistics attribute became a preference signal: %+v", signal)
			}
		}
	}
	value, err := service.Context(t.Context(), principal())
	if err != nil {
		t.Fatal(err)
	}
	for _, preference := range append(value.Explicit, value.Learned...) {
		if !preference.Key.Valid() {
			t.Fatalf("an unrecognised key entered the context: %q", preference.Key)
		}
	}
}

func TestAnonymousCallersAreRefusedEverywhere(t *testing.T) {
	service := newService(t, newStore(true), newFeedbackStore(), newCache(), user.Preferences{})
	anonymous := user.Principal{}
	if _, err := service.Context(t.Context(), anonymous); err == nil {
		t.Fatal("an anonymous caller read a personalization context")
	}
	if _, err := service.Insights(t.Context(), anonymous); err == nil {
		t.Fatal("an anonymous caller read insights")
	}
	if _, err := service.SetHistoryLearning(t.Context(), anonymous, false); err == nil {
		t.Fatal("an anonymous caller changed the toggle")
	}
	if err := service.ClearLearning(t.Context(), anonymous); err == nil {
		t.Fatal("an anonymous caller cleared learning")
	}
}
