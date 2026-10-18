package journey

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/intelligence"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
)

func (h *Handler) SearchNearby(ctx context.Context, request api.SearchNearbyRequestObject) (api.SearchNearbyResponseObject, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return nil, err
	}
	body := request.Body
	input := usecase.NearbyRequest{
		TripID: request.TripId.String(),
		Location: geo.Point{
			Latitude: float64(body.Location.Latitude), Longitude: float64(body.Location.Longitude),
		},
		Categories: categoriesOf(body.Categories),
		Language:   text(body.Language),
	}
	if body.RadiusMeters != nil {
		input.RadiusMeters = *body.RadiusMeters
	}
	if body.OpenNow != nil {
		input.OpenNow = *body.OpenNow
	}
	if body.AvailableMinutes != nil {
		input.AvailableMinutes = *body.AvailableMinutes
	}
	if body.Limit != nil {
		input.Limit = *body.Limit
	}
	result, err := h.service.Nearby(ctx, principal, input)
	if err != nil {
		return nil, err
	}
	return api.SearchNearby200JSONResponse(nearbyOf(result)), nil
}

func nearbyOf(result usecase.NearbyResult) api.NearbySearchResponse {
	ranked := make([]api.NearbyPlace, 0, len(result.Places))
	for _, value := range result.Places {
		ranked = append(ranked, rankedOf(value))
	}
	response := api.NearbySearchResponse{Places: ranked, AvailableMinutes: result.AvailableMinutes}
	if result.NextItem != nil {
		next := ItemOf(*result.NextItem)
		response.NextItem = &next
	}
	return response
}

func rankedOf(value places.Ranked) api.NearbyPlace {
	summary := intelligence.PlaceSummaryOf(value.Place)
	ranked := api.NearbyPlace{
		Place: summary, Score: value.Score, FitsAvailableTime: value.Fits,
		WeatherImpact: api.NearbyPlaceWeatherImpact(value.WeatherImpact), Reasons: value.Reasons,
	}
	if value.TravelMinutes >= 0 {
		minutes := value.TravelMinutes
		ranked.TravelMinutes = &minutes
	}
	if ranked.Reasons == nil {
		ranked.Reasons = []string{}
	}
	return ranked
}

func categoriesOf(values *[]api.NearbySearchRequestCategories) []place.Category {
	if values == nil {
		return nil
	}
	categories := make([]place.Category, 0, len(*values))
	for _, value := range *values {
		categories = append(categories, place.ParseCategory(string(value)))
	}
	return categories
}
