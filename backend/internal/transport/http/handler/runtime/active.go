package runtime

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/intelligence"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
	travelusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

func activeOf(result travelusecase.ActiveResult) api.ActiveTrip {
	snapshot := result.Snapshot
	today := make([]api.ItineraryItem, 0, len(snapshot.Today))
	for _, item := range snapshot.Today {
		today = append(today, journey.ItemOf(item))
	}
	value := api.ActiveTrip{
		TripId: uuid(snapshot.Trip.ID), Timezone: snapshot.Timezone, LocalNow: snapshot.LocalNow,
		Phase: api.ActiveTripPhase(snapshot.Phase), CurrentDay: snapshot.CurrentDay,
		Today: today, Departure: departureOf(snapshot.Departure),
		Actions: actionsOf(snapshot.Actions), Alerts: alertsOf(result.Alerts),
		GeneratedAt: snapshot.GeneratedAt.UTC(),
	}
	if snapshot.NextItem != nil {
		next := journey.ItemOf(*snapshot.NextItem)
		value.NextItem = &next
	}
	if snapshot.HasWeather {
		window := intelligence.WeatherOf(snapshot.Weather)
		value.Weather = &window
		value.WeatherFreshness = freshnessOf(snapshot.WeatherSection)
	}
	return value
}
