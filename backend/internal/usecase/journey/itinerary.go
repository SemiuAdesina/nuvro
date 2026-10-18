package journey

import (
	"context"
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

type Service struct {
	trips       Trips
	catalog     Catalog
	store       Store
	preferences Preferences
	planner     Planner
	forecaster  Forecaster
	accountant  Accountant
	finder      Finder
	identifiers id.Generator
	timer       clock.Clock
}

type Dependencies struct {
	Trips       Trips
	Catalog     Catalog
	Store       Store
	Preferences Preferences
	Planner     Planner
	Forecaster  Forecaster
	Accountant  Accountant
	Finder      Finder
	Identifiers id.Generator
}

func NewService(dependencies Dependencies, timer clock.Clock) *Service {
	return &Service{
		trips: dependencies.Trips, catalog: dependencies.Catalog, store: dependencies.Store,
		preferences: dependencies.Preferences, planner: dependencies.Planner,
		forecaster: dependencies.Forecaster, accountant: dependencies.Accountant,
		finder: dependencies.Finder, identifiers: dependencies.Identifiers, timer: timer,
	}
}

type GenerateRequest struct {
	TripID   string
	Pace     string
	Language string
}

type GenerateResult struct {
	Itinerary  itinerary.Itinerary
	Violations []itinerary.Violation
	Candidates int
}

func (s *Service) Generate(ctx context.Context, principal user.Principal, request GenerateRequest) (GenerateResult, error) {
	context, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return GenerateResult{}, err
	}
	if !context.Role.CanEdit() {
		return GenerateResult{}, apperror.New(apperror.CodeForbidden, "Editing this trip is not permitted.")
	}
	existing, hasExisting := s.existing(ctx, request.TripID)
	locked := []itinerary.Item{}
	if hasExisting {
		locked = existing.LockedItems()
	}
	point := geo.Point{Latitude: context.Destination.Latitude, Longitude: context.Destination.Longitude}
	window := s.forecast(ctx, point, context, request.Language)
	result, err := s.planner.Generate(ctx, itineraryengine.Request{
		TripID: request.TripID, Destination: point, Timezone: context.Destination.Timezone,
		Start: *context.Trip.StartDate, End: *context.Trip.EndDate,
		Pace: s.pace(ctx, principal, request.Pace), Interests: s.interests(ctx, principal),
		Language: request.Language, Currency: context.Trip.Currency, Weather: window,
		Locked: locked, Now: s.timer.Now(),
	})
	if err != nil {
		return GenerateResult{}, apperror.Wrap(apperror.CodeUnavailable, "Itinerary evidence is unavailable.", err)
	}
	saved, err := s.persist(ctx, result, existing, hasExisting)
	if err != nil {
		return GenerateResult{}, err
	}
	return GenerateResult{Itinerary: saved, Violations: result.Violations, Candidates: result.Candidates}, nil
}

func (s *Service) persist(ctx context.Context, result itineraryengine.Result, existing itinerary.Itinerary, hasExisting bool) (itinerary.Itinerary, error) {
	plan := result.Itinerary
	if hasExisting {
		plan.ID = existing.ID
	} else {
		identifier, err := s.identifiers.Generate()
		if err != nil {
			return itinerary.Itinerary{}, apperror.Wrap(apperror.CodeInternal, "", err)
		}
		plan.ID = identifier
	}
	return s.store.SaveItinerary(ctx, plan, nil)
}

func (s *Service) Read(ctx context.Context, principal user.Principal, tripID string) (itinerary.Itinerary, error) {
	if _, err := s.trips.Read(ctx, principal, tripID); err != nil {
		return itinerary.Itinerary{}, err
	}
	plan, err := s.store.Itinerary(ctx, tripID)
	if err != nil {
		return itinerary.Itinerary{}, apperror.New(apperror.CodeNotFound, "No itinerary has been generated for this trip.")
	}
	return plan, nil
}

type ItemUpdate struct {
	TripID     string
	ItemID     string
	StartAt    *time.Time
	EndAt      *time.Time
	Title      *string
	Status     *string
	UserLocked *bool
}

func (s *Service) UpdateItem(ctx context.Context, principal user.Principal, update ItemUpdate) (itinerary.Item, error) {
	context, err := s.context(ctx, principal, update.TripID)
	if err != nil {
		return itinerary.Item{}, err
	}
	if !context.Role.CanEdit() {
		return itinerary.Item{}, apperror.New(apperror.CodeForbidden, "Editing this trip is not permitted.")
	}
	item, err := s.store.ItineraryItem(ctx, update.ItemID)
	if err != nil || item.TripID != update.TripID {
		return itinerary.Item{}, apperror.New(apperror.CodeNotFound, "Itinerary item not found.")
	}
	updated, err := apply(item, update)
	if err != nil {
		return itinerary.Item{}, err
	}
	if err := s.store.UpdateItineraryItem(ctx, updated); err != nil {
		return itinerary.Item{}, err
	}
	return updated, nil
}

func apply(item itinerary.Item, update ItemUpdate) (itinerary.Item, error) {
	if update.StartAt != nil {
		item.StartAt = update.StartAt.UTC()
	}
	if update.EndAt != nil {
		item.EndAt = update.EndAt.UTC()
	}
	if update.Title != nil {
		item.Title = *update.Title
	}
	if update.UserLocked != nil {
		item.UserLocked = *update.UserLocked
	}
	if update.Status != nil {
		next := itinerary.Status(*update.Status)
		if !item.Status.CanTransitionTo(next) {
			return itinerary.Item{}, apperror.New(apperror.CodeInvalidState, "That itinerary status change is not allowed.")
		}
		item.Status = next
	}
	if err := item.Validate(); err != nil {
		return itinerary.Item{}, apperror.New(apperror.CodeInvalidInput, "The itinerary item is not valid.")
	}
	return item, nil
}

func (s *Service) existing(ctx context.Context, tripID string) (itinerary.Itinerary, bool) {
	plan, err := s.store.Itinerary(ctx, tripID)
	if err != nil {
		return itinerary.Itinerary{}, false
	}
	return plan, true
}

func (s *Service) pace(ctx context.Context, principal user.Principal, requested string) itineraryengine.Pace {
	if requested != "" {
		return itineraryengine.ParsePace(requested)
	}
	if s.preferences == nil {
		return itineraryengine.Balanced
	}
	preferences, err := s.preferences.Preferences(ctx, principal.UserID)
	if err != nil {
		return itineraryengine.Balanced
	}
	return itineraryengine.ParsePace(preferences.Pace)
}

func (s *Service) interests(ctx context.Context, principal user.Principal) []string {
	if s.preferences == nil {
		return nil
	}
	preferences, err := s.preferences.Preferences(ctx, principal.UserID)
	if err != nil {
		return nil
	}
	return preferences.Interests
}

var errNoDestination = errors.New("journey: trip has no destination")

type tripContext struct {
	Trip        trip.Trip
	Role        trip.Role
	Destination destinationRecord
}
