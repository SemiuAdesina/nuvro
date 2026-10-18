package runtime

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	domainlanguage "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
	languageusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/language"
	travelusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

type TravelService interface {
	Active(context.Context, user.Principal, travelusecase.ActiveRequest) (travelusecase.ActiveResult, error)
	Replan(context.Context, user.Principal, travelusecase.ReplanRequest) (travelusecase.ReplanResult, error)
	ApplyReplan(ctx context.Context, principal user.Principal, tripID, proposalID string) (itinerary.Itinerary, error)
	OfflinePack(ctx context.Context, principal user.Principal, tripID string, phrases travelusecase.PhraseSource) (travelusecase.OfflinePack, error)
}

type AlertService interface {
	List(ctx context.Context, principal user.Principal, tripID string, includeResolved bool) ([]alert.Alert, error)
	UpdateStatus(ctx context.Context, principal user.Principal, tripID, alertID string, next alert.Status) (alert.Alert, error)
}

type LanguageService interface {
	Translate(context.Context, user.Principal, languageusecase.TranslateRequest) (domainlanguage.Translation, error)
	Phrases(context.Context, user.Principal, languageusecase.PhraseRequest) (domainlanguage.Pack, error)
}

type Handler struct {
	travel   TravelService
	alerts   AlertService
	language LanguageService
	phrases  travelusecase.PhraseSource
}

func NewHandler(travel TravelService, alerts AlertService, language LanguageService, phrases travelusecase.PhraseSource) *Handler {
	return &Handler{travel: travel, alerts: alerts, language: language, phrases: phrases}
}

func (h *Handler) GetActiveTrip(ctx context.Context, request api.GetActiveTripRequestObject) (api.GetActiveTripResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	input := travelusecase.ActiveRequest{TripID: request.TripId.String(), Language: text(request.Params.Language)}
	if request.Params.Latitude != nil && request.Params.Longitude != nil {
		input.Location = &geo.Point{
			Latitude: float64(*request.Params.Latitude), Longitude: float64(*request.Params.Longitude),
		}
	}
	result, err := h.travel.Active(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.GetActiveTrip200JSONResponse(activeOf(result)), nil
}

func (h *Handler) ProposeReplan(ctx context.Context, request api.ProposeReplanRequestObject) (api.ProposeReplanResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A replanning trigger is required.")
	}
	input := travelusecase.ReplanRequest{
		TripID: request.TripId.String(), Trigger: string(request.Body.Trigger),
		Message: text(request.Body.Message), Language: text(request.Body.Language),
	}
	if request.Body.Location != nil {
		input.Location = &geo.Point{
			Latitude: float64(request.Body.Location.Latitude), Longitude: float64(request.Body.Location.Longitude),
		}
	}
	result, err := h.travel.Replan(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.ProposeReplan200JSONResponse(proposalOf(result.Proposal)), nil
}

func (h *Handler) ApplyReplan(ctx context.Context, request api.ApplyReplanRequestObject) (api.ApplyReplanResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := h.travel.ApplyReplan(ctx, principal, request.TripId.String(), request.ProposalId.String())
	if err != nil {
		return nil, err
	}
	return api.ApplyReplan200JSONResponse(journey.ItineraryOf(plan)), nil
}

func (h *Handler) ListTripAlerts(ctx context.Context, request api.ListTripAlertsRequestObject) (api.ListTripAlertsResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	includeResolved := request.Params.IncludeResolved != nil && *request.Params.IncludeResolved
	stored, err := h.alerts.List(ctx, principal, request.TripId.String(), includeResolved)
	if err != nil {
		return nil, err
	}
	return api.ListTripAlerts200JSONResponse(api.AlertList{Alerts: alertsOf(stored)}), nil
}

func (h *Handler) PatchTripAlert(ctx context.Context, request api.PatchTripAlertRequestObject) (api.PatchTripAlertResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A status is required.")
	}
	updated, err := h.alerts.UpdateStatus(ctx, principal, request.TripId.String(),
		request.AlertId.String(), alert.Status(request.Body.Status))
	if err != nil {
		return nil, err
	}
	return api.PatchTripAlert200JSONResponse(alertOf(updated)), nil
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
