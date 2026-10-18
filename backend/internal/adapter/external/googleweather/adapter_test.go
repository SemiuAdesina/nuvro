package googleweather_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleweather"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var observedAt = time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

func adapter(t *testing.T, status int, body string, seen *string) *googleweather.Adapter {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if seen != nil {
			*seen = request.URL.RawQuery
		}
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return googleweather.New(httpclient.New(httpclient.Options{BaseURL: server.URL, Timeout: time.Second}), "synthetic-test-key", clock.NewFixed(observedAt))
}

func TestCurrentConditionsAreNormalised(t *testing.T) {
	body := `{
      "currentTime":"2026-10-14T10:00:00Z",
      "weatherCondition":{"type":"LIGHT_RAIN_SHOWERS","description":{"text":"Light rain"}},
      "temperature":{"degrees":19.5,"unit":"CELSIUS"},
      "feelsLikeTemperature":{"degrees":18.0,"unit":"CELSIUS"},
      "relativeHumidity":78,
      "uvIndex":3,
      "precipitation":{"probability":{"percent":60,"type":"RAIN"},"qpf":{"value":1.4,"unit":"MILLIMETERS"}},
      "wind":{"speed":{"value":16.0,"unit":"KILOMETERS_PER_HOUR"}}
    }`
	observation, err := adapter(t, http.StatusOK, body, nil).Current(t.Context(), engine.Query{Location: funchal})
	if err != nil {
		t.Fatal(err)
	}
	if observation.Condition != weather.Rain {
		t.Fatalf("condition = %q, want rain", observation.Condition)
	}
	if observation.TemperatureCelsius != 19.5 || observation.HumidityPercent != 78 || observation.UVIndex != 3 {
		t.Fatalf("values lost: %+v", observation)
	}
	if observation.PrecipitationChance != 60 || observation.WindKilometresPerHour != 16 {
		t.Fatalf("precipitation or wind lost: %+v", observation)
	}
	if !observation.At.Equal(observedAt) {
		t.Fatalf("observation time = %v", observation.At)
	}
}

func TestImperialUnitsAreConvertedToCanonicalUnits(t *testing.T) {
	body := `{
      "currentTime":"2026-10-14T10:00:00Z",
      "weatherCondition":{"type":"CLEAR"},
      "temperature":{"degrees":68,"unit":"FAHRENHEIT"},
      "feelsLikeTemperature":{"degrees":68,"unit":"FAHRENHEIT"},
      "precipitation":{"probability":{"percent":0},"qpf":{"value":1,"unit":"INCHES"}},
      "wind":{"speed":{"value":10,"unit":"MILES_PER_HOUR"}}
    }`
	observation, err := adapter(t, http.StatusOK, body, nil).Current(t.Context(), engine.Query{Location: funchal})
	if err != nil {
		t.Fatal(err)
	}
	if observation.TemperatureCelsius < 19.9 || observation.TemperatureCelsius > 20.1 {
		t.Fatalf("fahrenheit not converted: %v", observation.TemperatureCelsius)
	}
	if observation.WindKilometresPerHour < 16 || observation.WindKilometresPerHour > 16.2 {
		t.Fatalf("miles per hour not converted: %v", observation.WindKilometresPerHour)
	}
	if observation.PrecipitationMillimetres < 25.3 || observation.PrecipitationMillimetres > 25.5 {
		t.Fatalf("inches not converted: %v", observation.PrecipitationMillimetres)
	}
}

