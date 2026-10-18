package journey

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
)

type Service interface {
	Generate(context.Context, user.Principal, usecase.GenerateRequest) (usecase.GenerateResult, error)
	Read(context.Context, user.Principal, string) (itinerary.Itinerary, error)
	UpdateItem(context.Context, user.Principal, usecase.ItemUpdate) (itinerary.Item, error)
	Budget(context.Context, user.Principal, string) (budgetdomain.Trip, error)
	Nearby(context.Context, user.Principal, usecase.NearbyRequest) (usecase.NearbyResult, error)
}

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) GenerateItinerary(ctx context.Context, request api.GenerateItineraryRequestObject) (api.GenerateItineraryResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	input := usecase.GenerateRequest{TripID: request.TripId.String()}
	if request.Body != nil {
		input.Pace = enumText(request.Body.Pace)
		input.Language = text(request.Body.Language)
	}
	result, err := h.service.Generate(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.GenerateItinerary200JSONResponse(api.GenerateItineraryResponse{
		Itinerary: ItineraryOf(result.Itinerary), Violations: ViolationsOf(result.Violations),
		Candidates: result.Candidates,
	}), nil
}

func (h *Handler) GetItinerary(ctx context.Context, request api.GetItineraryRequestObject) (api.GetItineraryResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := h.service.Read(ctx, principal, request.TripId.String())
	if err != nil {
		return nil, err
	}
	return api.GetItinerary200JSONResponse(ItineraryOf(plan)), nil
}

func (h *Handler) PatchItineraryItem(ctx context.Context, request api.PatchItineraryItemRequestObject) (api.PatchItineraryItemResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A change is required.")
	}
	update := usecase.ItemUpdate{
		TripID: request.TripId.String(), ItemID: request.ItemId.String(),
		StartAt: request.Body.StartAt, EndAt: request.Body.EndAt,
		Title: request.Body.Title, UserLocked: request.Body.UserLocked,
	}
	if request.Body.Status != nil {
		status := string(*request.Body.Status)
		update.Status = &status
	}
	item, err := h.service.UpdateItem(ctx, principal, update)
	if err != nil {
		return nil, err
	}
	return api.PatchItineraryItem200JSONResponse(ItemOf(item)), nil
}

func (h *Handler) GetTripBudget(ctx context.Context, request api.GetTripBudgetRequestObject) (api.GetTripBudgetResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.service.Budget(ctx, principal, request.TripId.String())
	if err != nil {
		return nil, err
	}
	return api.GetTripBudget200JSONResponse(budgetOf(result)), nil
}

func principalFrom(ctx context.Context) (user.Principal, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return principal, nil
}

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func enumText[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
