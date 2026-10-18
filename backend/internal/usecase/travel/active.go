package travel

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Service struct {
	trips       Trips
	catalog     Catalog
	itineraries Itineraries
	alerts      Alerts
	proposals   Proposals
	forecaster  Forecaster
	routes      RouteFinder
	planner     Planner
	preferences Preferences
	identifiers identifierSource
	timer       clock.Clock
	policy      domain.BufferPolicy
}

type identifierSource interface{ Generate() (string, error) }

type Dependencies struct {
	Trips       Trips
	Catalog     Catalog
	Itineraries Itineraries
	Alerts      Alerts
	Proposals   Proposals
	Forecaster  Forecaster
	Routes      RouteFinder
	Planner     Planner
	Preferences Preferences
	Identifiers identifierSource
}

func NewService(dependencies Dependencies, timer clock.Clock) *Service {
	return &Service{
		trips: dependencies.Trips, catalog: dependencies.Catalog, itineraries: dependencies.Itineraries,
		alerts: dependencies.Alerts, proposals: dependencies.Proposals, forecaster: dependencies.Forecaster,
		routes: dependencies.Routes, planner: dependencies.Planner, preferences: dependencies.Preferences,
		identifiers: dependencies.Identifiers, timer: timer, policy: domain.DefaultBufferPolicy(),
	}
}

type ActiveRequest struct {
	TripID   string
	Location *geo.Point
	Language string
}

type ActiveResult struct {
	Snapshot domain.Snapshot
	Alerts   []alert.Alert
}

func (s *Service) Active(ctx context.Context, principal user.Principal, request ActiveRequest) (ActiveResult, error) {
	record, place, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return ActiveResult{}, err
	}
	location := domain.LocationOf(place.Timezone)
	now := s.timer.Now().UTC()
	localNow := now.In(location)
	snapshot := domain.Snapshot{
		Trip: record, Timezone: place.Timezone, LocalNow: localNow,
		Phase: domain.PhaseOf(record, localNow), CurrentDay: domain.DayNumber(record, localNow),
		GeneratedAt: now,
	}
	plan, hasPlan := s.plan(ctx, request.TripID)
	if hasPlan {
		snapshot.Today = domain.ItemsOnDay(plan, snapshot.CurrentDay, location)
		snapshot.NextItem = domain.NextItem(plan, now)
	}
	snapshot.Departure = s.departure(ctx, snapshot.NextItem, request)
	snapshot.Weather, snapshot.HasWeather = s.forecast(ctx, place, record, request.Language)
	if snapshot.HasWeather {
		snapshot.WeatherSection = snapshot.Weather.Source.Freshness
	}
	snapshot.Actions = domain.ActionsFor(snapshot.NextItem, false)
	return ActiveResult{Snapshot: snapshot, Alerts: s.visibleAlerts(ctx, request.TripID, now)}, nil
}

func (s *Service) context(ctx context.Context, principal user.Principal, tripID string) (trip.Trip, destination.Destination, error) {
	record, err := s.trips.Read(ctx, principal, tripID)
	if err != nil {
		return trip.Trip{}, destination.Destination{}, err
	}
	if record.DestinationID == nil {
		return trip.Trip{}, destination.Destination{}, apperror.New(apperror.CodeInvalidState,
			"This trip needs a destination before Travel Mode can run.")
	}
	place, err := s.catalog.Destination(ctx, *record.DestinationID)
	if err != nil {
		return trip.Trip{}, destination.Destination{}, apperror.New(apperror.CodeNotFound, "Destination not found.")
	}
	return record, place, nil
}

func (s *Service) plan(ctx context.Context, tripID string) (itinerary.Itinerary, bool) {
	if s.itineraries == nil {
		return itinerary.Itinerary{}, false
	}
	plan, err := s.itineraries.Itinerary(ctx, tripID)
	if err != nil {
		return itinerary.Itinerary{}, false
	}
	return plan, true
}

func (s *Service) departure(ctx context.Context, next *itinerary.Item, request ActiveRequest) domain.Departure {
	if next == nil {
		return domain.EvaluateDeparture(s.timer.Now(), nil, route.Evidence{}, s.policy)
	}
	evidence := route.UnavailableEvidence(geo.Point{}, next.Location, route.Walk)
	if s.routes != nil && request.Location != nil && !request.Location.IsZero() && !next.Location.IsZero() {
		evidence = s.routes.Compute(ctx, *request.Location, next.Location, route.Walk, next.StartAt)
	}
	return domain.EvaluateDeparture(s.timer.Now(), next, evidence, s.policy)
}

func (s *Service) forecast(ctx context.Context, place destination.Destination, record trip.Trip, language string) (weather.Window, bool) {
	if s.forecaster == nil {
		return weather.Unavailable(weather.ProviderUnavailable), false
	}
	start, end := s.window(record)
	window := s.forecaster.For(ctx, geo.Point{Latitude: place.Latitude, Longitude: place.Longitude}, start, end, language)
	return window, window.Known()
}

func (s *Service) window(record trip.Trip) (time.Time, time.Time) {
	now := s.timer.Now().UTC()
	if record.EndDate == nil {
		return now, now.AddDate(0, 0, 2)
	}
	return now, record.EndDate.UTC()
}

func (s *Service) visibleAlerts(ctx context.Context, tripID string, now time.Time) []alert.Alert {
	if s.alerts == nil {
		return nil
	}
	stored, err := s.alerts.Alerts(ctx, tripID)
	if err != nil {
		return nil
	}
	visible := make([]alert.Alert, 0, len(stored))
	for _, value := range stored {
		if value.Visible(now) {
			visible = append(visible, value)
		}
	}
	return visible
}
