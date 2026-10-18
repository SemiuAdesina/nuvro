package weather

import (
	"context"
	"strconv"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const CurrentCacheTTL = 10 * time.Minute
const ForecastCacheTTL = time.Hour

type Engine struct {
	provider Provider
	cache    Cache
	timer    clock.Clock
}

func New(provider Provider, cache Cache, timer clock.Clock) *Engine {
	return &Engine{provider: provider, cache: cache, timer: timer}
}

func (e *Engine) Configured() bool { return e.provider != nil }

func (e *Engine) For(ctx context.Context, location geo.Point, start, end time.Time, language string) weather.Window {
	if e.provider == nil {
		return weather.Unavailable(weather.ProviderUnavailable)
	}
	now := e.timer.Now().UTC()
	if !weather.WithinDailyHorizon(now, start) && !weather.WithinHourlyHorizon(now, start) {
		return weather.Unavailable(weather.BeyondForecast)
	}
	key := windowKey(location, start, end, language)
	if cached, found := e.fromCache(ctx, key); found {
		return cached
	}
	window := e.collect(ctx, location, end, language, now)
	e.store(ctx, key, window)
	return window
}

func (e *Engine) collect(ctx context.Context, location geo.Point, end time.Time, language string, now time.Time) weather.Window {
	query := Query{Location: location, Language: language, Days: dayCount(now, end), Hours: hourCount(now, end)}
	window := weather.Window{Availability: weather.Available}
	if current, err := e.provider.Current(ctx, query); err == nil && current.Validate() == nil {
		window.Current = &current
	}
	if hourly, err := e.provider.Hourly(ctx, query); err == nil {
		window.Hourly = acceptObservations(hourly, now)
	}
	if daily, err := e.provider.Daily(ctx, query); err == nil {
		window.Daily = acceptDays(daily, now)
	}
	if window.Current == nil && len(window.Hourly) == 0 && len(window.Daily) == 0 {
		return weather.Unavailable(weather.ProviderUnavailable)
	}
	window.Source = source(e.timer.Now())
	return window
}

func (e *Engine) Impact(window weather.Window, at time.Time, sensitivity weather.Sensitivity) weather.Impact {
	if !window.Known() {
		return weather.UnknownImpact
	}
	if observation, found := window.HourAt(at); found {
		return weather.Assess(observation, sensitivity)
	}
	day, found := window.DayOn(at)
	if !found {
		return weather.UnknownImpact
	}
	return weather.Assess(weather.Observation{
		At: at, Condition: day.Condition, TemperatureCelsius: day.HighCelsius,
		PrecipitationChance: day.PrecipitationChance,
	}, sensitivity)
}

func acceptObservations(values []weather.Observation, now time.Time) []weather.Observation {
	accepted := make([]weather.Observation, 0, len(values))
	for _, observation := range values {
		if observation.Validate() != nil || !weather.WithinHourlyHorizon(now, observation.At) {
			continue
		}
		accepted = append(accepted, observation)
	}
	return accepted
}

func acceptDays(values []weather.Day, now time.Time) []weather.Day {
	accepted := make([]weather.Day, 0, len(values))
	for _, day := range values {
		if day.Date.IsZero() || !day.Condition.Valid() || !weather.WithinDailyHorizon(now, day.Date) {
			continue
		}
		accepted = append(accepted, day)
	}
	return accepted
}

func dayCount(now, end time.Time) int {
	days := int(end.UTC().Sub(now.Truncate(24*time.Hour)).Hours()/24) + 1
	return clampCount(days, weather.DailyHorizonDays)
}

func hourCount(now, end time.Time) int {
	hours := int(end.UTC().Sub(now).Hours()) + 1
	return clampCount(hours, weather.HourlyHorizonHours)
}

func clampCount(value, maximum int) int {
	if value < 1 {
		return 1
	}
	if value > maximum {
		return maximum
	}
	return value
}

func (e *Engine) fromCache(ctx context.Context, key string) (weather.Window, bool) {
	if e.cache == nil {
		return weather.Window{}, false
	}
	return e.cache.Window(ctx, key)
}

func (e *Engine) store(ctx context.Context, key string, window weather.Window) {
	if e.cache == nil || !window.Known() {
		return
	}
	e.cache.StoreWindow(ctx, key, window, ForecastCacheTTL)
}

func windowKey(location geo.Point, start, end time.Time, language string) string {
	return strconv.FormatFloat(location.Latitude, 'f', 2, 64) + "," +
		strconv.FormatFloat(location.Longitude, 'f', 2, 64) + "_" +
		start.UTC().Format(time.DateOnly) + "_" + end.UTC().Format(time.DateOnly) + "_" + language
}
