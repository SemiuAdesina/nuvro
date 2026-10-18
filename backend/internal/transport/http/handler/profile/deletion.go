package profile

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

type Eraser interface {
	Request(context.Context, user.Principal, string) (retention.Deletion, error)
}

type Linker interface {
	Link(context.Context, user.Principal, string) error
}

func (h *Handler) PutAppleIdentity(ctx context.Context, request api.PutAppleIdentityRequestObject) (api.PutAppleIdentityResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if h.linker == nil {
		return nil, apperror.New(apperror.CodeUnavailable, "Identity linking is unavailable.")
	}
	if err := h.linker.Link(ctx, principal, request.Body.AuthorizationCode); err != nil {
		return nil, err
	}
	return api.PutAppleIdentity204Response{}, nil
}

func (h *Handler) DeleteMe(ctx context.Context, request api.DeleteMeRequestObject) (api.DeleteMeResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	accepted, err := h.eraser.Request(ctx, principal, reasonOf(request.Body))
	if err != nil {
		return nil, err
	}
	return api.DeleteMe202JSONResponse(deletionResponse(accepted)), nil
}

func reasonOf(body *api.DeleteMeJSONRequestBody) string {
	if body == nil || body.Reason == nil {
		return ""
	}
	return *body.Reason
}

func deletionResponse(deletion retention.Deletion) api.ComponentsSchemasProfileAccountDeletion {
	return api.ComponentsSchemasProfileAccountDeletion{
		Status:      api.ComponentsSchemasProfileAccountDeletionStatus(deletion.Status),
		RequestedAt: deletion.RequestedAt,
		PurgeAfter:  deletion.PurgeAfter,
		PurgedAt:    deletion.PurgedAt,
	}
}
