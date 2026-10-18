package plan

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

type TripWriter interface {
	Insert(context.Context, trip.Trip) error
	AddOwner(context.Context, string, string) error
}

type Transaction interface {
	Within(context.Context, func(context.Context) error) error
}

type Creator struct {
	writer      TripWriter
	transaction Transaction
	identifiers id.Generator
	timer       clock.Clock
}

func NewCreator(writer TripWriter, transaction Transaction, identifiers id.Generator, timer clock.Clock) *Creator {
	return &Creator{writer: writer, transaction: transaction, identifiers: identifiers, timer: timer}
}

func (c *Creator) Create(ctx context.Context, principal user.Principal, input trip.Trip) (trip.Trip, error) {
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	identifier, err := c.identifiers.Generate()
	if err != nil {
		return trip.Trip{}, apperror.Wrap(apperror.CodeInternal, "", err)
	}
	input.ID = identifier
	input.OwnerID = principal.UserID
	input.Status = trip.Draft
	input.CreatedAt = c.timer.Now().UTC()
	input.UpdatedAt = input.CreatedAt
	if validationErr := input.Validate(); validationErr != nil {
		return trip.Trip{}, apperror.New(apperror.CodeInvalidInput, "Trip values are invalid.")
	}
	err = c.transaction.Within(ctx, func(txctx context.Context) error {
		if insertErr := c.writer.Insert(txctx, input); insertErr != nil {
			return insertErr
		}
		return c.writer.AddOwner(txctx, input.ID, principal.UserID)
	})
	if err != nil {
		return trip.Trip{}, err
	}
	return input, nil
}
