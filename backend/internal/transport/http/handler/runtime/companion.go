package runtime

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
	languageusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/language"
	travelusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

func (h *Handler) TranslateForTrip(ctx context.Context, request api.TranslateForTripRequestObject) (api.TranslateForTripResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "Text to translate is required.")
	}
	result, err := h.language.Translate(ctx, principal, languageusecase.TranslateRequest{
		TripID: request.TripId.String(), Text: request.Body.Text,
		Source: text(request.Body.SourceLanguage), Target: text(request.Body.TargetLanguage),
		Context: text(request.Body.Context),
	})
	if err != nil {
		return nil, err
	}
	return api.TranslateForTrip200JSONResponse(translationOf(result)), nil
}

func (h *Handler) GetTripPhrases(ctx context.Context, request api.GetTripPhrasesRequestObject) (api.GetTripPhrasesResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	input := languageusecase.PhraseRequest{
		TripID: request.TripId.String(), Target: text(request.Params.TargetLanguage),
	}
	if request.Params.Category != nil {
		input.Categories = append(input.Categories, *request.Params.Category...)
	}
	pack, err := h.language.Phrases(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.GetTripPhrases200JSONResponse(packOf(pack)), nil
}

func (h *Handler) GetOfflinePack(ctx context.Context, request api.GetOfflinePackRequestObject) (api.GetOfflinePackResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	pack, err := h.travel.OfflinePack(ctx, principal, request.TripId.String(), h.phrases)
	if err != nil {
		return nil, err
	}
	return api.GetOfflinePack200JSONResponse(offlineOf(pack)), nil
}

func offlineOf(pack travelusecase.OfflinePack) api.OfflinePack {
	days := make([]api.ItineraryDay, 0, len(pack.Days))
	for _, day := range pack.Days {
		days = append(days, journey.DayOf(day))
	}
	value := api.OfflinePack{
		Version: pack.Version, TripId: uuid(pack.TripID), Timezone: pack.Timezone,
		GeneratedAt: pack.GeneratedAt.UTC(), ExpiresAt: pack.ExpiresAt.UTC(),
		Days: days, Phrases: phrasesOf(pack.Phrases), Alerts: alertsOf(pack.Alerts),
		Excluded: pack.Excluded,
	}
	if freshness := freshnessOf(pack.Freshness); freshness != nil {
		value.Freshness = *freshness
	}
	if pack.TripVersion > 0 {
		version := pack.TripVersion
		value.ItineraryVersion = &version
	}
	return value
}
