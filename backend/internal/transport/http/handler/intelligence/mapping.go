package intelligence

import (
	"time"

	openapi "github.com/oapi-codegen/runtime/types"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func sectionOf(section intel.Section) api.Section {
	value := api.Section{Status: api.SectionStatus(section.Status)}
	if section.Reason != "" {
		reason := section.Reason
		value.Reason = &reason
	}
	if !section.Freshness.ObservedAt.IsZero() {
		value.Freshness = &api.Freshness{
			ObservedAt: section.Freshness.ObservedAt.UTC(),
			ValidUntil: section.Freshness.ValidUntil.UTC(),
			Status:     api.FreshnessStatus(section.Freshness.Status),
		}
	}
	return value
}

func mediaOf(assets []domain.Media) []api.MediaAsset {
	values := make([]api.MediaAsset, 0, len(assets))
	for _, asset := range assets {
		values = append(values, api.MediaAsset{
			Url: asset.URL, Width: asset.Width, Height: asset.Height,
			AspectRatio: float32(asset.AspectRatio), AltText: optional(asset.AltText),
			Attribution: attributionOf(asset.Attribution), ExpiresAt: asset.ExpiresAt,
		})
	}
	return values
}

func attributionOf(values []evidence.Attribution) []api.Attribution {
	attributions := make([]api.Attribution, 0, len(values))
	for _, value := range values {
		if !value.Valid() {
			continue
		}
		attributions = append(attributions, api.Attribution{
			Text: value.Text, Provider: value.Provider, Uri: optional(value.URI),
		})
	}
	return attributions
}

func placesOf(found []place.Place) []api.PlaceSummary {
	summaries := make([]api.PlaceSummary, 0, len(found))
	for _, value := range found {
		summaries = append(summaries, PlaceSummaryOf(value))
	}
	return summaries
}

func PlaceSummaryOf(value place.Place) api.PlaceSummary {
	summary := api.PlaceSummary{
		Id: value.ID, Name: value.Name, Category: api.PlaceSummaryCategory(value.Category),
		Location:      api.GeoPoint{Latitude: float32(value.Location.Latitude), Longitude: float32(value.Location.Longitude)},
		OpeningStatus: api.PlaceSummaryOpeningStatus(value.Status),
		Address:       optional(value.Address),
		Media:         pointerTo(mediaOf(value.Media)),
		Attribution:   pointerTo(attributionOf(value.Attribution)),
	}
	if value.Rating > 0 {
		rating := float32(value.Rating)
		summary.Rating = &rating
	}
	if value.ReviewCount > 0 {
		count := value.ReviewCount
		summary.ReviewCount = &count
	}
	if value.PriceLevel > 0 {
		level := value.PriceLevel
		summary.PriceLevel = &level
	}
	minutes := int(value.Duration().Minutes())
	summary.TypicalMinutes = &minutes
	return summary
}

func WeatherOf(window weather.Window) api.WeatherWindow {
	value := api.WeatherWindow{Availability: api.WeatherWindowAvailability(window.Availability)}
	if window.Current != nil {
		current := observationOf(*window.Current)
		value.Current = &current
	}
	if len(window.Hourly) > 0 {
		hourly := make([]api.WeatherObservation, 0, len(window.Hourly))
		for _, observation := range window.Hourly {
			hourly = append(hourly, observationOf(observation))
		}
		value.Hourly = &hourly
	}
	if len(window.Daily) > 0 {
		daily := make([]api.WeatherDay, 0, len(window.Daily))
		for _, day := range window.Daily {
			daily = append(daily, dayOf(day))
		}
		value.Daily = &daily
	}
	return value
}

func observationOf(observation weather.Observation) api.WeatherObservation {
	chance := observation.PrecipitationChance
	humidity := observation.HumidityPercent
	uv := observation.UVIndex
	feelsLike := float32(observation.FeelsLikeCelsius)
	wind := float32(observation.WindKilometresPerHour)
	return api.WeatherObservation{
		At: observation.At.UTC(), Condition: api.WeatherObservationCondition(observation.Condition),
		TemperatureCelsius: float32(observation.TemperatureCelsius), FeelsLikeCelsius: &feelsLike,
		PrecipitationChance: &chance, WindKilometresPerHour: &wind,
		HumidityPercent: &humidity, UvIndex: &uv,
	}
}

func dayOf(day weather.Day) api.WeatherDay {
	chance := day.PrecipitationChance
	return api.WeatherDay{
		Date: dateOf(day.Date), Condition: api.WeatherDayCondition(day.Condition),
		HighCelsius: float32(day.HighCelsius), LowCelsius: float32(day.LowCelsius),
		PrecipitationChance: &chance,
	}
}

func dateOf(value time.Time) openapi.Date {
	return openapi.Date{Time: value.UTC()}
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func pointerTo[T any](values []T) *[]T {
	if len(values) == 0 {
		return nil
	}
	return &values
}
