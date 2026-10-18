package commerce

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

type Service interface {
	SearchFlights(context.Context, user.Principal, usecase.FlightRequest) (usecase.FlightResult, error)
	SearchStays(context.Context, user.Principal, usecase.StayRequest) (usecase.StayResult, error)
	SearchActivities(context.Context, user.Principal, usecase.ActivityRequest) (usecase.ActivityResult, error)
	CreateHandoff(context.Context, user.Principal, usecase.HandoffRequest) (domain.Handoff, error)
	CreateWatch(context.Context, user.Principal, usecase.WatchRequest) (pricewatch.Watch, error)
	Watches(context.Context, user.Principal) ([]pricewatch.Watch, error)
	DeleteWatch(context.Context, user.Principal, string) error
}

type Handler struct {
	service Service
	timer   clock.Clock
}

func NewHandler(service Service, timer clock.Clock) *Handler {
	return &Handler{service: service, timer: timer}
}

func (h *Handler) SearchFlightOffers(ctx context.Context, request api.SearchFlightOffersRequestObject) (api.SearchFlightOffersResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "Search criteria are required.")
	}
	input := usecase.FlightRequest{
		TripID: request.TripId.String(), Origin: request.Body.Origin, Destination: request.Body.Destination,
		DepartDate: calendar(request.Body.DepartDate.Time), Cabin: enumText(request.Body.Cabin),
	}
	if request.Body.ReturnDate != nil {
		input.ReturnDate = calendar(request.Body.ReturnDate.Time)
	}
	if request.Body.Budget != nil {
		input.BudgetMinor = request.Body.Budget.AmountMinor
	}
	if request.Body.MaximumMinutes != nil {
		input.MaximumMinutes = *request.Body.MaximumMinutes
	}
	if request.Body.PreferFewerStops != nil {
		input.PreferFewerStops = *request.Body.PreferFewerStops
	}
	result, err := h.service.SearchFlights(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.SearchFlightOffers200JSONResponse(api.FlightSearchResponse{
		Offers: flightOffersOf(result.Offers, h.timer.Now()), Status: statusOf(result.Status),
		Disclosure: disclosureOf(result.Status.Partners),
	}), nil
}

func (h *Handler) SearchStayOffers(ctx context.Context, request api.SearchStayOffersRequestObject) (api.SearchStayOffersResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "Search criteria are required.")
	}
	input := usecase.StayRequest{
		TripID:  request.TripId.String(),
		CheckIn: calendar(request.Body.CheckIn.Time), CheckOut: calendar(request.Body.CheckOut.Time),
	}
	if request.Body.Budget != nil {
		input.BudgetMinor = request.Body.Budget.AmountMinor
	}
	if request.Body.RadiusMeters != nil {
		input.RadiusMeters = *request.Body.RadiusMeters
	}
	if request.Body.MaximumWalkMeters != nil {
		input.MaximumWalkMeters = *request.Body.MaximumWalkMeters
	}
	result, err := h.service.SearchStays(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.SearchStayOffers200JSONResponse(api.StaySearchResponse{
		Offers: stayOffersOf(result.Offers, h.timer.Now()), Status: statusOf(result.Status),
		Disclosure: disclosureOf(result.Status.Partners),
	}), nil
}

func (h *Handler) SearchActivityOffers(ctx context.Context, request api.SearchActivityOffersRequestObject) (api.SearchActivityOffersResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	input := usecase.ActivityRequest{TripID: request.TripId.String()}
	if request.Body != nil {
		input = applyActivityBody(input, *request.Body)
	}
	result, err := h.service.SearchActivities(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.SearchActivityOffers200JSONResponse(api.ActivitySearchResponse{
		Offers: activityOffersOf(result.Offers, h.timer.Now()), Status: statusOf(result.Status),
		Disclosure: disclosureOf(result.Status.Partners),
	}), nil
}

func applyActivityBody(input usecase.ActivityRequest, body api.ActivitySearchRequest) usecase.ActivityRequest {
	if body.Date != nil {
		input.Date = calendar(body.Date.Time)
	}
	if body.FreeFrom != nil {
		input.FreeFrom = body.FreeFrom.UTC()
	}
	if body.FreeUntil != nil {
		input.FreeUntil = body.FreeUntil.UTC()
	}
	if body.Budget != nil {
		input.BudgetMinor = body.Budget.AmountMinor
	}
	if body.Interests != nil {
		input.Interests = *body.Interests
	}
	if body.MaximumDistanceMeters != nil {
		input.MaximumDistanceMeters = *body.MaximumDistanceMeters
	}
	return input
}

func (h *Handler) CreateBookingHandoff(ctx context.Context, request api.CreateBookingHandoffRequestObject) (api.CreateBookingHandoffResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "An offer is required.")
	}
	input := usecase.HandoffRequest{OfferID: request.Body.OfferId}
	if request.Body.TripId != nil {
		input.TripID = request.Body.TripId.String()
	}
	handoff, err := h.service.CreateHandoff(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.CreateBookingHandoff201JSONResponse(handoffOf(handoff)), nil
}

func principalFrom(ctx context.Context) (user.Principal, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return principal, nil
}

func calendar(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func enumText[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
