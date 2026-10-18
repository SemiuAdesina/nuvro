package intelligence

import (
	"context"
	"time"

	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/destination"
)

type Service interface {
	Detail(context.Context, user.Principal, destination.Request) (intel.Intelligence, error)
}

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetDestination(ctx context.Context, request api.GetDestinationRequestObject) (api.GetDestinationResponseObject, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	result, err := h.service.Detail(ctx, principal, destination.Request{
		DestinationID: request.DestinationId,
		Start:         dateValue(request.Params.StartDate),
		End:           dateValue(request.Params.EndDate),
	})
	if err != nil {
		return nil, err
	}
	return api.GetDestination200JSONResponse(detailOf(result)), nil
}

func (h *Handler) GetDestinationMedia(ctx context.Context, request api.GetDestinationMediaRequestObject) (api.GetDestinationMediaResponseObject, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	result, err := h.service.Detail(ctx, principal, destination.Request{DestinationID: request.DestinationId})
	if err != nil {
		return nil, err
	}
	return api.GetDestinationMedia200JSONResponse(api.DestinationMedia{
		DestinationId: result.Destination.ID,
		Media:         mediaOf(result.Media),
		Section:       sectionOf(result.MediaSection),
	}), nil
}

func detailOf(result intel.Intelligence) api.DestinationDetail {
	record := result.Destination
	summary := api.DestinationSummary{
		Id: record.ID, Name: record.Name, CountryCode: record.CountryCode, Timezone: record.Timezone,
		PrimaryAirport: record.PrimaryAirport,
		Coordinates:    api.GeoPoint{Latitude: float32(record.Latitude), Longitude: float32(record.Longitude)},
	}
	if record.Region != "" {
		region := record.Region
		summary.Region = &region
	}
	return api.DestinationDetail{
		Destination: summary, GeneratedAt: result.GeneratedAt.UTC(),
		Weather: WeatherOf(result.Weather), WeatherSection: sectionOf(result.WeatherSection),
		ThingsToDo: placesOf(result.ThingsToDo), PlacesSection: sectionOf(result.PlacesSection),
		Media: mediaOf(result.Media), MediaSection: sectionOf(result.MediaSection),
	}
}

func dateValue(value *openapi.Date) time.Time {
	if value == nil {
		return time.Time{}
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
