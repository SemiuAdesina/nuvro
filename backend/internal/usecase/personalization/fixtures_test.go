package personalization_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

type trips struct{ err error }

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if t.err != nil {
		return trip.Trip{}, t.err
	}
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return trip.Trip{ID: tripID, OwnerID: owner, Status: trip.Completed}, nil
}

type preferences struct{ values user.Preferences }

func (p preferences) Preferences(context.Context, string) (user.Preferences, error) {
	return p.values, nil
}

type store struct {
	enabled        bool
	signals        []domain.Signal
	preferences    map[string]domain.LearnedPreference
	cleared        int
	signalsCleared int
	err            error
}

func newStore(enabled bool) *store {
	return &store{enabled: enabled, preferences: map[string]domain.LearnedPreference{}}
}

func (s *store) RecordSignal(_ context.Context, signalID string, signal domain.Signal) error {
	if s.err != nil {
		return s.err
	}
	for _, existing := range s.signals {
		if existing.SourceEvent != "" && existing.SourceEvent == signal.SourceEvent {
			return nil
		}
	}
	signal.ID = signalID
	s.signals = append(s.signals, signal)
	return nil
}

func (s *store) Signals(_ context.Context, _ string, since time.Time) ([]domain.Signal, error) {
	kept := make([]domain.Signal, 0, len(s.signals))
	for _, signal := range s.signals {
		if !signal.OccurredAt.Before(since) {
			kept = append(kept, signal)
		}
	}
	return kept, nil
}

func (s *store) ClearSignals(context.Context, string) error {
	s.signals = nil
	s.signalsCleared++
	return nil
}

func (s *store) SaveLearnedPreference(_ context.Context, preferenceID string, preference domain.LearnedPreference) error {
	existing, found := s.preferences[string(preference.Key)]
	if found && !existing.Status.Influences() {
		return nil
	}
	if found {
		preference.ID = existing.ID
		preference.Status = existing.Status
	} else {
		preference.ID = preferenceID
	}
	s.preferences[string(preference.Key)] = preference
	return nil
}

func (s *store) LearnedPreferences(context.Context, string) ([]domain.LearnedPreference, error) {
	values := make([]domain.LearnedPreference, 0, len(s.preferences))
	for _, preference := range s.preferences {
		values = append(values, preference)
	}
	return values, nil
}

func (s *store) LearnedPreference(_ context.Context, preferenceID string) (domain.LearnedPreference, error) {
	for _, preference := range s.preferences {
		if preference.ID == preferenceID {
			return preference, nil
		}
	}
	return domain.LearnedPreference{}, errors.New("not found")
}

func (s *store) UpdateLearnedPreferenceStatus(_ context.Context, preferenceID string, status domain.Status) error {
	for key, preference := range s.preferences {
		if preference.ID == preferenceID {
			preference.Status = status
			s.preferences[key] = preference
			return nil
		}
	}
	return errors.New("not found")
}

func (s *store) ClearLearnedPreferences(context.Context, string) error {
	s.preferences = map[string]domain.LearnedPreference{}
	s.cleared++
	return nil
}

func (s *store) SetHistoryLearning(_ context.Context, _ string, enabled bool) error {
	s.enabled = enabled
	return nil
}

func (s *store) HistoryLearning(context.Context, string) (bool, error) { return s.enabled, nil }

type feedbackStore struct {
	stored map[string]feedbackdomain.TripFeedback
	err    error
}

func newFeedbackStore() *feedbackStore {
	return &feedbackStore{stored: map[string]feedbackdomain.TripFeedback{}}
}

func (f *feedbackStore) SaveTripFeedback(_ context.Context, submission feedbackdomain.TripFeedback) (feedbackdomain.TripFeedback, error) {
	if f.err != nil {
		return feedbackdomain.TripFeedback{}, f.err
	}
	submission.UpdatedAt = submission.SubmittedAt
	f.stored[submission.TripID] = submission
	return submission, nil
}

func (f *feedbackStore) TripFeedback(_ context.Context, tripIdentifier string) (feedbackdomain.TripFeedback, error) {
	stored, found := f.stored[tripIdentifier]
	if !found {
		return feedbackdomain.TripFeedback{}, apperror.New(apperror.CodeNotFound, "No feedback.")
	}
	return stored, nil
}

type cache struct {
	stored      map[string]domain.Context
	invalidated int
	hits        int
}

func newCache() *cache { return &cache{stored: map[string]domain.Context{}} }

func (c *cache) Context(_ context.Context, userID string) (domain.Context, bool) {
	value, found := c.stored[userID]
	if found {
		c.hits++
	}
	return value, found
}

func (c *cache) StoreContext(_ context.Context, userID string, value domain.Context, _ time.Duration) {
	c.stored[userID] = value
}

func (c *cache) Invalidate(_ context.Context, userID string) {
	delete(c.stored, userID)
	c.invalidated++
}

func newService(t *testing.T, repository *store, feedback *feedbackStore, contexts *cache, values user.Preferences) *usecase.Service {
	t.Helper()
	learning, err := engine.New(engine.DefaultConfig(), clock.NewFixed(now))
	if err != nil {
		t.Fatal(err)
	}
	return usecase.NewService(usecase.Dependencies{
		Trips: trips{}, Preferences: preferences{values: values}, Store: repository,
		Feedback: feedback, Cache: contexts, Engine: learning, Identifiers: id.NewUUID(),
	}, clock.NewFixed(now))
}

func principal() user.Principal { return user.Principal{UserID: owner} }
