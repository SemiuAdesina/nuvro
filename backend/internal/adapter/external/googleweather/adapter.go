package googleweather

import (
	"context"
	"strconv"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultBaseURL = "https://weather.googleapis.com"
const Provider = "google_weather"

type Adapter struct {
	client *httpclient.Client
	apiKey string
	timer  clock.Clock
}

func New(client *httpclient.Client, apiKey string, timer clock.Clock) *Adapter {
	return &Adapter{client: client, apiKey: apiKey, timer: timer}
}

func (a *Adapter) Current(ctx context.Context, query engine.Query) (weather.Observation, error) {
	var payload conditionsPayload
	if err := a.client.Get(ctx, httpclient.Request{Path: a.path("/v1/currentConditions:lookup", query, 0)}, &payload); err != nil {
		return weather.Observation{}, err
	}
	if payload.WeatherCondition.Type == "" && payload.Temperature.Unit == "" {
		return weather.Observation{}, httpclient.ErrMalformed
	}
	at, ok := instant(payload.CurrentTime)
	if !ok {
		at = a.timer.Now().UTC()
	}
	observation, ok := toObservation(hourPayload{
		Interval: struct {
			StartTime string `json:"startTime"`
		}{StartTime: at.Format("2006-01-02T15:04:05Z")},
		WeatherCondition:     payload.WeatherCondition,
		Temperature:          payload.Temperature,
		FeelsLikeTemperature: payload.FeelsLikeTemperature,
		RelativeHumidity:     payload.RelativeHumidity,
		UVIndex:              payload.UVIndex,
		Precipitation:        payload.Precipitation,
		Wind:                 payload.Wind,
	})
	if !ok {
		return weather.Observation{}, httpclient.ErrMalformed
	}
	return observation, nil
}

func (a *Adapter) Hourly(ctx context.Context, query engine.Query) ([]weather.Observation, error) {
	var response hourlyResponse
	if err := a.client.Get(ctx, httpclient.Request{Path: a.path("/v1/forecast/hours:lookup", query, query.Hours)}, &response); err != nil {
		return nil, err
	}
	observations := make([]weather.Observation, 0, len(response.ForecastHours))
	for _, payload := range response.ForecastHours {
		observation, ok := toObservation(payload)
		if !ok {
			continue
		}
		observations = append(observations, observation)
	}
	return observations, nil
}

func (a *Adapter) Daily(ctx context.Context, query engine.Query) ([]weather.Day, error) {
	var response dailyResponse
	if err := a.client.Get(ctx, httpclient.Request{Path: a.path("/v1/forecast/days:lookup", query, query.Days)}, &response); err != nil {
		return nil, err
	}
	days := make([]weather.Day, 0, len(response.ForecastDays))
	for _, payload := range response.ForecastDays {
		day, ok := toDay(payload)
		if !ok {
			continue
		}
		days = append(days, day)
	}
	return days, nil
}

func (a *Adapter) path(base string, query engine.Query, count int) string {
	path := base + "?key=" + a.apiKey +
		"&location.latitude=" + strconv.FormatFloat(query.Location.Latitude, 'f', 6, 64) +
		"&location.longitude=" + strconv.FormatFloat(query.Location.Longitude, 'f', 6, 64) +
		"&unitsSystem=METRIC"
	if query.Language != "" {
		path += "&languageCode=" + query.Language
	}
	if count > 0 {
		path += "&" + countParameter(base) + "=" + strconv.Itoa(count)
	}
	return path
}

func countParameter(base string) string {
	if base == "/v1/forecast/days:lookup" {
		return "days"
	}
	return "hours"
}

var _ engine.Provider = (*Adapter)(nil)
