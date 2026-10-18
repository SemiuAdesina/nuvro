package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
)

func (s *Store) SaveItinerary(ctx context.Context, plan itinerary.Itinerary, references []place.Place) (itinerary.Itinerary, error) {
	tripID, err := identifier(plan.TripID)
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	planID, planErr := identifier(plan.ID)
	if planErr != nil {
		return itinerary.Itinerary{}, planErr
	}
	store := s.queries(ctx)
	if referenceErr := s.saveReferences(ctx, references); referenceErr != nil {
		return itinerary.Itinerary{}, referenceErr
	}
	saved, err := store.UpsertItinerary(ctx, queries.UpsertItineraryParams{
		ID: planID, TripID: tripID, Timezone: plan.Timezone,
		AlgorithmVersion: plan.AlgorithmVersion, GeneratedAt: plan.GeneratedAt.UTC(),
	})
	if err != nil {
		return itinerary.Itinerary{}, databaseError(err)
	}
	if err := store.DeleteUnlockedItems(ctx, saved.ID); err != nil {
		return itinerary.Itinerary{}, databaseError(err)
	}
	if err := s.saveItems(ctx, saved.ID, tripID, plan.Items); err != nil {
		return itinerary.Itinerary{}, err
	}
	plan.ID = uuidText(saved.ID)
	plan.Version = int(saved.Version)
	return plan, nil
}

func (s *Store) saveReferences(ctx context.Context, references []place.Place) error {
	store := s.queries(ctx)
	for _, reference := range references {
		provider, found := firstReference(reference)
		if !found {
			continue
		}
		err := store.UpsertPlaceReference(ctx, queries.UpsertPlaceReferenceParams{
			ID: reference.ID, Provider: provider.Provider, ProviderPlaceID: provider.ExternalID,
			Name: reference.Name, Category: string(reference.Category),
			Latitude: reference.Location.Latitude, Longitude: reference.Location.Longitude,
			Address: reference.Address, Timezone: reference.Timezone,
			ObservedAt: observedAt(reference),
		})
		if err != nil {
			return databaseError(err)
		}
	}
	return nil
}

func (s *Store) saveItems(ctx context.Context, planID, tripID pgtypeUUID, items []itinerary.Item) error {
	store := s.queries(ctx)
	for _, item := range items {
		itemID, err := identifier(item.ID)
		if err != nil {
			return err
		}
		if err := store.InsertItineraryItem(ctx, itemParams(itemID, planID, tripID, item)); err != nil {
			return databaseError(err)
		}
	}
	return nil
}

func (s *Store) Itinerary(ctx context.Context, tripID string) (itinerary.Itinerary, error) {
	identity, err := identifier(tripID)
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	store := s.queries(ctx)
	record, err := store.GetItineraryByTrip(ctx, identity)
	if err != nil {
		return itinerary.Itinerary{}, databaseError(err)
	}
	rows, err := store.ListItineraryItems(ctx, record.ID)
	if err != nil {
		return itinerary.Itinerary{}, databaseError(err)
	}
	items := make([]itinerary.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, toItem(row))
	}
	return itinerary.Itinerary{
		ID: uuidText(record.ID), TripID: uuidText(record.TripID), Timezone: record.Timezone,
		Version: int(record.Version), AlgorithmVersion: record.AlgorithmVersion,
		GeneratedAt: record.GeneratedAt.UTC(), Items: items,
	}, nil
}

func (s *Store) ItineraryItem(ctx context.Context, itemID string) (itinerary.Item, error) {
	identity, err := identifier(itemID)
	if err != nil {
		return itinerary.Item{}, err
	}
	row, err := s.queries(ctx).GetItineraryItem(ctx, identity)
	if err != nil {
		return itinerary.Item{}, databaseError(err)
	}
	return toItem(queries.ListItineraryItemsRow(row)), nil
}

func (s *Store) UpdateItineraryItem(ctx context.Context, item itinerary.Item) error {
	identity, err := identifier(item.ID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdateItineraryItem(ctx, queries.UpdateItineraryItemParams{
		ID: identity, StartsAt: item.StartAt.UTC(), EndsAt: item.EndAt.UTC(),
		Title: item.Title, Status: string(item.Status), UserLocked: item.UserLocked,
	}))
}

func firstReference(value place.Place) (place.Reference, bool) {
	if len(value.References) == 0 {
		return place.Reference{}, false
	}
	return value.References[0], true
}

func observedAt(value place.Place) time.Time {
	if value.Freshness.ObservedAt.IsZero() {
		return time.Now().UTC()
	}
	return value.Freshness.ObservedAt.UTC()
}

func (s *Store) ReplaceItineraryItems(ctx context.Context, tripID string, expectedVersion int, items []itinerary.Item) (itinerary.Itinerary, error) {
	identity, err := identifier(tripID)
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	var saved itinerary.Itinerary
	err = s.Within(ctx, func(txCtx context.Context) error {
		store := s.queries(txCtx)
		record, bumpErr := store.BumpItineraryVersion(txCtx, queries.BumpItineraryVersionParams{
			TripID: identity, Version: bounded(expectedVersion),
		})
		if bumpErr != nil {
			return apperror.New(apperror.CodeConflict, "The plan changed while this update was in flight.")
		}
		if deleteErr := store.DeleteAllItems(txCtx, record.ID); deleteErr != nil {
			return databaseError(deleteErr)
		}
		if insertErr := s.saveItems(txCtx, record.ID, identity, items); insertErr != nil {
			return insertErr
		}
		saved = itinerary.Itinerary{
			ID: uuidText(record.ID), TripID: tripID, Timezone: record.Timezone,
			Version: int(record.Version), AlgorithmVersion: record.AlgorithmVersion,
			GeneratedAt: record.GeneratedAt.UTC(), Items: items,
		}
		return nil
	})
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	return saved, nil
}
