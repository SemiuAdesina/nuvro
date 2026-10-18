package personalization

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
}

type Preferences interface {
	Preferences(context.Context, string) (user.Preferences, error)
}

type Store interface {
	RecordSignal(ctx context.Context, signalID string, signal domain.Signal) error
	Signals(ctx context.Context, userID string, since time.Time) ([]domain.Signal, error)
	ClearSignals(context.Context, string) error
	SaveLearnedPreference(ctx context.Context, preferenceID string, preference domain.LearnedPreference) error
	LearnedPreferences(context.Context, string) ([]domain.LearnedPreference, error)
	LearnedPreference(context.Context, string) (domain.LearnedPreference, error)
	UpdateLearnedPreferenceStatus(context.Context, string, domain.Status) error
	ClearLearnedPreferences(context.Context, string) error
	SetHistoryLearning(context.Context, string, bool) error
	HistoryLearning(context.Context, string) (bool, error)
}

type FeedbackStore interface {
	SaveTripFeedback(context.Context, feedback.TripFeedback) (feedback.TripFeedback, error)
	TripFeedback(context.Context, string) (feedback.TripFeedback, error)
}

type ContextCache interface {
	Context(ctx context.Context, userID string) (domain.Context, bool)
	StoreContext(ctx context.Context, userID string, value domain.Context, ttl time.Duration)
	Invalidate(ctx context.Context, userID string)
}

type Identifiers interface{ Generate() (string, error) }
