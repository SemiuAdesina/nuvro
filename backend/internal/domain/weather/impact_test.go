package weather_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

var noon = time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

func observation(condition weather.Condition, celsius float64) weather.Observation {
	return weather.Observation{At: noon, Condition: condition, TemperatureCelsius: celsius}
}

func TestSensitivityFollowsCategory(t *testing.T) {
	for category, want := range map[place.Category]weather.Sensitivity{
		place.Beach: weather.Outdoor, place.Park: weather.Outdoor, place.Viewpoint: weather.Outdoor,
		place.Museum: weather.Indoor, place.Restaurant: weather.Indoor, place.Shopping: weather.Indoor,
		place.Experience: weather.Sheltered, place.Transport: weather.Sheltered,
	} {
		if got := weather.SensitivityOf(category); got != want {
			t.Fatalf("%s sensitivity = %q, want %q", category, got, want)
		}
	}
}

func TestOutdoorActivitiesAreBlockedByStormsButNotByClearSkies(t *testing.T) {
	if impact := weather.Assess(observation(weather.Clear, 24), weather.Outdoor); impact != weather.NoImpact {
		t.Fatalf("clear weather outdoors = %q", impact)
	}
	if impact := weather.Assess(observation(weather.Storm, 20), weather.Outdoor); !impact.Blocks() {
		t.Fatalf("a storm did not block an outdoor activity: %q", impact)
	}
	if impact := weather.Assess(observation(weather.HeavyRain, 18), weather.Outdoor); !impact.Blocks() {
		t.Fatalf("heavy rain did not block an outdoor activity: %q", impact)
	}
}

func TestMildRainBarelyAffectsIndoorPlans(t *testing.T) {
	indoor := weather.Assess(observation(weather.Rain, 18), weather.Indoor)
	outdoor := weather.Assess(observation(weather.Rain, 18), weather.Outdoor)
	if indoor != weather.NoImpact {
		t.Fatalf("mild rain indoors = %q, want none", indoor)
	}
	if outdoor != weather.ModerateImpact {
		t.Fatalf("mild rain outdoors = %q, want moderate", outdoor)
	}
}

func TestExtremeTemperatureAndWindRaiseImpactWithoutRain(t *testing.T) {
	severe := weather.Assess(observation(weather.Clear, 41), weather.Outdoor)
	if severe != weather.HighImpact {
		t.Fatalf("severe heat outdoors = %q, want high", severe)
	}
	warm := weather.Assess(observation(weather.Clear, 36), weather.Outdoor)
	if warm != weather.ModerateImpact {
		t.Fatalf("high heat outdoors = %q, want moderate", warm)
	}
	windy := weather.Observation{At: noon, Condition: weather.Clear, TemperatureCelsius: 20, WindKilometresPerHour: 70}
	if impact := weather.Assess(windy, weather.Outdoor); impact != weather.HighImpact {
		t.Fatalf("strong wind outdoors = %q, want high", impact)
	}
}

func TestUnknownConditionsNeverClaimAnImpact(t *testing.T) {
	if impact := weather.Assess(observation(weather.UnknownCondition, 20), weather.Outdoor); impact != weather.UnknownImpact {
		t.Fatalf("unknown condition = %q", impact)
	}
	if impact := weather.Assess(weather.Observation{At: noon, Condition: "sunshowers"}, weather.Outdoor); impact != weather.UnknownImpact {
		t.Fatal("an invented condition produced an impact")
	}
}

func TestImpactIsDeterministic(t *testing.T) {
	subject := observation(weather.Rain, 19)
	first := weather.Assess(subject, weather.Outdoor)
	for range 50 {
		if weather.Assess(subject, weather.Outdoor) != first {
			t.Fatal("impact assessment is not deterministic")
		}
	}
}

func TestForecastHorizonsAreEnforced(t *testing.T) {
	now := noon
	if !weather.WithinHourlyHorizon(now, now.Add(200*time.Hour)) {
		t.Fatal("a time inside the hourly horizon was rejected")
	}
	if weather.WithinHourlyHorizon(now, now.Add(300*time.Hour)) {
		t.Fatal("a time beyond the hourly horizon was accepted")
	}
	if !weather.WithinDailyHorizon(now, now.AddDate(0, 0, 9)) {
		t.Fatal("a date inside the daily horizon was rejected")
	}
	if weather.WithinDailyHorizon(now, now.AddDate(0, 6, 0)) {
		t.Fatal("a date six months out was treated as forecastable")
	}
}

func TestWindowLookupsAndValidation(t *testing.T) {
	window := weather.Window{
		Availability: weather.Available,
		Hourly:       []weather.Observation{observation(weather.Rain, 18)},
		Daily:        []weather.Day{{Date: noon, Condition: weather.Clear, HighCelsius: 25, LowCelsius: 17}},
	}
	if _, found := window.HourAt(noon.Add(20 * time.Minute)); !found {
		t.Fatal("hour lookup failed inside the same hour")
	}
	if _, found := window.HourAt(noon.Add(5 * time.Hour)); found {
		t.Fatal("hour lookup returned an observation that was never provided")
	}
	if _, found := window.DayOn(noon); !found {
		t.Fatal("day lookup failed")
	}
	unavailable := weather.Unavailable(weather.BeyondForecast)
	if unavailable.Known() {
		t.Fatal("an unavailable window claimed knowledge")
	}
	for name, broken := range map[string]weather.Observation{
		"no time":       {Condition: weather.Clear},
		"absurd heat":   {At: noon, Condition: weather.Clear, TemperatureCelsius: 200},
		"bad chance":    {At: noon, Condition: weather.Clear, PrecipitationChance: 150},
		"negative wind": {At: noon, Condition: weather.Clear, WindKilometresPerHour: -5},
	} {
		if broken.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}
