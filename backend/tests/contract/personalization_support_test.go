package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	personalizationusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

func (s *store) RecordSignal(_ context.Context, signalID string, signal domain.Signal) error {
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
	return nil
}

func (s *store) SaveLearnedPreference(_ context.Context, preferenceID string, preference domain.LearnedPreference) error {
	if s.learned == nil {
		s.learned = map[string]domain.LearnedPreference{}
	}
	existing, found := s.learned[string(preference.Key)]
	if found && !existing.Status.Influences() {
		return nil
	}
	if found {
		preference.ID = existing.ID
		preference.Status = existing.Status
	} else {
		preference.ID = preferenceID
	}
	s.learned[string(preference.Key)] = preference
	return nil
}

func (s *store) LearnedPreferences(context.Context, string) ([]domain.LearnedPreference, error) {
	values := make([]domain.LearnedPreference, 0, len(s.learned))
	for _, preference := range s.learned {
		values = append(values, preference)
	}
	return values, nil
}

func (s *store) LearnedPreference(_ context.Context, preferenceID string) (domain.LearnedPreference, error) {
	for _, preference := range s.learned {
		if preference.ID == preferenceID {
			return preference, nil
		}
	}
	return domain.LearnedPreference{}, errors.New("not found")
}

func (s *store) UpdateLearnedPreferenceStatus(_ context.Context, preferenceID string, status domain.Status) error {
	for key, preference := range s.learned {
		if preference.ID == preferenceID {
			preference.Status = status
			s.learned[key] = preference
			return nil
		}
	}
	return errors.New("not found")
}

func (s *store) ClearLearnedPreferences(context.Context, string) error {
	s.learned = map[string]domain.LearnedPreference{}
	return nil
}

func (s *store) SetHistoryLearning(_ context.Context, _ string, enabled bool) error {
	s.historyLearning = &enabled
	return nil
}

func (s *store) HistoryLearning(context.Context, string) (bool, error) {
	if s.historyLearning == nil {
		return true, nil
	}
	return *s.historyLearning, nil
}

func (s *store) SaveTripFeedback(_ context.Context, submission feedbackdomain.TripFeedback) (feedbackdomain.TripFeedback, error) {
	if s.feedback == nil {
		s.feedback = map[string]feedbackdomain.TripFeedback{}
	}
	submission.UpdatedAt = submission.SubmittedAt
	s.feedback[submission.TripID] = submission
	return submission, nil
}

func (s *store) TripFeedback(_ context.Context, tripID string) (feedbackdomain.TripFeedback, error) {
	stored, found := s.feedback[tripID]
	if !found {
		return feedbackdomain.TripFeedback{}, errors.New("no feedback")
	}
	return stored, nil
}

func personalizationService(t *testing.T, repository *store, timer clock.Clock) *personalizationusecase.Service {
	t.Helper()
	learning, err := engine.New(engine.DefaultConfig(), timer)
	if err != nil {
		t.Fatal(err)
	}
	return personalizationusecase.NewService(personalizationusecase.Dependencies{
		Trips: tripAccess{store: repository}, Preferences: repository, Store: repository,
		Feedback: repository, Engine: learning, Identifiers: id.NewUUID(),
	}, timer)
}
