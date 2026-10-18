package personalization

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

type Service interface {
	Context(context.Context, user.Principal) (domain.Context, error)
	Insights(context.Context, user.Principal) ([]domain.Insight, error)
	UpdateInsight(ctx context.Context, principal user.Principal, insightID string, next domain.Status) (domain.Insight, error)
	SetHistoryLearning(context.Context, user.Principal, bool) (domain.Context, error)
	ClearLearning(context.Context, user.Principal) error
	SubmitFeedback(context.Context, user.Principal, usecase.FeedbackRequest) (feedbackdomain.TripFeedback, error)
	Feedback(context.Context, user.Principal, string) (feedbackdomain.TripFeedback, error)
}

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetPersonalization(ctx context.Context, _ api.GetPersonalizationRequestObject) (api.GetPersonalizationResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	value, err := h.service.Context(ctx, principal)
	if err != nil {
		return nil, err
	}
	return api.GetPersonalization200JSONResponse(settingsOf(value)), nil
}

func (h *Handler) PatchPersonalization(ctx context.Context, request api.PatchPersonalizationRequestObject) (api.PatchPersonalizationResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A setting is required.")
	}
	value, err := h.service.SetHistoryLearning(ctx, principal, request.Body.HistoryLearningEnabled)
	if err != nil {
		return nil, err
	}
	return api.PatchPersonalization200JSONResponse(settingsOf(value)), nil
}

func (h *Handler) ClearPersonalization(ctx context.Context, _ api.ClearPersonalizationRequestObject) (api.ClearPersonalizationResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.service.ClearLearning(ctx, principal); err != nil {
		return nil, err
	}
	return api.ClearPersonalization204Response{}, nil
}

func (h *Handler) ListPersonalizationInsights(ctx context.Context, _ api.ListPersonalizationInsightsRequestObject) (api.ListPersonalizationInsightsResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	insights, err := h.service.Insights(ctx, principal)
	if err != nil {
		return nil, err
	}
	return api.ListPersonalizationInsights200JSONResponse(api.PersonalizationInsightList{
		Insights: insightsOf(insights),
	}), nil
}

func (h *Handler) PatchPersonalizationInsight(ctx context.Context, request api.PatchPersonalizationInsightRequestObject) (api.PatchPersonalizationInsightResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A status is required.")
	}
	insight, err := h.service.UpdateInsight(ctx, principal, request.InsightId.String(),
		domain.Status(request.Body.Status))
	if err != nil {
		return nil, err
	}
	return api.PatchPersonalizationInsight200JSONResponse(insightOf(insight)), nil
}

func (h *Handler) SubmitTripFeedback(ctx context.Context, request api.SubmitTripFeedbackRequestObject) (api.SubmitTripFeedbackResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "Feedback is required.")
	}
	input := usecase.FeedbackRequest{
		TripID:    request.TripId.String(),
		Positives: aspectCodes(request.Body.Positives),
		Negatives: aspectCodes(request.Body.Negatives),
	}
	if request.Body.Pace != nil {
		input.Pace = string(*request.Body.Pace)
	}
	if request.Body.TravelBurden != nil {
		input.TravelBurden = string(*request.Body.TravelBurden)
	}
	if request.Body.ReturnIntent != nil {
		input.ReturnIntent = string(*request.Body.ReturnIntent)
	}
	stored, err := h.service.SubmitFeedback(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.SubmitTripFeedback200JSONResponse(feedbackOf(stored)), nil
}

func (h *Handler) GetTripFeedback(ctx context.Context, request api.GetTripFeedbackRequestObject) (api.GetTripFeedbackResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	stored, err := h.service.Feedback(ctx, principal, request.TripId.String())
	if err != nil {
		return nil, err
	}
	return api.GetTripFeedback200JSONResponse(feedbackOf(stored)), nil
}

func principalFrom(ctx context.Context) (user.Principal, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return principal, nil
}
