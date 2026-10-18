package commerce

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

func (h *Handler) ListPriceWatches(ctx context.Context, _ api.ListPriceWatchesRequestObject) (api.ListPriceWatchesResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	stored, err := h.service.Watches(ctx, principal)
	if err != nil {
		return nil, err
	}
	watches := make([]api.PriceWatch, 0, len(stored))
	for _, watch := range stored {
		watches = append(watches, watchOf(watch))
	}
	return api.ListPriceWatches200JSONResponse(api.PriceWatchList{Watches: watches}), nil
}

func (h *Handler) CreatePriceWatch(ctx context.Context, request api.CreatePriceWatchRequestObject) (api.CreatePriceWatchResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if request.Body == nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "A price watch is required.")
	}
	input := usecase.WatchRequest{
		Product: string(request.Body.Product), Target: targetFrom(request.Body.Target),
		BaselineMinor: request.Body.Baseline.AmountMinor, Currency: request.Body.Baseline.Currency,
	}
	if request.Body.TripId != nil {
		input.TripID = request.Body.TripId.String()
	}
	if request.Body.Threshold != nil {
		input.Threshold = thresholdFrom(*request.Body.Threshold)
	}
	watch, err := h.service.CreateWatch(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.CreatePriceWatch201JSONResponse(watchOf(watch)), nil
}

func (h *Handler) DeletePriceWatch(ctx context.Context, request api.DeletePriceWatchRequestObject) (api.DeletePriceWatchResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.service.DeleteWatch(ctx, principal, request.WatchId.String()); err != nil {
		return nil, err
	}
	return api.DeletePriceWatch204Response{}, nil
}

func targetFrom(target api.PriceWatchTarget) pricewatch.Target {
	value := pricewatch.Target{
		StartDate: calendar(target.StartDate.Time), EndDate: calendar(target.EndDate.Time),
		Adults: target.Adults, Cabin: enumText(target.Cabin),
	}
	if target.Children != nil {
		value.Children = *target.Children
	}
	if target.Origin != nil {
		value.Origin = *target.Origin
	}
	if target.Destination != nil {
		value.Destination = *target.Destination
	}
	if target.PropertyId != nil {
		value.PropertyID = *target.PropertyId
	}
	if target.PlaceId != nil {
		value.PlaceID = *target.PlaceId
	}
	return value
}

func thresholdFrom(threshold api.PriceWatchThreshold) pricewatch.Threshold {
	value := pricewatch.Threshold{}
	if threshold.PercentageDrop != nil {
		value.PercentageDrop = *threshold.PercentageDrop
	}
	if threshold.AbsoluteMinor != nil {
		value.AbsoluteMinor = *threshold.AbsoluteMinor
	}
	return value
}
