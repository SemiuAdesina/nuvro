package googleweather

import (
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

func condition(raw string) weather.Condition {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case normalized == "":
		return weather.UnknownCondition
	case strings.Contains(normalized, "THUNDERSTORM"), strings.Contains(normalized, "HAIL"):
		return weather.Storm
	case strings.Contains(normalized, "HEAVY_RAIN"), strings.Contains(normalized, "HEAVY_SHOWERS"):
		return weather.HeavyRain
	case strings.Contains(normalized, "SNOW"), strings.Contains(normalized, "SLEET"), strings.Contains(normalized, "BLIZZARD"):
		return weather.Snow
	case strings.Contains(normalized, "RAIN"), strings.Contains(normalized, "SHOWERS"), strings.Contains(normalized, "DRIZZLE"):
		return weather.Rain
	case strings.Contains(normalized, "FOG"), strings.Contains(normalized, "HAZE"), strings.Contains(normalized, "MIST"):
		return weather.Fog
	case strings.Contains(normalized, "WIND"):
		return weather.Wind
	case strings.Contains(normalized, "CLOUD"), strings.Contains(normalized, "OVERCAST"):
		return weather.Cloudy
	case strings.Contains(normalized, "CLEAR"), strings.Contains(normalized, "SUNNY"):
		return weather.Clear
	default:
		return weather.UnknownCondition
	}
}

func celsius(temperature value) (float64, bool) {
	switch strings.ToUpper(temperature.Unit) {
	case "CELSIUS", "":
		return temperature.Degrees, true
	case "FAHRENHEIT":
		return (temperature.Degrees - 32) * 5 / 9, true
	default:
		return 0, false
	}
}

func kilometresPerHour(measurement speed) float64 {
	switch strings.ToUpper(measurement.Unit) {
	case "MILES_PER_HOUR":
		return measurement.Value * 1.609344
	default:
		return measurement.Value
	}
}

func millimetres(measurement quantity) float64 {
	switch strings.ToUpper(measurement.Unit) {
	case "INCHES":
		return measurement.Value * 25.4
	default:
		return measurement.Value
	}
}

func instant(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func toObservation(payload hourPayload) (weather.Observation, bool) {
	at, ok := instant(payload.Interval.StartTime)
	if !ok {
		return weather.Observation{}, false
	}
	temperature, ok := celsius(payload.Temperature)
	if !ok {
		return weather.Observation{}, false
	}
	feelsLike, ok := celsius(payload.FeelsLikeTemperature)
	if !ok {
		feelsLike = temperature
	}
	observation := weather.Observation{
		At: at, Condition: condition(payload.WeatherCondition.Type),
		TemperatureCelsius: temperature, FeelsLikeCelsius: feelsLike,
		PrecipitationChance:      clampPercent(payload.Precipitation.Probability.Percent),
		PrecipitationMillimetres: millimetres(payload.Precipitation.QPF),
		WindKilometresPerHour:    kilometresPerHour(payload.Wind.Speed),
		HumidityPercent:          clampPercent(payload.RelativeHumidity),
		UVIndex:                  max(payload.UVIndex, 0),
	}
	if observation.Validate() != nil {
		return weather.Observation{}, false
	}
	return observation, true
}

func toDay(payload dayPayload) (weather.Day, bool) {
	if payload.DisplayDate.Year == 0 || payload.DisplayDate.Month == 0 || payload.DisplayDate.Day == 0 {
		return weather.Day{}, false
	}
	high, ok := celsius(payload.MaxTemperature)
	if !ok {
		return weather.Day{}, false
	}
	low, ok := celsius(payload.MinTemperature)
	if !ok {
		low = high
	}
	day := weather.Day{
		Date:        time.Date(payload.DisplayDate.Year, time.Month(payload.DisplayDate.Month), payload.DisplayDate.Day, 0, 0, 0, 0, time.UTC),
		Condition:   condition(payload.DaytimeForecast.WeatherCondition.Type),
		HighCelsius: high, LowCelsius: low,
		PrecipitationChance: clampPercent(payload.DaytimeForecast.Precipitation.Probability.Percent),
	}
	if sunrise, ok := instant(payload.SunEvents.SunriseTime); ok {
		day.SunriseAt = sunrise
	}
	if sunset, ok := instant(payload.SunEvents.SunsetTime); ok {
		day.SunsetAt = sunset
	}
	return day, true
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
