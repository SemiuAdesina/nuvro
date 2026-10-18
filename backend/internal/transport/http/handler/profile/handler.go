package profile

import (
	"context"
	"reflect"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
)

type Service interface {
	Profile(context.Context, user.Principal) (user.Profile, error)
	UpdateProfile(context.Context, user.Principal, user.ProfilePatch) (user.Profile, error)
	Preferences(context.Context, user.Principal) (user.Preferences, error)
	UpdatePreferences(context.Context, user.Principal, user.PreferencesPatch) (user.Preferences, error)
}

type Handler struct {
	service Service
	eraser  Eraser
	linker  Linker
}

func NewHandler(service Service, eraser Eraser) *Handler {
	return &Handler{service: service, eraser: eraser}
}

func (h *Handler) WithLinker(linker Linker) *Handler {
	if linker == nil || reflect.ValueOf(linker).IsNil() {
		return h
	}
	h.linker = linker
	return h
}

func (h *Handler) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	found, err := h.service.Profile(ctx, principal)
	if err != nil {
		return nil, err
	}
	return api.GetMe200JSONResponse(profileResponse(found)), nil
}

func (h *Handler) PatchMe(ctx context.Context, request api.PatchMeRequestObject) (api.PatchMeResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := h.service.UpdateProfile(ctx, principal, profilePatch(*request.Body))
	if err != nil {
		return nil, err
	}
	return api.PatchMe200JSONResponse(profileResponse(updated)), nil
}

func (h *Handler) GetPreferences(ctx context.Context, _ api.GetPreferencesRequestObject) (api.GetPreferencesResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	found, err := h.service.Preferences(ctx, principal)
	if err != nil {
		return nil, err
	}
	return api.GetPreferences200JSONResponse(preferencesResponse(found)), nil
}

func (h *Handler) PatchPreferences(ctx context.Context, request api.PatchPreferencesRequestObject) (api.PatchPreferencesResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	patch, ok := preferencesPatch(*request.Body)
	if !ok {
		return nil, apperror.New(apperror.CodeInvalidInput, "Preference values are invalid.")
	}
	updated, err := h.service.UpdatePreferences(ctx, principal, patch)
	if err != nil {
		return nil, err
	}
	return api.PatchPreferences200JSONResponse(preferencesResponse(updated)), nil
}

func principalFrom(ctx context.Context) (user.Principal, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return principal, nil
}
