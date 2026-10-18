package plan

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type TripUpdater interface {
	TripReader
	Save(context.Context, trip.Trip) error
}

type Updater struct {
	repository  TripUpdater
	transaction Transaction
	timer       clock.Clock
}

func NewUpdater(repository TripUpdater, transaction Transaction, timer clock.Clock) *Updater {
	return &Updater{repository: repository, transaction: transaction, timer: timer}
}

func (u *Updater) Update(ctx context.Context, principal user.Principal, tripID string, patch trip.Patch) (trip.Trip, error) {
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	var updated trip.Trip
	err := u.transaction.Within(ctx, func(txctx context.Context) error {
		role, err := u.repository.Role(txctx, tripID, principal.UserID)
		if err != nil {
			return notFoundInstead(err)
		}
		if !role.CanRead() {
			return apperror.New(apperror.CodeNotFound, "Trip not found.")
		}
		if !role.CanEdit() {
			return apperror.New(apperror.CodeForbidden, "This trip is read only for you.")
		}
		current, err := u.repository.Get(txctx, tripID)
		if err != nil {
			return err
		}
		updated, err = current.Apply(patch, u.timer.Now())
		if err != nil {
			return apperror.New(apperror.CodeInvalidInput, "Trip values are invalid.")
		}
		return u.repository.Save(txctx, updated)
	})
	if err != nil {
		return trip.Trip{}, err
	}
	return updated, nil
}

func notFoundInstead(err error) error {
	if apperror.CodeOf(err) == apperror.CodeNotFound {
		return apperror.New(apperror.CodeNotFound, "Trip not found.")
	}
	return err
}
