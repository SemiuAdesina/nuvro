package user_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

const identifier = "00112233-4455-6677-8899-aabbccddeeff"

func TestDefaultsAreValid(t *testing.T) {
	if err := user.DefaultProfile(identifier).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := user.DefaultPreferences().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProfileValidationRejectsMalformedValues(t *testing.T) {
	country := "usa"
	airport := "heathrow"
	empty := ""
	for name, mutate := range map[string]func(*user.Profile){
		"no user":        func(p *user.Profile) { p.UserID = "" },
		"bad language":   func(p *user.Profile) { p.PreferredLanguage = "English" },
		"bad locale":     func(p *user.Profile) { p.Locale = "EN_us" },
		"bad currency":   func(p *user.Profile) { p.PreferredCurrency = "usd" },
		"bad country":    func(p *user.Profile) { p.HomeCountry = &country },
		"bad airport":    func(p *user.Profile) { p.DepartureAirport = &airport },
		"empty timezone": func(p *user.Profile) { p.HomeTimezone = &empty },
		"bad units":      func(p *user.Profile) { p.Units = "furlongs" },
		"bad hours":      func(p *user.Profile) { p.HourCycle = "25" },
		"empty tag":      func(p *user.Profile) { p.AccessibilityPreferences = []string{""} },
	} {
		profile := user.DefaultProfile(identifier)
		mutate(&profile)
		if err := profile.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestPreferencesValidationRejectsMalformedValues(t *testing.T) {
	tooLong := int32(3000)
	negative := int32(-1)
	for name, mutate := range map[string]func(*user.Preferences){
		"bad weather":  func(p *user.Preferences) { p.Weather = "tropical" },
		"bad crowds":   func(p *user.Preferences) { p.Crowds = "busy" },
		"bad pace":     func(p *user.Preferences) { p.Pace = "frantic" },
		"bad scope":    func(p *user.Preferences) { p.Scope = "galactic" },
		"bad entry":    func(p *user.Preferences) { p.EntrySimplicity = "easy" },
		"long flight":  func(p *user.Preferences) { p.MaxFlightMinutes = &tooLong },
		"negative":     func(p *user.Preferences) { p.MaxFlightMinutes = &negative },
		"too many":     func(p *user.Preferences) { p.Interests = make([]string, 51) },
		"empty string": func(p *user.Preferences) { p.TravelStyle = []string{""} },
	} {
		preferences := user.DefaultPreferences()
		mutate(&preferences)
		if err := preferences.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestProfilePatchHonoursThreeStates(t *testing.T) {
	country := "PT"
	profile := user.DefaultProfile(identifier)
	profile.HomeCountry = &country
	unchanged, err := profile.Apply(user.ProfilePatch{})
	if err != nil || unchanged.HomeCountry == nil || *unchanged.HomeCountry != "PT" || unchanged.Units != "metric" {
		t.Fatalf("omitted fields changed: %+v %v", unchanged, err)
	}
	cleared, err := profile.Apply(user.ProfilePatch{HomeCountry: change.Clear[string]()})
	if err != nil || cleared.HomeCountry != nil {
		t.Fatalf("clear failed: %v", err)
	}
	assigned, err := profile.Apply(user.ProfilePatch{PreferredCurrency: change.Set("NGN"), Units: change.Set("imperial")})
	if err != nil || assigned.PreferredCurrency != "NGN" || assigned.Units != "imperial" {
		t.Fatalf("assignment failed: %v", err)
	}
	if _, err := profile.Apply(user.ProfilePatch{PreferredCurrency: change.Clear[string]()}); err == nil {
		t.Fatal("required currency was cleared")
	}
	if _, err := profile.Apply(user.ProfilePatch{Units: change.Set("parsecs")}); err == nil {
		t.Fatal("invalid unit accepted")
	}
}

func TestPreferencesPatchHonoursThreeStates(t *testing.T) {
	minutes := int32(600)
	preferences := user.DefaultPreferences()
	preferences.MaxFlightMinutes = &minutes
	unchanged, err := preferences.Apply(user.PreferencesPatch{})
	if err != nil || unchanged.MaxFlightMinutes == nil || *unchanged.MaxFlightMinutes != 600 {
		t.Fatalf("omitted field changed: %v", err)
	}
	cleared, err := preferences.Apply(user.PreferencesPatch{MaxFlightMinutes: change.Clear[int32]()})
	if err != nil || cleared.MaxFlightMinutes != nil {
		t.Fatalf("clear failed: %v", err)
	}
	assigned, err := preferences.Apply(user.PreferencesPatch{Interests: change.Set([]string{"food"}), Pace: change.Set("relaxed")})
	if err != nil || len(assigned.Interests) != 1 || assigned.Pace != "relaxed" {
		t.Fatalf("assignment failed: %v", err)
	}
	if _, err := preferences.Apply(user.PreferencesPatch{Pace: change.Clear[string]()}); err == nil {
		t.Fatal("required pace was cleared")
	}
}
