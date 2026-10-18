package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	commercedomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
)

const watchPageSize = 100

func (s *Store) SavePriceWatch(ctx context.Context, watch domain.Watch) error {
	if err := watch.Validate(); err != nil {
		return err
	}
	watchID, err := identifier(watch.ID)
	if err != nil {
		return err
	}
	userID, err := identifier(watch.UserID)
	if err != nil {
		return err
	}
	params := queries.InsertPriceWatchParams{
		ID: watchID, UserID: userID, ProductType: string(watch.Product),
		Origin: watch.Target.Origin, Destination: watch.Target.Destination,
		PropertyID: watch.Target.PropertyID, PlaceID: watch.Target.PlaceID,
		StartDate: pgtypeDate(watch.Target.StartDate), EndDate: pgtypeDate(watch.Target.EndDate),
		Adults: bounded(watch.Target.Adults), Children: bounded(watch.Target.Children),
		Cabin: watch.Target.Cabin, BaselineMinor: watch.Baseline.Minor(),
		BaselineCurrency:       watch.Baseline.Currency().String(),
		ThresholdPercentage:    bounded(watch.Threshold.PercentageDrop),
		ThresholdAbsoluteMinor: watch.Threshold.AbsoluteMinor,
		Status:                 string(watch.Status), ExpiresAt: watch.ExpiresAt.UTC(),
	}
	if watch.TripID != "" {
		tripID, tripErr := identifier(watch.TripID)
		if tripErr != nil {
			return tripErr
		}
		params.TripID = tripID
	}
	if !watch.NextCheckAt.IsZero() {
		next := watch.NextCheckAt.UTC()
		params.NextCheckAt = &next
	}
	return databaseError(s.queries(ctx).InsertPriceWatch(ctx, params))
}

func (s *Store) PriceWatches(ctx context.Context, userID string) ([]domain.Watch, error) {
	identity, err := identifier(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries(ctx).ListPriceWatches(ctx, queries.ListPriceWatchesParams{
		UserID: identity, Limit: watchPageSize,
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return toWatches(rows)
}

func (s *Store) PriceWatch(ctx context.Context, watchID string) (domain.Watch, error) {
	identity, err := identifier(watchID)
	if err != nil {
		return domain.Watch{}, err
	}
	row, err := s.queries(ctx).GetPriceWatch(ctx, identity)
	if err != nil {
		return domain.Watch{}, databaseError(err)
	}
	return toWatch(row)
}

func (s *Store) ActivePriceWatchCount(ctx context.Context, userID string) (int64, error) {
	identity, err := identifier(userID)
	if err != nil {
		return 0, err
	}
	count, err := s.queries(ctx).CountActivePriceWatches(ctx, identity)
	return count, databaseError(err)
}

func (s *Store) EndPriceWatch(ctx context.Context, watchID string, status domain.Status) error {
	identity, err := identifier(watchID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).EndPriceWatch(ctx, queries.EndPriceWatchParams{
		ID: identity, Status: string(status),
	}))
}

func (s *Store) DuePriceWatches(ctx context.Context, now time.Time, limit int) ([]domain.Watch, error) {
	rows, err := s.queries(ctx).ListDuePriceWatches(ctx, queries.ListDuePriceWatchesParams{
		ExpiresAt: now.UTC(), Limit: bounded(limit),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return toWatches(rows)
}

func (s *Store) RecordObservation(ctx context.Context, observationID string, observation domain.Observation) error {
	identity, err := identifier(observationID)
	if err != nil {
		return err
	}
	watchID, err := identifier(observation.WatchID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).RecordPriceObservation(ctx, queries.RecordPriceObservationParams{
		ID: identity, WatchID: watchID, PriceMinor: observation.Price.Minor(),
		PriceCurrency: observation.Price.Currency().String(), Available: observation.Available,
		ObservedAt: observation.ObservedAt.UTC(),
	}))
}

func (s *Store) ReschedulePriceWatch(ctx context.Context, watchID string, checkedAt, nextCheckAt time.Time, baselineMinor int64) error {
	identity, err := identifier(watchID)
	if err != nil {
		return err
	}
	checked := checkedAt.UTC()
	next := nextCheckAt.UTC()
	return databaseError(s.queries(ctx).ReschedulePriceWatch(ctx, queries.ReschedulePriceWatchParams{
		ID: identity, LastCheckedAt: &checked, NextCheckAt: &next, BaselineMinor: baselineMinor,
	}))
}

func (s *Store) ExpirePriceWatches(ctx context.Context, now time.Time) error {
	return databaseError(s.queries(ctx).ExpirePriceWatches(ctx, now.UTC()))
}

func toWatches(rows []queries.PriceWatch) ([]domain.Watch, error) {
	watches := make([]domain.Watch, 0, len(rows))
	for _, row := range rows {
		watch, err := toWatch(row)
		if err != nil {
			continue
		}
		watches = append(watches, watch)
	}
	return watches, nil
}

func toWatch(row queries.PriceWatch) (domain.Watch, error) {
	baseline, err := money.New(row.BaselineMinor, money.Currency(row.BaselineCurrency))
	if err != nil {
		return domain.Watch{}, err
	}
	watch := domain.Watch{
		ID: uuidText(row.ID), UserID: uuidText(row.UserID),
		Product: commercedomain.ProductType(row.ProductType),
		Target: domain.Target{
			Origin: row.Origin, Destination: row.Destination, PropertyID: row.PropertyID,
			PlaceID: row.PlaceID, StartDate: row.StartDate.Time.UTC(), EndDate: row.EndDate.Time.UTC(),
			Adults: int(row.Adults), Children: int(row.Children), Cabin: row.Cabin,
		},
		Baseline: baseline,
		Threshold: domain.Threshold{
			PercentageDrop: int(row.ThresholdPercentage), AbsoluteMinor: row.ThresholdAbsoluteMinor,
		},
		Status: domain.Status(row.Status), CreatedAt: row.CreatedAt.UTC(), ExpiresAt: row.ExpiresAt.UTC(),
	}
	if row.TripID.Valid {
		watch.TripID = uuidText(row.TripID)
	}
	if row.LastCheckedAt != nil {
		watch.LastCheckedAt = row.LastCheckedAt.UTC()
	}
	if row.NextCheckAt != nil {
		watch.NextCheckAt = row.NextCheckAt.UTC()
	}
	return watch, nil
}
