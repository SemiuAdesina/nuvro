package app

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

func newAccountant(timer clock.Clock) *budgetengine.Engine {
	return budgetengine.New(nil, nil, nil, timer)
}

type tripAccess struct {
	reader *plan.Reader
	store  *postgres.Store
}

func (t tripAccess) Read(ctx context.Context, principal user.Principal, tripID string) (trip.Trip, error) {
	return t.reader.Read(ctx, principal, tripID)
}

func (t tripAccess) Role(ctx context.Context, tripID, userID string) (trip.Role, error) {
	return t.store.Role(ctx, tripID, userID)
}

type destinationQueries struct{ store *postgres.Store }

func (d destinationQueries) Query(ctx context.Context, destinationID string) (places.TextQuery, bool) {
	record, err := d.store.Destination(ctx, destinationID)
	if err != nil {
		return places.TextQuery{}, false
	}
	return places.TextQuery{
		Text:         record.Name + " " + record.CountryCode,
		Centre:       point(record),
		RadiusMeters: 20000,
	}, true
}

func point(record destination.Destination) geo.Point {
	return geo.Point{Latitude: record.Latitude, Longitude: record.Longitude}
}
