package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
)

type Service interface {
	Issue(ctx context.Context, credential, authorizationCode string) (session.Grant, error)
	Refresh(ctx context.Context, refreshToken string) (session.Grant, error)
	Revoke(context.Context, user.Principal) error
}

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateSession(ctx context.Context, request api.CreateSessionRequestObject) (api.CreateSessionResponseObject, error) {
	if h.service == nil {
		return nil, unavailable()
	}
	granted, err := h.service.Issue(ctx, request.Body.IdentityToken, codeOf(request.Body.AuthorizationCode))
	if err != nil {
		return nil, err
	}
	return api.CreateSession201JSONResponse(sessionResponse(granted)), nil
}

func (h *Handler) RefreshSession(ctx context.Context, request api.RefreshSessionRequestObject) (api.RefreshSessionResponseObject, error) {
	if h.service == nil {
		return nil, unavailable()
	}
	granted, err := h.service.Refresh(ctx, request.Body.RefreshToken)
	if err != nil {
		return nil, err
	}
	return api.RefreshSession201JSONResponse(sessionResponse(granted)), nil
}

func (h *Handler) DeleteSession(ctx context.Context, _ api.DeleteSessionRequestObject) (api.DeleteSessionResponseObject, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if h.service == nil {
		return nil, unavailable()
	}
	if err := h.service.Revoke(ctx, principal); err != nil {
		return nil, err
	}
	return api.DeleteSession204Response{}, nil
}

func codeOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func sessionResponse(granted session.Grant) api.ComponentsSchemasAuthSession {
	accountID, _ := uuid.Parse(granted.UserID)
	return api.ComponentsSchemasAuthSession{
		AccessToken: granted.AccessToken, RefreshToken: granted.RefreshToken,
		ExpiresAt: granted.ExpiresAt, AccountId: accountID,
	}
}

func unavailable() error {
	return apperror.New(apperror.CodeUnavailable, "Sessions are unavailable.")
}
