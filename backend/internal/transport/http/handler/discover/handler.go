package discover

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

type Service interface {
	Matches(context.Context, user.Principal, discover.Request) (discover.Response, error)
}

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) DiscoverMatches(ctx context.Context, request api.DiscoverMatchesRequestObject) (api.DiscoverMatchesResponseObject, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	input, valid := discoveryRequest(*request.Body)
	if !valid {
		return nil, apperror.New(apperror.CodeInvalidInput, "The discovery request is not valid.")
	}
	result, err := h.service.Matches(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.DiscoverMatches200JSONResponse(response(result)), nil
}
