package alerts

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
}

type Store interface {
	Alerts(context.Context, string) ([]alert.Alert, error)
	Alert(context.Context, string) (alert.Alert, error)
	SaveAlert(context.Context, alert.Alert) (alert.Alert, error)
	UpdateAlertStatus(context.Context, string, alert.Status) error
	ExpireAlerts(ctx context.Context, tripID string, now time.Time) error
}

type Identifiers interface{ Generate() (string, error) }

type Service struct {
	trips       Trips
	store       Store
	identifiers Identifiers
	timer       clock.Clock
}

func NewService(trips Trips, store Store, identifiers Identifiers, timer clock.Clock) *Service {
	return &Service{trips: trips, store: store, identifiers: identifiers, timer: timer}
}

func (s *Service) List(ctx context.Context, principal user.Principal, tripID string, includeResolved bool) ([]alert.Alert, error) {
	if _, err := s.trips.Read(ctx, principal, tripID); err != nil {
		return nil, err
	}
	now := s.timer.Now().UTC()
	if err := s.store.ExpireAlerts(ctx, tripID, now); err != nil {
		return nil, err
	}
	stored, err := s.store.Alerts(ctx, tripID)
	if err != nil {
		return nil, err
	}
	selected := make([]alert.Alert, 0, len(stored))
	for _, value := range stored {
		if includeResolved || value.Visible(now) {
			selected = append(selected, value)
		}
	}
	return selected, nil
}

func (s *Service) UpdateStatus(ctx context.Context, principal user.Principal, tripID, alertID string, next alert.Status) (alert.Alert, error) {
	if _, err := s.trips.Read(ctx, principal, tripID); err != nil {
		return alert.Alert{}, err
	}
	stored, err := s.store.Alert(ctx, alertID)
	if err != nil || stored.TripID != tripID {
		return alert.Alert{}, apperror.New(apperror.CodeNotFound, "Alert not found.")
	}
	if stored.Status == next {
		return stored, nil
	}
	if !stored.Status.CanTransitionTo(next) {
		return alert.Alert{}, apperror.New(apperror.CodeInvalidState, "That alert status change is not allowed.")
	}
	if err := s.store.UpdateAlertStatus(ctx, alertID, next); err != nil {
		return alert.Alert{}, err
	}
	stored.Status = next
	return stored, nil
}

func (s *Service) Record(ctx context.Context, produced []alert.Alert) ([]alert.Alert, error) {
	saved := make([]alert.Alert, 0, len(produced))
	for _, value := range produced {
		identifier, err := s.identifiers.Generate()
		if err != nil {
			return nil, apperror.Wrap(apperror.CodeInternal, "", err)
		}
		value.ID = identifier
		stored, err := s.store.SaveAlert(ctx, value)
		if err != nil {
			return nil, err
		}
		saved = append(saved, stored)
	}
	return saved, nil
}
