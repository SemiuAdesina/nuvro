package postgres

import (
	"math"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

func bounded(value int) int32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func itemParams(itemID, planID, tripID pgtypeUUID, item itinerary.Item) queries.InsertItineraryItemParams {
	params := queries.InsertItineraryItemParams{
		ID: itemID, ItineraryID: planID, TripID: tripID, DayNumber: bounded(item.Day),
		StartsAt: item.StartAt.UTC(), EndsAt: item.EndAt.UTC(), Kind: string(item.Kind),
		Title: item.Title, Category: string(item.Category), WeatherImpact: string(item.WeatherImpact),
		EstimatedCostMinor: item.EstimatedCostMinor, CostCurrency: item.CostCurrency,
		Status: string(item.Status), Source: string(item.Source),
		UserLocked: item.UserLocked, ReservationLocked: item.ReservationLocked,
	}
	if item.PlaceID != "" {
		placeID := item.PlaceID
		params.PlaceID = &placeID
	}
	if !item.Location.IsZero() {
		latitude := item.Location.Latitude
		longitude := item.Location.Longitude
		params.Latitude = &latitude
		params.Longitude = &longitude
	}
	if item.RouteFromPrevious != nil && item.RouteFromPrevious.Known() {
		seconds := bounded(item.RouteFromPrevious.DurationSeconds)
		meters := bounded(item.RouteFromPrevious.DistanceMeters)
		mode := string(item.RouteFromPrevious.Mode)
		params.RouteSeconds = &seconds
		params.RouteMeters = &meters
		params.RouteMode = &mode
	}
	return params
}

func toItem(row queries.ListItineraryItemsRow) itinerary.Item {
	item := itinerary.Item{
		ID: uuidText(row.ID), TripID: uuidText(row.TripID), Day: int(row.DayNumber),
		StartAt: row.StartsAt.UTC(), EndAt: row.EndsAt.UTC(), Kind: itinerary.Kind(row.Kind),
		Title: row.Title, Category: place.ParseCategory(row.Category),
		WeatherImpact: weather.Impact(row.WeatherImpact), EstimatedCostMinor: row.EstimatedCostMinor,
		CostCurrency: row.CostCurrency, Status: itinerary.Status(row.Status),
		Source: itinerary.Source(row.Source), UserLocked: row.UserLocked, ReservationLocked: row.ReservationLocked,
	}
	if row.PlaceID != nil {
		item.PlaceID = *row.PlaceID
	}
	if row.Latitude != nil && row.Longitude != nil {
		item.Location = geo.Point{Latitude: *row.Latitude, Longitude: *row.Longitude}
	}
	if row.RouteSeconds != nil && row.RouteMeters != nil && row.RouteMode != nil {
		item.RouteFromPrevious = &route.Evidence{
			Status: route.Found, Mode: route.Mode(*row.RouteMode),
			DurationSeconds: int(*row.RouteSeconds), DistanceMeters: int(*row.RouteMeters),
			Confidence: route.Medium,
		}
	}
	return item
}
