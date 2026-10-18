package user

import "github.com/SemiuAdesina/nuvro/backend/internal/domain/change"

type ProfilePatch struct {
	PreferredLanguage        change.Change[string]
	Locale                   change.Change[string]
	PreferredCurrency        change.Change[string]
	HomeCountry              change.Change[string]
	HomeTimezone             change.Change[string]
	DepartureAirport         change.Change[string]
	Units                    change.Change[string]
	HourCycle                change.Change[string]
	AccessibilityPreferences change.Change[[]string]
}

type PreferencesPatch struct {
	TravelStyle           change.Change[[]string]
	Weather               change.Change[string]
	Crowds                change.Change[string]
	MaxFlightMinutes      change.Change[int32]
	Pace                  change.Change[string]
	Interests             change.Change[[]string]
	TravelingWithChildren change.Change[bool]
	Scope                 change.Change[string]
	EntrySimplicity       change.Change[string]
}

func (p Profile) Apply(patch ProfilePatch) (Profile, error) {
	updated := p
	patch.HomeCountry.ApplyToOptional(&updated.HomeCountry)
	patch.HomeTimezone.ApplyToOptional(&updated.HomeTimezone)
	patch.DepartureAirport.ApplyToOptional(&updated.DepartureAirport)
	required := []error{
		patch.PreferredLanguage.ApplyTo(&updated.PreferredLanguage),
		patch.Locale.ApplyTo(&updated.Locale),
		patch.PreferredCurrency.ApplyTo(&updated.PreferredCurrency),
		patch.Units.ApplyTo(&updated.Units),
		patch.HourCycle.ApplyTo(&updated.HourCycle),
		patch.AccessibilityPreferences.ApplyTo(&updated.AccessibilityPreferences),
	}
	for _, err := range required {
		if err != nil {
			return Profile{}, ErrInvalidProfile
		}
	}
	if err := updated.Validate(); err != nil {
		return Profile{}, err
	}
	return updated, nil
}

func (p Preferences) Apply(patch PreferencesPatch) (Preferences, error) {
	updated := p
	patch.MaxFlightMinutes.ApplyToOptional(&updated.MaxFlightMinutes)
	required := []error{
		patch.TravelStyle.ApplyTo(&updated.TravelStyle),
		patch.Weather.ApplyTo(&updated.Weather),
		patch.Crowds.ApplyTo(&updated.Crowds),
		patch.Pace.ApplyTo(&updated.Pace),
		patch.Interests.ApplyTo(&updated.Interests),
		patch.TravelingWithChildren.ApplyTo(&updated.TravelingWithChildren),
		patch.Scope.ApplyTo(&updated.Scope),
		patch.EntrySimplicity.ApplyTo(&updated.EntrySimplicity),
	}
	for _, err := range required {
		if err != nil {
			return Preferences{}, ErrInvalidPreferences
		}
	}
	if err := updated.Validate(); err != nil {
		return Preferences{}, err
	}
	return updated, nil
}

func DefaultProfile(userID string) Profile {
	return Profile{
		UserID: userID, PreferredLanguage: "en", Locale: "en", PreferredCurrency: "USD",
		Units: "metric", HourCycle: "24", AccessibilityPreferences: []string{},
	}
}

func DefaultPreferences() Preferences {
	return Preferences{
		TravelStyle: []string{}, Weather: "any", Crowds: "any", Pace: "balanced",
		Interests: []string{}, Scope: "any", EntrySimplicity: "any",
	}
}
