package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
)

func (s *Store) Purge(ctx context.Context, deletion retention.Deletion, now time.Time) (retention.Deletion, error) {
	deletionID, err := identifier(deletion.ID)
	if err != nil {
		return retention.Deletion{}, err
	}
	userID, err := identifier(deletion.UserID)
	if err != nil {
		return retention.Deletion{}, err
	}
	var completed retention.Deletion
	err = s.Within(ctx, func(txCtx context.Context) error {
		store := s.queries(txCtx)
		if memberErr := store.DeleteTripMemberships(txCtx, userID); memberErr != nil {
			return databaseError(memberErr)
		}
		transferred, deleted, tripErr := s.resolveTrips(txCtx, userID)
		if tripErr != nil {
			return tripErr
		}
		if userErr := store.DeleteUser(txCtx, userID); userErr != nil {
			return databaseError(userErr)
		}
		completed = deletion.Complete(now, transferred, deleted)
		return databaseError(store.CompleteAccountDeletion(txCtx, queries.CompleteAccountDeletionParams{
			ID: deletionID, PurgedAt: completed.PurgedAt,
			TripsTransferred: bounded(transferred), TripsDeleted: bounded(deleted),
		}))
	})
	if err != nil {
		return retention.Deletion{}, err
	}
	return completed, nil
}

func (s *Store) resolveTrips(ctx context.Context, userID pgtypeUUID) (int, int, error) {
	store := s.queries(ctx)
	sole, err := store.ListSoleOwnedTrips(ctx, userID)
	if err != nil {
		return 0, 0, databaseError(err)
	}
	for _, tripID := range sole {
		if deleteErr := store.DeleteTrip(ctx, tripID); deleteErr != nil {
			return 0, 0, databaseError(deleteErr)
		}
	}
	successions, err := s.successions(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	now := time.Now().UTC()
	for _, succession := range successions {
		if err := s.transfer(ctx, succession, now); err != nil {
			return 0, 0, err
		}
	}
	return len(successions), len(sole), nil
}

func (s *Store) successions(ctx context.Context, userID pgtypeUUID) ([]retention.Succession, error) {
	rows, err := s.queries(ctx).ListSharedOwnedTrips(ctx, userID)
	if err != nil {
		return nil, databaseError(err)
	}
	members := map[string][]retention.Membership{}
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		tripID := uuidText(row.ID)
		if _, seen := members[tripID]; !seen {
			order = append(order, tripID)
		}
		members[tripID] = append(members[tripID], retention.Membership{
			UserID: uuidText(row.UserID), JoinedAt: row.CreatedAt.UTC(),
		})
	}
	chosen := make([]retention.Succession, 0, len(order))
	for _, tripID := range order {
		if succession, found := retention.Succeed(tripID, members[tripID]); found {
			chosen = append(chosen, succession)
		}
	}
	return chosen, nil
}

func (s *Store) transfer(ctx context.Context, succession retention.Succession, now time.Time) error {
	tripID, err := identifier(succession.TripID)
	if err != nil {
		return err
	}
	successor, err := identifier(succession.Successor)
	if err != nil {
		return err
	}
	store := s.queries(ctx)
	if err := store.TransferTripOwnership(ctx, queries.TransferTripOwnershipParams{
		ID: tripID, OwnerID: successor, UpdatedAt: now,
	}); err != nil {
		return databaseError(err)
	}
	return databaseError(store.PromoteTripMember(ctx, queries.PromoteTripMemberParams{
		TripID: tripID, UserID: successor,
	}))
}

func (s *Store) Sweep(ctx context.Context, rule retention.Rule, now time.Time) (int64, error) {
	cutoff := rule.CutoffFrom(now)
	store := s.queries(ctx)
	sweeps := map[retention.Class]func(context.Context, time.Time) (int64, error){
		retention.PersonalizationSignals: store.DeleteExpiredPersonalizationSignals,
		retention.CommerceHandoffs:       store.DeleteExpiredCommerceHandoffs,
		retention.ProviderWebhookEvents:  store.DeleteExpiredWebhookEvents,
		retention.TripAlerts:             store.DeleteExpiredTripAlerts,
		retention.ReplanProposals:        store.DeleteExpiredReplanProposals,
		retention.ExpiredPriceWatches:    store.DeleteExpiredPriceWatches,
	}
	sweep, known := sweeps[rule.Class]
	if !known {
		return 0, nil
	}
	removed, err := sweep(ctx, cutoff)
	if err != nil {
		return 0, databaseError(err)
	}
	return removed, nil
}
