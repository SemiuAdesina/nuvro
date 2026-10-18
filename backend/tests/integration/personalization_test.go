//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

func signalFixture(t *testing.T, userID, tripID string, subject personalization.SubjectType, value string) personalization.Signal {
	t.Helper()
	signal := personalization.NewSignal(userID, personalization.DestinationSaved, subject, value,
		time.Now().UTC().Truncate(time.Second))
	signal.TripID = tripID
	signal.SourceEvent = "integration:" + tripID + ":" + value
	return signal
}

func TestSignalsPersistAndDeduplicateBySourceEvent(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	signal := signalFixture(t, userID, tripID, personalization.Interest, "coastal")

	if err := store.RecordSignal(t.Context(), newIdentifier(t), signal); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordSignal(t.Context(), newIdentifier(t), signal); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Signals(t.Context(), userID, signal.OccurredAt.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Fatalf("the same event was recorded %d times", len(stored))
	}
	if stored[0].SubjectID != "coastal" || stored[0].TripID != tripID {
		t.Fatalf("stored = %+v", stored[0])
	}
	if stored[0].Strength != personalization.Weak {
		t.Fatalf("strength = %q, want the weak strength of a save", stored[0].Strength)
	}
}

func TestClearingSignalsLeavesOtherTravellersUntouched(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	otherTripID := seedTrip(t, pool, store)
	otherUserID := ownerOf(t, pool, otherTripID)

	for _, owned := range []struct{ user, trip string }{{userID, tripID}, {otherUserID, otherTripID}} {
		signal := signalFixture(t, owned.user, owned.trip, personalization.Interest, "food")
		if err := store.RecordSignal(t.Context(), newIdentifier(t), signal); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.ClearSignals(t.Context(), userID); err != nil {
		t.Fatal(err)
	}
	since := time.Now().UTC().Add(-time.Hour)
	cleared, err := store.Signals(t.Context(), userID, since)
	if err != nil {
		t.Fatal(err)
	}
	kept, err := store.Signals(t.Context(), otherUserID, since)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared) != 0 || len(kept) != 1 {
		t.Fatalf("cleared = %d, other traveller = %d", len(cleared), len(kept))
	}
}

func learnedFixture(userID string, status personalization.Status) personalization.LearnedPreference {
	now := time.Now().UTC().Truncate(time.Second)
	return personalization.LearnedPreference{
		UserID: userID, Key: personalization.InterestCoastal, Confidence: personalization.MediumConfidence,
		Score: 12, EvidenceCount: 3, Explanation: personalization.RepeatedChoices, Status: status,
		RulesVersion: personalization.RulesVersion, FirstObservedAt: now.Add(-72 * time.Hour),
		LastObservedAt: now, UpdatedAt: now,
	}
}

func TestDismissedPreferencesAreNeverResurrectedByRelearning(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	userID := ownerOf(t, pool, seedTrip(t, pool, store))

	preferenceID := newIdentifier(t)
	if err := store.SaveLearnedPreference(t.Context(), preferenceID, learnedFixture(userID, personalization.Active)); err != nil {
		t.Fatal(err)
	}
	stored, err := store.LearnedPreferences(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].Score != 12 {
		t.Fatalf("stored = %+v", stored)
	}
	if err := store.UpdateLearnedPreferenceStatus(t.Context(), stored[0].ID, personalization.Dismissed); err != nil {
		t.Fatal(err)
	}
	revived := learnedFixture(userID, personalization.Active)
	revived.Score = 40
	if err := store.SaveLearnedPreference(t.Context(), newIdentifier(t), revived); err != nil {
		t.Fatal(err)
	}
	after, err := store.LearnedPreferences(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].Status != personalization.Dismissed || after[0].Score == 40 {
		t.Fatalf("a dismissed preference came back: %+v", after)
	}
}

func TestClearingLearnedPreferencesLeavesTheHistorySwitch(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	userID := ownerOf(t, pool, seedTrip(t, pool, store))

	enabled, err := store.HistoryLearning(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Fatal("learning was off by default")
	}
	if err := store.SetHistoryLearning(t.Context(), userID, false); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveLearnedPreference(t.Context(), newIdentifier(t), learnedFixture(userID, personalization.Active)); err != nil {
		t.Fatal(err)
	}
	if err := store.ClearLearnedPreferences(t.Context(), userID); err != nil {
		t.Fatal(err)
	}
	remaining, err := store.LearnedPreferences(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	stillOff, err := store.HistoryLearning(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 || stillOff {
		t.Fatalf("remaining = %d, learning enabled = %v", len(remaining), stillOff)
	}
}

func TestTripFeedbackIsRevisable(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	now := time.Now().UTC().Truncate(time.Second)

	first := feedback.TripFeedback{
		TripID: tripID, UserID: userID,
		Positives: []feedback.Aspect{feedback.CoastalPlaces}, Negatives: []feedback.Aspect{feedback.BusySchedule},
		Pace: feedback.TooBusy, ReturnIntent: feedback.WouldReturn, SubmittedAt: now,
	}
	if _, err := store.SaveTripFeedback(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	revised := first
	revised.Positives = []feedback.Aspect{feedback.CoastalPlaces, feedback.LocalFood}
	revised.Pace = feedback.JustRight
	revised.SubmittedAt = now.Add(time.Hour)
	if _, err := store.SaveTripFeedback(t.Context(), revised); err != nil {
		t.Fatal(err)
	}
	stored, err := store.TripFeedback(t.Context(), tripID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Positives) != 2 || stored.Pace != feedback.JustRight {
		t.Fatalf("stored = %+v", stored)
	}
	if stored.ReturnIntent != feedback.WouldReturn {
		t.Fatalf("return intent = %q", stored.ReturnIntent)
	}
}

func TestPersonalizationContextCacheExpiresAndInvalidates(t *testing.T) {
	cache := redisadapter.NewPersonalizationCache(newCache(t))
	userID := newIdentifier(t)
	if _, found := cache.Context(t.Context(), userID); found {
		t.Fatal("a cold cache reported a hit")
	}
	value := personalization.EmptyContext(userID, true, time.Now().UTC().Truncate(time.Second))
	value.Learned = []personalization.Preference{{
		Key: personalization.InterestCoastal, Origin: personalization.InferredLearned,
		Confidence: personalization.MediumConfidence, Weight: personalization.InferredMediumWeight,
	}}
	cache.StoreContext(t.Context(), userID, value, time.Minute)

	cached, found := cache.Context(t.Context(), userID)
	if !found || len(cached.Learned) != 1 || cached.Learned[0].Key != personalization.InterestCoastal {
		t.Fatalf("cached = %+v found = %v", cached, found)
	}
	if !cached.HistoryLearningEnabled {
		t.Fatal("the learning switch was lost in the cache")
	}
	keys, err := inspector(t).Keys(t.Context(), "nuvro:v1:personalization:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("the context was not written under the personalization namespace")
	}
	cache.Invalidate(t.Context(), userID)
	if _, stillThere := cache.Context(t.Context(), userID); stillThere {
		t.Fatal("invalidation left the context behind")
	}
}
