package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
)

func (s *Store) Insert(ctx context.Context, input trip.Trip) error {
	tripValue, err := identifier(input.ID)
	if err != nil {
		return err
	}
	owner, err := identifier(input.OwnerID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).InsertTrip(ctx, queries.InsertTripParams{
		ID: tripValue, OwnerID: owner, DestinationID: input.DestinationID, OriginAirport: input.OriginAirport,
		StartDate: dateValue(input.StartDate), EndDate: dateValue(input.EndDate), Adults: input.Adults, Children: input.Children,
		PreferredCurrency: input.Currency, Status: string(input.Status), CreatedAt: input.CreatedAt.UTC(), UpdatedAt: input.UpdatedAt.UTC(),
	}))
}

func (s *Store) AddOwner(ctx context.Context, tripID, userID string) error {
	tripValue, err := identifier(tripID)
	if err != nil {
		return err
	}
	userValue, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).InsertOwner(ctx, queries.InsertOwnerParams{TripID: tripValue, UserID: userValue}))
}

func (s *Store) Get(ctx context.Context, tripID string) (trip.Trip, error) {
	value, err := identifier(tripID)
	if err != nil {
		return trip.Trip{}, err
	}
	row, err := s.queries(ctx).GetTrip(ctx, value)
	if err != nil {
		return trip.Trip{}, databaseError(err)
	}
	return tripFromRow(row), nil
}

func (s *Store) Save(ctx context.Context, input trip.Trip) error {
	value, err := identifier(input.ID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdateTrip(ctx, queries.UpdateTripParams{
		ID: value, DestinationID: input.DestinationID, OriginAirport: input.OriginAirport,
		StartDate: dateValue(input.StartDate), EndDate: dateValue(input.EndDate), Adults: input.Adults,
		Children: input.Children, PreferredCurrency: input.Currency, Status: string(input.Status),
		UpdatedAt: input.UpdatedAt.UTC(),
	}))
}

func (s *Store) ListForMember(ctx context.Context, userID string, after *cursor.Position, limit int32) ([]trip.Trip, error) {
	value, err := identifier(userID)
	if err != nil {
		return nil, err
	}
	params := queries.ListTripsForMemberParams{UserID: value, PageSize: limit}
	if after != nil {
		instant := after.CreatedAt.UTC()
		position, positionErr := identifier(after.ID)
		if positionErr != nil {
			return nil, positionErr
		}
		params.CursorCreatedAt = &instant
		params.CursorID = position
	}
	rows, err := s.queries(ctx).ListTripsForMember(ctx, params)
	if err != nil {
		return nil, databaseError(err)
	}
	found := make([]trip.Trip, 0, len(rows))
	for _, row := range rows {
		found = append(found, tripFromRow(row))
	}
	return found, nil
}

func tripFromRow(row queries.Trip) trip.Trip {
	return trip.Trip{
		ID: row.ID.String(), OwnerID: row.OwnerID.String(), DestinationID: row.DestinationID,
		OriginAirport: row.OriginAirport, StartDate: datePointer(row.StartDate), EndDate: datePointer(row.EndDate),
		Adults: row.Adults, Children: row.Children, Currency: row.PreferredCurrency, Status: trip.Status(row.Status),
		CreatedAt: row.CreatedAt.UTC(), UpdatedAt: row.UpdatedAt.UTC(),
	}
}

func (s *Store) Role(ctx context.Context, tripID, userID string) (trip.Role, error) {
	tripValue, err := identifier(tripID)
	if err != nil {
		return "", err
	}
	userValue, err := identifier(userID)
	if err != nil {
		return "", err
	}
	role, err := s.queries(ctx).GetMemberRole(ctx, queries.GetMemberRoleParams{TripID: tripValue, UserID: userValue})
	return trip.Role(role), databaseError(err)
}

func (s *Store) TripsNeedingEvaluation(ctx context.Context, from, until time.Time, limit int) ([]trip.Trip, error) {
	rows, err := s.queries(ctx).ListTripsNeedingEvaluation(ctx, queries.ListTripsNeedingEvaluationParams{
		Column1: pgtypeDate(from), Column2: pgtypeDate(until), Limit: bounded(limit),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	trips := make([]trip.Trip, 0, len(rows))
	for _, row := range rows {
		trips = append(trips, tripFromRow(row))
	}
	return trips, nil
}
