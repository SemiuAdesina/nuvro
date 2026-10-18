package weather

import (
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

var ErrInvalid = errors.New("weather: invalid observation")

const HourlyHorizonHours = 240
const DailyHorizonDays = 10

type Condition string

const (
	Clear            Condition = "clear"
	Cloudy           Condition = "cloudy"
	Rain             Condition = "rain"
	HeavyRain        Condition = "heavy_rain"
	Snow             Condition = "snow"
	Storm            Condition = "storm"
	Fog              Condition = "fog"
	Wind             Condition = "wind"
	UnknownCondition Condition = "unknown"
)

func (c Condition) Valid() bool {
	switch c {
	case Clear, Cloudy, Rain, HeavyRain, Snow, Storm, Fog, Wind, UnknownCondition:
		return true
	default:
		return false
	}
}

type Observation struct {
	At                       time.Time
	Condition                Condition
	TemperatureCelsius       float64
	FeelsLikeCelsius         float64
	PrecipitationChance      int
	PrecipitationMillimetres float64
	WindKilometresPerHour    float64
	HumidityPercent          int
	UVIndex                  int
}

func (o Observation) Validate() error {
	if o.At.IsZero() || !o.Condition.Valid() {
		return ErrInvalid
	}
	if o.TemperatureCelsius < -100 || o.TemperatureCelsius > 70 {
		return ErrInvalid
	}
	if o.PrecipitationChance < 0 || o.PrecipitationChance > 100 || o.HumidityPercent < 0 || o.HumidityPercent > 100 {
		return ErrInvalid
	}
	if o.WindKilometresPerHour < 0 || o.PrecipitationMillimetres < 0 || o.UVIndex < 0 {
		return ErrInvalid
	}
	return nil
}

type Day struct {
	Date                time.Time
	Condition           Condition
	HighCelsius         float64
	LowCelsius          float64
	PrecipitationChance int
	SunriseAt           time.Time
	SunsetAt            time.Time
}

type Availability string

const (
	Available           Availability = "available"
	BeyondForecast      Availability = "beyond_forecast_horizon"
	ProviderUnavailable Availability = "unavailable"
)

type Window struct {
	Availability Availability
	Current      *Observation
	Hourly       []Observation
	Daily        []Day
	Source       evidence.Source
}

func Unavailable(availability Availability) Window {
	return Window{Availability: availability}
}

func (w Window) Known() bool { return w.Availability == Available }

func (w Window) HourAt(moment time.Time) (Observation, bool) {
	target := moment.UTC().Truncate(time.Hour)
	for _, observation := range w.Hourly {
		if observation.At.UTC().Truncate(time.Hour).Equal(target) {
			return observation, true
		}
	}
	return Observation{}, false
}

func (w Window) DayOn(date time.Time) (Day, bool) {
	target := date.UTC().Format(time.DateOnly)
	for _, day := range w.Daily {
		if day.Date.UTC().Format(time.DateOnly) == target {
			return day, true
		}
	}
	return Day{}, false
}

func WithinHourlyHorizon(now, moment time.Time) bool {
	delta := moment.UTC().Sub(now.UTC())
	return delta >= -time.Hour && delta <= HourlyHorizonHours*time.Hour
}

func WithinDailyHorizon(now, date time.Time) bool {
	delta := date.UTC().Sub(now.UTC().Truncate(24 * time.Hour))
	return delta >= -24*time.Hour && delta <= DailyHorizonDays*24*time.Hour
}