func TestConditionVocabularyMapsToNuvroConditions(t *testing.T) {
	for providerType, want := range map[string]weather.Condition{
		"CLEAR": weather.Clear, "MOSTLY_CLOUDY": weather.Cloudy, "HEAVY_RAIN": weather.HeavyRain,
		"LIGHT_RAIN": weather.Rain, "THUNDERSTORM": weather.Storm, "SNOW_SHOWERS": weather.Snow,
		"FOG": weather.Fog, "WINDY": weather.Wind, "SOMETHING_NEW": weather.UnknownCondition, "": weather.UnknownCondition,
	} {
		body := `{"currentTime":"2026-10-14T10:00:00Z","weatherCondition":{"type":"` + providerType + `"},
                 "temperature":{"degrees":20,"unit":"CELSIUS"},"feelsLikeTemperature":{"degrees":20,"unit":"CELSIUS"}}`
		observation, err := adapter(t, http.StatusOK, body, nil).Current(t.Context(), engine.Query{Location: funchal})
		if providerType == "" || providerType == "SOMETHING_NEW" {
			if err == nil && observation.Condition != weather.UnknownCondition {
				t.Fatalf("%q became %q", providerType, observation.Condition)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if observation.Condition != want {
			t.Fatalf("%q became %q, want %q", providerType, observation.Condition, want)
		}
	}
}

func TestHourlyForecastSkipsUnusableEntries(t *testing.T) {
	body := `{"forecastHours":[
      {"interval":{"startTime":"2026-10-14T11:00:00Z"},"weatherCondition":{"type":"CLEAR"},
       "temperature":{"degrees":21,"unit":"CELSIUS"},"feelsLikeTemperature":{"degrees":21,"unit":"CELSIUS"}},
      {"interval":{"startTime":""},"weatherCondition":{"type":"CLEAR"},"temperature":{"degrees":21,"unit":"CELSIUS"}},
      {"interval":{"startTime":"2026-10-14T13:00:00Z"},"weatherCondition":{"type":"CLEAR"},
       "temperature":{"degrees":900,"unit":"CELSIUS"}},
      {"interval":{"startTime":"2026-10-14T14:00:00Z"},"weatherCondition":{"type":"RAIN"},
       "temperature":{"degrees":19,"unit":"KELVIN"}}
    ]}`
	hourly, err := adapter(t, http.StatusOK, body, nil).Hourly(t.Context(), engine.Query{Location: funchal, Hours: 12})
	if err != nil {
		t.Fatal(err)
	}
	if len(hourly) != 1 {
		t.Fatalf("kept %d hours, want only the usable one", len(hourly))
	}
	if !hourly[0].At.Equal(observedAt.Add(time.Hour)) {
		t.Fatalf("timestamp = %v", hourly[0].At)
	}
}

func TestDailyForecastCarriesHighLowAndSunEvents(t *testing.T) {
	body := `{"forecastDays":[
      {"displayDate":{"year":2026,"month":10,"day":14},
       "daytimeForecast":{"weatherCondition":{"type":"PARTLY_CLOUDY"},"precipitation":{"probability":{"percent":30}}},
       "maxTemperature":{"degrees":24,"unit":"CELSIUS"},"minTemperature":{"degrees":17,"unit":"CELSIUS"},
       "sunEvents":{"sunriseTime":"2026-10-14T06:30:00Z","sunsetTime":"2026-10-14T18:15:00Z"}},
      {"displayDate":{"year":0,"month":0,"day":0},"maxTemperature":{"degrees":20,"unit":"CELSIUS"}}
    ]}`
	daily, err := adapter(t, http.StatusOK, body, nil).Daily(t.Context(), engine.Query{Location: funchal, Days: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(daily) != 1 {
		t.Fatalf("kept %d days, want the dated one", len(daily))
	}
	day := daily[0]
	if day.Condition != weather.Cloudy || day.HighCelsius != 24 || day.LowCelsius != 17 {
		t.Fatalf("day = %+v", day)
	}
	if day.SunriseAt.IsZero() || day.SunsetAt.IsZero() || day.PrecipitationChance != 30 {
		t.Fatalf("sun events or precipitation lost: %+v", day)
	}
}

func TestRequestsAskForMetricUnitsAndBoundedCounts(t *testing.T) {
	var query string
	if _, err := adapter(t, http.StatusOK, `{"forecastDays":[]}`, &query).
		Daily(t.Context(), engine.Query{Location: funchal, Days: 7, Language: "en"}); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"unitsSystem=METRIC", "days=7", "languageCode=en", "location.latitude=32.6478"} {
		if !contains(query, expected) {
			t.Fatalf("query %q is missing %q", query, expected)
		}
	}
}

func TestProviderFailureIsNotSilentlySuccessful(t *testing.T) {
	if _, err := adapter(t, http.StatusTooManyRequests, "{}", nil).
		Current(t.Context(), engine.Query{Location: funchal}); err == nil {
		t.Fatal("a rate-limited response reported success")
	}
	if _, err := adapter(t, http.StatusOK, `{"forecastHours":[]}`, nil).
		Current(t.Context(), engine.Query{Location: funchal}); err == nil {
		t.Fatal("a response with no usable conditions reported success")
	}
}

func contains(haystack, needle string) bool {
	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}
