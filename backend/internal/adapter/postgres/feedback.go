package postgres

import (
	"context"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
)

func (s *Store) SaveTripFeedback(ctx context.Context, submission domain.TripFeedback) (domain.TripFeedback, error) {
	if err := submission.Validate(); err != nil {
		return domain.TripFeedback{}, err
	}
	tripID, err := identifier(submission.TripID)
	if err != nil {
		return domain.TripFeedback{}, err
	}
	userID, err := identifier(submission.UserID)
	if err != nil {
		return domain.TripFeedback{}, err
	}
	row, err := s.queries(ctx).UpsertTripFeedback(ctx, queries.UpsertTripFeedbackParams{
		TripID: tripID, UserID: userID,
		Positives: aspectCodes(submission.Positives), Negatives: aspectCodes(submission.Negatives),
		Pace: string(submission.Pace), TravelBurden: string(submission.TravelBurden),
		ReturnIntent: string(submission.ReturnIntent), SubmittedAt: submission.SubmittedAt.UTC(),
	})
	if err != nil {
		return domain.TripFeedback{}, databaseError(err)
	}
	return toFeedback(row), nil
}

func (s *Store) TripFeedback(ctx context.Context, tripID string) (domain.TripFeedback, error) {
	identity, err := identifier(tripID)
	if err != nil {
		return domain.TripFeedback{}, err
	}
	row, err := s.queries(ctx).GetTripFeedback(ctx, identity)
	if err != nil {
		return domain.TripFeedback{}, databaseError(err)
	}
	return toFeedback(row), nil
}

func aspectCodes(aspects []domain.Aspect) []string {
	codes := make([]string, 0, len(aspects))
	for _, aspect := range aspects {
		codes = append(codes, string(aspect))
	}
	return codes
}

func toAspects(codes []string) []domain.Aspect {
	aspects := make([]domain.Aspect, 0, len(codes))
	for _, code := range codes {
		aspects = append(aspects, domain.Aspect(code))
	}
	return aspects
}

func toFeedback(row queries.TripFeedback) domain.TripFeedback {
	return domain.TripFeedback{
		TripID: uuidText(row.TripID), UserID: uuidText(row.UserID),
		Positives: toAspects(row.Positives), Negatives: toAspects(row.Negatives),
		Pace: domain.PaceAssessment(row.Pace), TravelBurden: domain.BurdenAssessment(row.TravelBurden),
		ReturnIntent: domain.ReturnIntent(row.ReturnIntent),
		SubmittedAt:  row.SubmittedAt.UTC(), UpdatedAt: row.UpdatedAt.UTC(),
	}
}
