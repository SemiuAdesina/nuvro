package personalization

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
)

func (s *Service) Record(ctx context.Context, principal user.Principal, signals []domain.Signal) error {
	if principal.UserID == "" {
		return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	enabled, err := s.store.HistoryLearning(ctx, principal.UserID)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	for _, signal := range signals {
		signal.UserID = principal.UserID
		if err := s.record(ctx, signal); err != nil {
			return err
		}
	}
	return s.relearn(ctx, principal.UserID)
}

func (s *Service) record(ctx context.Context, signal domain.Signal) error {
	if !signal.Valid() {
		return nil
	}
	identifier, err := s.identifiers.Generate()
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
	return s.store.RecordSignal(ctx, identifier, signal)
}

func (s *Service) relearn(ctx context.Context, userID string) error {
	signals, err := s.store.Signals(ctx, userID, s.timer.Now().UTC().Add(-signalWindow))
	if err != nil {
		return err
	}
	existing, err := s.store.LearnedPreferences(ctx, userID)
	if err != nil {
		return err
	}
	resolved := map[domain.Key]domain.Status{}
	for _, preference := range existing {
		resolved[preference.Key] = preference.Status
	}
	for _, learned := range s.engine.Aggregate(signals) {
		if status, known := resolved[learned.Key]; known && !status.Influences() {
			continue
		}
		learned.UserID = userID
		if status, known := resolved[learned.Key]; known {
			learned.Status = status
		}
		identifier, generateErr := s.identifiers.Generate()
		if generateErr != nil {
			return apperror.Wrap(apperror.CodeInternal, "", generateErr)
		}
		if err := s.store.SaveLearnedPreference(ctx, identifier, learned); err != nil {
			return err
		}
	}
	s.invalidate(ctx, userID)
	return nil
}

type FeedbackRequest struct {
	TripID       string
	Positives    []string
	Negatives    []string
	Pace         string
	TravelBurden string
	ReturnIntent string
}

func (s *Service) SubmitFeedback(ctx context.Context, principal user.Principal, request FeedbackRequest) (feedbackdomain.TripFeedback, error) {
	if principal.UserID == "" {
		return feedbackdomain.TripFeedback{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if _, err := s.trips.Read(ctx, principal, request.TripID); err != nil {
		return feedbackdomain.TripFeedback{}, err
	}
	submission := feedbackdomain.TripFeedback{
		TripID: request.TripID, UserID: principal.UserID,
		Positives: aspects(request.Positives), Negatives: aspects(request.Negatives),
		Pace:         feedbackdomain.PaceAssessment(request.Pace),
		TravelBurden: feedbackdomain.BurdenAssessment(request.TravelBurden),
		ReturnIntent: feedbackdomain.ReturnIntent(request.ReturnIntent),
		SubmittedAt:  s.timer.Now().UTC(),
	}
	if submission.Empty() {
		return feedbackdomain.TripFeedback{}, apperror.New(apperror.CodeInvalidInput, "Tell us at least one thing about the trip.")
	}
	if err := submission.Validate(); err != nil {
		return feedbackdomain.TripFeedback{}, apperror.Wrap(apperror.CodeInvalidInput, "That feedback is not valid.", err)
	}
	stored, err := s.feedback.SaveTripFeedback(ctx, submission)
	if err != nil {
		return feedbackdomain.TripFeedback{}, err
	}
	if err := s.Record(ctx, principal, engine.SignalsFromFeedback(stored)); err != nil {
		return feedbackdomain.TripFeedback{}, err
	}
	return stored, nil
}

func (s *Service) Feedback(ctx context.Context, principal user.Principal, tripID string) (feedbackdomain.TripFeedback, error) {
	if _, err := s.trips.Read(ctx, principal, tripID); err != nil {
		return feedbackdomain.TripFeedback{}, err
	}
	stored, err := s.feedback.TripFeedback(ctx, tripID)
	if err != nil {
		return feedbackdomain.TripFeedback{}, apperror.New(apperror.CodeNotFound, "No feedback has been given for this trip.")
	}
	return stored, nil
}

func aspects(values []string) []feedbackdomain.Aspect {
	converted := make([]feedbackdomain.Aspect, 0, len(values))
	for _, value := range values {
		converted = append(converted, feedbackdomain.Aspect(value))
	}
	return converted
}
