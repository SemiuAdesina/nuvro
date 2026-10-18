package intel

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
)

type PlaceReader interface {
	Nearby(ctx context.Context, centre geo.Point, categories []place.Category, language string, limit int) ([]place.Place, error)
}

type WeatherReader interface {
	For(ctx context.Context, location geo.Point, start, end time.Time, language string) weather.Window
}

type MediaSelector interface {
	Select(ctx context.Context, context media.Context, limit int) []destination.Media
}

func (e *Engine) describeWeather(ctx context.Context, point geo.Point, request Request) (weather.Window, Section) {
	if e.weather == nil {
		return weather.Unavailable(weather.ProviderUnavailable), unavailable("weather_provider_unavailable")
	}
	window := e.weather.For(ctx, point, request.Start, request.End, request.Language)
	switch window.Availability {
	case weather.Available:
		return window, available(e.timer.Now(), weatherValidity)
	case weather.BeyondForecast:
		return window, Section{Status: NotApplicable, Reason: "beyond_forecast_horizon"}
	default:
		return window, unavailable("weather_provider_unavailable")
	}
}

func (e *Engine) describePlaces(ctx context.Context, point geo.Point, request Request) ([]place.Place, Section) {
	if e.places == nil {
		return nil, unavailable("place_provider_unavailable")
	}
	found, err := e.places.Nearby(ctx, point, interestCategories(request.Interests), request.Language, previewPlaceLimit)
	if err != nil || len(found) == 0 {
		return nil, unavailable("place_provider_unavailable")
	}
	return found, available(e.timer.Now(), placesValidity)
}

func (e *Engine) describeMedia(ctx context.Context, request Request) ([]destination.Media, Section) {
	if e.media == nil {
		return nil, unavailable("media_provider_unavailable")
	}
	selected := e.media.Select(ctx, media.Context{
		DestinationID: request.Destination.ID, StartDate: request.Start, EndDate: request.End,
		Landscape: true, Language: request.Language,
	}, mediaLimit)
	if len(selected) == 0 {
		return nil, unavailable("media_unavailable")
	}
	return selected, available(e.timer.Now(), media.SelectionTTL)
}

func interestCategories(interests []string) []place.Category {
	categories := []place.Category{place.Attraction}
	for _, interest := range interests {
		switch interest {
		case "beach":
			categories = append(categories, place.Beach)
		case "food":
			categories = append(categories, place.Restaurant, place.Cafe)
		case "culture", "history":
			categories = append(categories, place.Museum)
		case "nature":
			categories = append(categories, place.Park, place.Viewpoint)
		}
	}
	return categories
}
