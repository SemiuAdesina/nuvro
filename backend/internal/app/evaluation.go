package app

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
	weatherengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type snapshotEvaluator struct {
	store      *postgres.Store
	forecaster *weatherengine.Engine
	timer      clock.Clock
	policy     domain.BufferPolicy
}

func newSnapshotEvaluator(store *postgres.Store, forecaster *weatherengine.Engine, timer clock.Clock) snapshotEvaluator {
	return snapshotEvaluator{store: store, forecaster: forecaster, timer: timer, policy: domain.DefaultBufferPolicy()}
}

func (s snapshotEvaluator) Evaluate(ctx context.Context, tripID string) (alerting.Input, bool) {
	record, err := s.store.Get(ctx, tripID)
	if err != nil || record.DestinationID == nil {
		return alerting.Input{}, false
	}
	place, err := s.store.Destination(ctx, *record.DestinationID)
	if err != nil {
		return alerting.Input{}, false
	}
	plan, err := s.store.Itinerary(ctx, tripID)
	if err != nil || len(plan.Items) == 0 {
		return alerting.Input{}, false
	}
	now := s.timer.Now().UTC()
	next := domain.NextItem(plan, now)
	window := weather.Unavailable(weather.ProviderUnavailable)
	if s.forecaster != nil && record.EndDate != nil {
		window = s.forecaster.For(ctx, geo.Point{Latitude: place.Latitude, Longitude: place.Longitude},
			now, record.EndDate.UTC(), "")
	}
	input := alerting.Input{
		TripID: tripID, Timezone: place.Timezone, Now: now, Itinerary: plan, NextItem: next, Weather: window,
	}
	input.Departure = domain.EvaluateDeparture(now, next,
		route.UnavailableEvidence(geo.Point{}, geo.Point{}, route.Walk), s.policy)
	return input, true
}
