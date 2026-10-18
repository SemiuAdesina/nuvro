package googleplaces

import (
	"context"
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultBaseURL = "https://places.googleapis.com"

const searchFieldMask = "places.id,places.displayName,places.formattedAddress,places.location," +
	"places.primaryType,places.types,places.rating,places.userRatingCount,places.priceLevel," +
	"places.currentOpeningHours.openNow,places.photos"

const detailsFieldMask = "id,displayName,formattedAddress,location,primaryType,types,rating," +
	"userRatingCount,priceLevel,regularOpeningHours,currentOpeningHours.openNow,photos,utcOffsetMinutes"

type Adapter struct {
	client *httpclient.Client
	apiKey string
	timer  clock.Clock
}

func New(client *httpclient.Client, apiKey string, timer clock.Clock) *Adapter {
	return &Adapter{client: client, apiKey: apiKey, timer: timer}
}

func (a *Adapter) Nearby(ctx context.Context, query places.NearbyQuery) ([]place.Place, error) {
	body := map[string]any{
		"maxResultCount": query.Limit,
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{"latitude": query.Centre.Latitude, "longitude": query.Centre.Longitude},
				"radius": float64(query.RadiusMeters),
			},
		},
	}
	if types := includedTypes(query.Categories); len(types) > 0 {
		body["includedTypes"] = types
	}
	if query.Language != "" {
		body["languageCode"] = query.Language
	}
	var response searchResponse
	err := a.client.Post(ctx, httpclient.Request{
		Path: "/v1/places:searchNearby", Body: body, Headers: a.headers(searchFieldMask),
	}, &response)
	if err != nil {
		return nil, err
	}
	return a.convert(response.Places), nil
}

func (a *Adapter) Text(ctx context.Context, query places.TextQuery) ([]place.Place, error) {
	body := map[string]any{"textQuery": strings.TrimSpace(query.Text), "maxResultCount": query.Limit}
	if !query.Centre.IsZero() {
		body["locationBias"] = map[string]any{
			"circle": map[string]any{
				"center": map[string]any{"latitude": query.Centre.Latitude, "longitude": query.Centre.Longitude},
				"radius": float64(maxRadius(query.RadiusMeters)),
			},
		}
	}
	if query.Language != "" {
		body["languageCode"] = query.Language
	}
	var response searchResponse
	err := a.client.Post(ctx, httpclient.Request{
		Path: "/v1/places:searchText", Body: body, Headers: a.headers(searchFieldMask),
	}, &response)
	if err != nil {
		return nil, err
	}
	return a.convert(response.Places), nil
}

func (a *Adapter) Details(ctx context.Context, query places.DetailsQuery) (place.Place, error) {
	if query.Reference.Provider != Provider || query.Reference.ExternalID == "" {
		return place.Place{}, httpclient.ErrMalformed
	}
	path := "/v1/places/" + query.Reference.ExternalID
	if query.Language != "" {
		path += "?languageCode=" + query.Language
	}
	var payload placePayload
	if err := a.client.Get(ctx, httpclient.Request{Path: path, Headers: a.headers(detailsFieldMask)}, &payload); err != nil {
		return place.Place{}, err
	}
	converted, ok := toPlace(payload, a.timer.Now())
	if !ok {
		return place.Place{}, httpclient.ErrMalformed
	}
	return converted, nil
}

func (a *Adapter) convert(payloads []placePayload) []place.Place {
	observedAt := a.timer.Now()
	results := make([]place.Place, 0, len(payloads))
	for _, payload := range payloads {
		converted, ok := toPlace(payload, observedAt)
		if !ok {
			continue
		}
		results = append(results, converted)
	}
	return results
}

func (a *Adapter) headers(fieldMask string) map[string]string {
	return map[string]string{"X-Goog-Api-Key": a.apiKey, "X-Goog-FieldMask": fieldMask}
}

func maxRadius(requested int) int {
	if requested < 1 || requested > 50000 {
		return 50000
	}
	return requested
}

var _ places.SearchProvider = (*Adapter)(nil)
var _ places.DetailsProvider = (*Adapter)(nil)
