package personalization

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const ContextCacheTTL = 10 * time.Minute
const signalWindow = 730 * 24 * time.Hour

type Service struct {
	trips       Trips
	preferences Preferences
	store       Store
	feedback    FeedbackStore
	cache       ContextCache
	engine      *engine.Engine
	identifiers Identifiers
	timer       clock.Clock
}

type Dependencies struct {
	Trips       Trips
	Preferences Preferences
	Store       Store
	Feedback    FeedbackStore
	Cache       ContextCache
	Engine      *engine.Engine
	Identifiers Identifiers
}

func NewService(dependencies Dependencies, timer clock.Clock) *Service {
	return &Service{
		trips: dependencies.Trips, preferences: dependencies.Preferences, store: dependencies.Store,
		feedback: dependencies.Feedback, cache: dependencies.Cache, engine: dependencies.Engine,
		identifiers: dependencies.Identifiers, timer: timer,
	}
}

func (s *Service) Context(ctx context.Context, principal user.Principal) (domain.Context, error) {
	if principal.UserID == "" {
		return domain.Context{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if s.cache != nil {
		if cached, found := s.cache.Context(ctx, principal.UserID); found {
			return cached, nil
		}
	}
	value, err := s.build(ctx, principal.UserID)
	if err != nil {
		return domain.Context{}, err
	}
	if s.cache != nil {
		s.cache.StoreContext(ctx, principal.UserID, value, ContextCacheTTL)
	}
	return value, nil
}

func (s *Service) build(ctx context.Context, userID string) (domain.Context, error) {
	enabled, err := s.store.HistoryLearning(ctx, userID)
	if err != nil {
		return domain.Context{}, err
	}
	preferences, err := s.preferences.Preferences(ctx, userID)
	if err != nil {
		return domain.Context{}, err
	}
	learned := []domain.LearnedPreference{}
	if enabled {
		stored, learnedErr := s.store.LearnedPreferences(ctx, userID)
		if learnedErr != nil {
			return domain.Context{}, learnedErr
		}
		learned = stored
	}
	value := s.engine.Context(userID, preferences, learned, enabled)
	return value, nil
}

func (s *Service) Insights(ctx context.Context, principal user.Principal) ([]domain.Insight, error) {
	if principal.UserID == "" {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	stored, err := s.store.LearnedPreferences(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	insights := make([]domain.Insight, 0, len(stored))
	for _, preference := range stored {
		if preference.Status == domain.Suppressed {
			continue
		}
		insights = append(insights, domain.InsightFrom(preference))
	}
	return insights, nil
}

func (s *Service) UpdateInsight(ctx context.Context, principal user.Principal, insightID string, next domain.Status) (domain.Insight, error) {
	if principal.UserID == "" {
		return domain.Insight{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	preference, err := s.store.LearnedPreference(ctx, insightID)
	if err != nil || preference.UserID != principal.UserID {
		return domain.Insight{}, apperror.New(apperror.CodeNotFound, "Insight not found.")
	}
	if preference.Status == next {
		return domain.InsightFrom(preference), nil
	}
	if !preference.Status.CanTransitionTo(next) {
		return domain.Insight{}, apperror.New(apperror.CodeInvalidState, "That change is not allowed for this insight.")
	}
	if err := s.store.UpdateLearnedPreferenceStatus(ctx, insightID, next); err != nil {
		return domain.Insight{}, err
	}
	preference.Status = next
	preference.UpdatedAt = s.timer.Now().UTC()
	s.invalidate(ctx, principal.UserID)
	if next == domain.Confirmed {
		if recordErr := s.record(ctx, confirmationSignal(principal.UserID, preference.Key, s.timer.Now())); recordErr != nil {
			return domain.Insight{}, recordErr
		}
	}
	return domain.InsightFrom(preference), nil
}

func (s *Service) SetHistoryLearning(ctx context.Context, principal user.Principal, enabled bool) (domain.Context, error) {
	if principal.UserID == "" {
		return domain.Context{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if err := s.store.SetHistoryLearning(ctx, principal.UserID, enabled); err != nil {
		return domain.Context{}, err
	}
	s.invalidate(ctx, principal.UserID)
	return s.build(ctx, principal.UserID)
}

func (s *Service) ClearLearning(ctx context.Context, principal user.Principal) error {
	if principal.UserID == "" {
		return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if err := s.store.ClearLearnedPreferences(ctx, principal.UserID); err != nil {
		return err
	}
	if err := s.store.ClearSignals(ctx, principal.UserID); err != nil {
		return err
	}
	s.invalidate(ctx, principal.UserID)
	return nil
}

func (s *Service) invalidate(ctx context.Context, userID string) {
	if s.cache != nil {
		s.cache.Invalidate(ctx, userID)
	}
}

func confirmationSignal(userID string, key domain.Key, now time.Time) domain.Signal {
	signal := domain.Signal{
		UserID: userID, Type: domain.ExplicitConfirmation, Subject: domain.Interest,
		SubjectID: string(key), Strength: domain.Strong, OccurredAt: now.UTC(),
		SourceEvent: "insight_confirmed:" + string(key),
	}
	return signal
}
