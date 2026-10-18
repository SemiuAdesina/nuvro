package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
)

func (s *Store) RequestDeletion(ctx context.Context, deletion retention.Deletion) (retention.Deletion, error) {
	deletionID, err := identifier(deletion.ID)
	if err != nil {
		return retention.Deletion{}, err
	}
	userID, err := identifier(deletion.UserID)
	if err != nil {
		return retention.Deletion{}, err
	}
	var saved retention.Deletion
	err = s.Within(ctx, func(txCtx context.Context) error {
		store := s.queries(txCtx)
		record, insertErr := store.InsertAccountDeletion(txCtx, queries.InsertAccountDeletionParams{
			ID: deletionID, UserID: userID, Reason: deletion.Reason,
			RequestedAt: deletion.RequestedAt.UTC(), PurgeAfter: deletion.PurgeAfter.UTC(),
		})
		if insertErr != nil {
			return databaseError(insertErr)
		}
		if markErr := store.MarkUserDeleted(txCtx, queries.MarkUserDeletedParams{
			ID: userID, UpdatedAt: deletion.RequestedAt.UTC(),
		}); markErr != nil {
			return databaseError(markErr)
		}
		saved = toDeletion(record)
		return nil
	})
	if err != nil {
		return retention.Deletion{}, err
	}
	return saved, nil
}

func (s *Store) Deletion(ctx context.Context, userID string) (retention.Deletion, error) {
	identity, err := identifier(userID)
	if err != nil {
		return retention.Deletion{}, err
	}
	record, err := s.queries(ctx).GetAccountDeletion(ctx, identity)
	if err != nil {
		return retention.Deletion{}, databaseError(err)
	}
	return toDeletion(record), nil
}

func (s *Store) DueDeletions(ctx context.Context, now time.Time, limit int) ([]retention.Deletion, error) {
	records, err := s.queries(ctx).ListDueAccountDeletions(ctx, queries.ListDueAccountDeletionsParams{
		PurgeAfter: now.UTC(), Limit: bounded(limit),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	due := make([]retention.Deletion, 0, len(records))
	for _, record := range records {
		due = append(due, toDeletion(record))
	}
	return due, nil
}

func toDeletion(record queries.AccountDeletion) retention.Deletion {
	return retention.Deletion{
		ID: uuidText(record.ID), UserID: uuidText(record.UserID),
		Status: retention.DeletionStatus(record.Status), Reason: record.Reason,
		RequestedAt: record.RequestedAt.UTC(), PurgeAfter: record.PurgeAfter.UTC(),
		PurgedAt: record.PurgedAt, TripsTransferred: int(record.TripsTransferred),
		TripsDeleted: int(record.TripsDeleted),
	}
}
