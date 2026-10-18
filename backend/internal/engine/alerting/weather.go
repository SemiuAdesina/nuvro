package alerting

import (
	"strconv"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

const weatherHorizon = 12 * time.Hour

type WeatherImpact struct{}

func (WeatherImpact) Name() string { return "weather_impact" }

func (w WeatherImpact) Evaluate(input Input) []alert.Alert {
	if !input.Weather.Known() {
		return nil
	}
	alerts := make([]alert.Alert, 0)
	for _, item := range input.Itinerary.Items {
		if !affected(item, input.Now) {
			continue
		}
		impact := impactFor(input, item)
		if impact != weather.HighImpact && impact != weather.ModerateImpact {
			continue
		}
		alerts = append(alerts, w.alert(input, item, impact))
	}
	return alerts
}

func (w WeatherImpact) alert(input Input, item itinerary.Item, impact weather.Impact) alert.Alert {
	severity := alert.Attention
	if impact == weather.HighImpact {
		severity = alert.Important
	}
	value := base(input, alert.WeatherImpact, severity, "alert.weather_impact", item.ID, alert.AdjustPlan)
	value.MessageData = map[string]string{
		"activity": item.Title, "impact": string(impact),
		"startsAt": item.StartAt.UTC().Format(time.RFC3339),
	}
	value.ExpiresAt = item.EndAt.UTC()
	value.Freshness = input.Weather.Source.Freshness
	value.Fingerprint = alert.Fingerprint(input.TripID, alert.WeatherImpact, item.ID, impactWindow(item, impact))
	return value
}

func impactWindow(item itinerary.Item, impact weather.Impact) string {
	return item.StartAt.UTC().Format("2006-01-02T15") + ":" + string(impact)
}

func impactFor(input Input, item itinerary.Item) weather.Impact {
	if observation, found := input.Weather.HourAt(item.StartAt); found {
		return weather.Assess(observation, weather.SensitivityOf(item.Category))
	}
	day, found := input.Weather.DayOn(item.StartAt)
	if !found {
		return weather.UnknownImpact
	}
	return weather.Assess(weather.Observation{
		At: item.StartAt, Condition: day.Condition, TemperatureCelsius: day.HighCelsius,
		PrecipitationChance: day.PrecipitationChance,
	}, weather.SensitivityOf(item.Category))
}

func affected(item itinerary.Item, now time.Time) bool {
	switch item.Status {
	case itinerary.Completed, itinerary.Skipped, itinerary.Cancelled:
		return false
	}
	if item.StartAt.Before(now.UTC()) {
		return false
	}
	return item.StartAt.Sub(now.UTC()) <= weatherHorizon
}

func minutes(value time.Duration) string { return strconv.Itoa(int(value.Minutes())) }
