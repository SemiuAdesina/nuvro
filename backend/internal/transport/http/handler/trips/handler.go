package trips

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

type Creator interface {
	Create(context.Context, user.Principal, trip.Trip) (trip.Trip, error)
}

type Reader interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
}

type Lister interface {
	List(context.Context, user.Principal, string, int) (plan.Page, error)
}

type Updater interface {
	Update(context.Context, user.Principal, string, trip.Patch) (trip.Trip, error)
}

type Handler struct {
	creator Creator
	reader  Reader
	lister  Lister
	updater Updater
}

func NewHandler(creator Creator, reader Reader, lister Lister, updater Updater) *Handler {
	return &Handler{creator: creator, reader: reader, lister: lister, updater: updater}
}

func (h *Handler) CreateTrip(ctx context.Context, request api.CreateTripRequestObject) (api.CreateTripResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	input, ok := creationInput(*request.Body)
	if !ok {
		return nil, apperror.New(apperror.CodeInvalidInput, "Trip values are invalid.")
	}
	created, err := h.creator.Create(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.CreateTrip201JSONResponse(tripResponse(created)), nil
}

func (h *Handler) ListTrips(ctx context.Context, request api.ListTripsRequestObject) (api.ListTripsResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	var token string
	if request.Params.Cursor != nil {
		token = *request.Params.Cursor
	}
	limit := 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	page, err := h.lister.List(ctx, principal, token, limit)
	if err != nil {
		return nil, err
	}
	return api.ListTrips200JSONResponse(pageResponse(page)), nil
}

func (h *Handler) GetTrip(ctx context.Context, request api.GetTripRequestObject) (api.GetTripResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	found, err := h.reader.Read(ctx, principal, request.TripId.String())
	if err != nil {
		return nil, err
	}
	return api.GetTrip200JSONResponse(tripResponse(found)), nil
}

func (h *Handler) PatchTrip(ctx context.Context, request api.PatchTripRequestObject) (api.PatchTripResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	patch, ok := patchInput(*request.Body)
	if !ok {
		return nil, apperror.New(apperror.CodeInvalidInput, "Start and end dates change together.")
	}
	updated, err := h.updater.Update(ctx, principal, request.TripId.String(), patch)
	if err != nil {
		return nil, err
	}
	return api.PatchTrip200JSONResponse(tripResponse(updated)), nil
}

func principalFrom(ctx context.Context) (user.Principal, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return principal, nil
}
