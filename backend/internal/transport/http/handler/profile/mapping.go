package profile

import (
	"math"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func optional[T any](value nullable.Nullable[T]) change.Change[T] {
	if !value.IsSpecified() {
		return change.Unchanged[T]()
	}
	if value.IsNull() {
		return change.Clear[T]()
	}
	assigned, err := value.Get()
	if err != nil {
		return change.Unchanged[T]()
	}
	return change.Set(assigned)
}

func present[T any](value *T) change.Change[T] {
	if value == nil {
		return change.Unchanged[T]()
	}
	return change.Set(*value)
}

func enum[T ~string](value *T) change.Change[string] {
	if value == nil {
		return change.Unchanged[string]()
	}
	return change.Set(string(*value))
}

func profileResponse(profile user.Profile) api.Profile {
	identifier, _ := uuid.Parse(profile.UserID)
	response := api.Profile{
		UserId:                   identifier,
		PreferredLanguage:        profile.PreferredLanguage,
		Locale:                   profile.Locale,
		PreferredCurrency:        profile.PreferredCurrency,
		Units:                    api.ProfileUnits(profile.Units),
		HourCycle:                api.ProfileHourCycle(profile.HourCycle),
		AccessibilityPreferences: profile.AccessibilityPreferences,
		HomeCountry:              nullableFrom(profile.HomeCountry),
		HomeTimezone:             nullableFrom(profile.HomeTimezone),
		DepartureAirport:         nullableFrom(profile.DepartureAirport),
	}
	if response.AccessibilityPreferences == nil {
		response.AccessibilityPreferences = []string{}
	}
	return response
}

func preferencesResponse(preferences user.Preferences) api.Preferences {
	response := api.Preferences{
		TravelStyle:           preferences.TravelStyle,
		WeatherPreference:     api.PreferencesWeatherPreference(preferences.Weather),
		CrowdPreference:       api.PreferencesCrowdPreference(preferences.Crowds),
		TripPace:              api.PreferencesTripPace(preferences.Pace),
		Interests:             preferences.Interests,
		TravelingWithChildren: preferences.TravelingWithChildren,
		Scope:                 api.PreferencesScope(preferences.Scope),
		EntrySimplicity:       api.PreferencesEntrySimplicity(preferences.EntrySimplicity),
		MaxFlightMinutes:      nullable.NewNullNullable[int](),
	}
	if preferences.MaxFlightMinutes != nil {
		response.MaxFlightMinutes = nullable.NewNullableWithValue(int(*preferences.MaxFlightMinutes))
	}
	if response.TravelStyle == nil {
		response.TravelStyle = []string{}
	}
	if response.Interests == nil {
		response.Interests = []string{}
	}
	return response
}

func nullableFrom(value *string) nullable.Nullable[string] {
	if value == nil {
		return nullable.NewNullNullable[string]()
	}
	return nullable.NewNullableWithValue(*value)
}

func profilePatch(body api.ProfilePatch) user.ProfilePatch {
	return user.ProfilePatch{
		PreferredLanguage:        present(body.PreferredLanguage),
		Locale:                   present(body.Locale),
		PreferredCurrency:        present(body.PreferredCurrency),
		HomeCountry:              optional(body.HomeCountry),
		HomeTimezone:             optional(body.HomeTimezone),
		DepartureAirport:         optional(body.DepartureAirport),
		Units:                    enum(body.Units),
		HourCycle:                enum(body.HourCycle),
		AccessibilityPreferences: present(body.AccessibilityPreferences),
	}
}

func preferencesPatch(body api.PreferencesPatch) (user.PreferencesPatch, bool) {
	patch := user.PreferencesPatch{
		TravelStyle:           present(body.TravelStyle),
		Weather:               enum(body.WeatherPreference),
		Crowds:                enum(body.CrowdPreference),
		Pace:                  enum(body.TripPace),
		Interests:             present(body.Interests),
		TravelingWithChildren: present(body.TravelingWithChildren),
		Scope:                 enum(body.Scope),
		EntrySimplicity:       enum(body.EntrySimplicity),
		MaxFlightMinutes:      change.Unchanged[int32](),
	}
	minutes := optional(body.MaxFlightMinutes)
	if value, ok := minutes.Value(); ok {
		if value < math.MinInt32 || value > math.MaxInt32 {
			return user.PreferencesPatch{}, false
		}
		patch.MaxFlightMinutes = change.Set(int32(value))
	} else if minutes.IsCleared() {
		patch.MaxFlightMinutes = change.Clear[int32]()
	}
	return patch, true
}
