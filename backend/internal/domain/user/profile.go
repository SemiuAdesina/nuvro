package user

import (
	"errors"
	"regexp"
)

var ErrInvalidProfile = errors.New("user: invalid profile")
var ErrInvalidPreferences = errors.New("user: invalid preferences")

var (
	languagePattern = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
	countryPattern  = regexp.MustCompile(`^[A-Z]{2}$`)
	airportPattern  = regexp.MustCompile(`^[A-Z]{3}$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

const maximumTimezoneLength = 64
const maximumTagLength = 40

func (p Profile) Validate() error {
	if p.UserID == "" || !p.hasValidLocalization() || !p.hasValidHome() {
		return ErrInvalidProfile
	}
	if !isEnum(p.Units, "metric", "imperial") || !isEnum(p.HourCycle, "12", "24") {
		return ErrInvalidProfile
	}
	if !areTags(p.AccessibilityPreferences, 20) {
		return ErrInvalidProfile
	}
	return nil
}

func (p Profile) hasValidLocalization() bool {
	return languagePattern.MatchString(p.PreferredLanguage) &&
		languagePattern.MatchString(p.Locale) &&
		currencyPattern.MatchString(p.PreferredCurrency)
}

func (p Profile) hasValidHome() bool {
	if p.HomeCountry != nil && !countryPattern.MatchString(*p.HomeCountry) {
		return false
	}
	if p.DepartureAirport != nil && !airportPattern.MatchString(*p.DepartureAirport) {
		return false
	}
	if p.HomeTimezone != nil && (*p.HomeTimezone == "" || len(*p.HomeTimezone) > maximumTimezoneLength) {
		return false
	}
	return true
}

func (p Preferences) Validate() error {
	if !isEnum(p.Weather, "any", "warm", "mild", "cold") || !isEnum(p.Crowds, "any", "quiet", "lively") {
		return ErrInvalidPreferences
	}
	if !isEnum(p.Pace, "relaxed", "balanced", "packed") || !isEnum(p.Scope, "any", "domestic", "international") {
		return ErrInvalidPreferences
	}
	if !isEnum(p.EntrySimplicity, "any", "simple") {
		return ErrInvalidPreferences
	}
	if p.MaxFlightMinutes != nil && (*p.MaxFlightMinutes < 0 || *p.MaxFlightMinutes > 2880) {
		return ErrInvalidPreferences
	}
	if !areTags(p.TravelStyle, 20) || !areTags(p.Interests, 50) {
		return ErrInvalidPreferences
	}
	return nil
}

func isEnum(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func areTags(values []string, limit int) bool {
	if len(values) > limit {
		return false
	}
	for _, value := range values {
		if value == "" || len(value) > maximumTagLength {
			return false
		}
	}
	return true
}
