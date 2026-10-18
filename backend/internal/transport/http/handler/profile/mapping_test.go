package profile

import (
	"testing"

	"github.com/oapi-codegen/nullable"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func TestProfilePatchMapsThreeStates(t *testing.T) {
	patch := profilePatch(api.ProfilePatch{})
	if !patch.HomeCountry.IsUnchanged() || !patch.PreferredCurrency.IsUnchanged() || !patch.Units.IsUnchanged() {
		t.Fatal("an empty body was not neutral")
	}
	patch = profilePatch(api.ProfilePatch{HomeCountry: nullable.NewNullNullable[string]()})
	if !patch.HomeCountry.IsCleared() {
		t.Fatal("explicit null did not clear the country")
	}
	currency := "NGN"
	units := api.ProfilePatchUnits("imperial")
	patch = profilePatch(api.ProfilePatch{PreferredCurrency: &currency, Units: &units})
	value, present := patch.PreferredCurrency.Value()
	if !present || value != "NGN" {
		t.Fatal("currency assignment lost")
	}
	if unit, present := patch.Units.Value(); !present || unit != "imperial" {
		t.Fatal("enum assignment lost")
	}
}

func TestPreferencesPatchMapsNullableFlightTolerance(t *testing.T) {
	patch, _ := preferencesPatch(api.PreferencesPatch{})
	if !patch.MaxFlightMinutes.IsUnchanged() || !patch.Interests.IsUnchanged() {
		t.Fatal("an empty body was not neutral")
	}
	patch, _ = preferencesPatch(api.PreferencesPatch{MaxFlightMinutes: nullable.NewNullNullable[int]()})
	if !patch.MaxFlightMinutes.IsCleared() {
		t.Fatal("explicit null did not clear the tolerance")
	}
	patch, _ = preferencesPatch(api.PreferencesPatch{MaxFlightMinutes: nullable.NewNullableWithValue(600)})
	if value, present := patch.MaxFlightMinutes.Value(); !present || value != 600 {
		t.Fatal("tolerance assignment lost")
	}
}

func TestResponsesAlwaysCarryCollectionsAndNulls(t *testing.T) {
	response := profileResponse(user.DefaultProfile("3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"))
	if !response.HomeCountry.IsNull() || !response.DepartureAirport.IsNull() || !response.HomeTimezone.IsNull() {
		t.Fatal("absent profile values were not rendered as null")
	}
	if response.AccessibilityPreferences == nil {
		t.Fatal("an empty collection became null")
	}
	preferences := preferencesResponse(user.DefaultPreferences())
	if !preferences.MaxFlightMinutes.IsNull() || preferences.Interests == nil || preferences.TravelStyle == nil {
		t.Fatalf("preferences rendered badly: %+v", preferences)
	}
	minutes := int32(600)
	valued := user.DefaultPreferences()
	valued.MaxFlightMinutes = &minutes
	rendered := preferencesResponse(valued)
	if value, err := rendered.MaxFlightMinutes.Get(); err != nil || value != 600 {
		t.Fatal("tolerance was not rendered")
	}
}
