package plan

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type TripReader interface {
	Get(context.Context, string) (trip.Trip, error)
	Role(context.Context, string, string) (trip.Role, error)
}

type Reader struct{ repository TripReader }

func NewReader(repository TripReader) *Reader { return &Reader{repository: repository} }

func (r *Reader) Read(ctx context.Context, principal user.Principal, tripID string) (trip.Trip, error) {
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	role, err := r.repository.Role(ctx, tripID, principal.UserID)
	if err != nil {
		return trip.Trip{}, notFoundInstead(err)
	}
	if !role.CanRead() {
		return trip.Trip{}, apperror.New(apperror.CodeNotFound, "Trip not found.")
	}
	return r.repository.Get(ctx, tripID)
}
