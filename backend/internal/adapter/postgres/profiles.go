package postgres

import (
	"context"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

func (s *Store) Profile(ctx context.Context, userID string) (user.Profile, error) {
	value, err := identifier(userID)
	if err != nil {
		return user.Profile{}, err
	}
	row, err := s.queries(ctx).GetProfile(ctx, value)
	if err != nil {
		return user.Profile{}, databaseError(err)
	}
	return user.Profile{
		UserID: row.UserID.String(), PreferredLanguage: row.PreferredLanguage, Locale: row.Locale,
		HomeCountry: row.HomeCountry, HomeTimezone: row.HomeTimezone, DepartureAirport: row.DepartureAirport,
		PreferredCurrency: row.PreferredCurrency, Units: row.Units, HourCycle: row.HourCycle,
		AccessibilityPreferences: tags(row.AccessibilityPreferences),
	}, nil
}

func (s *Store) SaveProfile(ctx context.Context, profile user.Profile) error {
	value, err := identifier(profile.UserID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdateProfile(ctx, queries.UpdateProfileParams{
		UserID: value, PreferredLanguage: profile.PreferredLanguage, Locale: profile.Locale,
		HomeCountry: profile.HomeCountry, HomeTimezone: profile.HomeTimezone, DepartureAirport: profile.DepartureAirport,
		PreferredCurrency: profile.PreferredCurrency, Units: profile.Units, HourCycle: profile.HourCycle,
		AccessibilityPreferences: tags(profile.AccessibilityPreferences),
	}))
}

func (s *Store) Preferences(ctx context.Context, userID string) (user.Preferences, error) {
	value, err := identifier(userID)
	if err != nil {
		return user.Preferences{}, err
	}
	row, err := s.queries(ctx).GetPreferences(ctx, value)
	if err != nil {
		return user.Preferences{}, databaseError(err)
	}
	return user.Preferences{
		TravelStyle: tags(row.TravelStyle), Weather: row.WeatherPreference, Crowds: row.CrowdPreference,
		MaxFlightMinutes: row.MaxFlightMinutes, Pace: row.TripPace, Interests: tags(row.Interests),
		TravelingWithChildren: row.TravelingWithChildren, Scope: row.Scope, EntrySimplicity: row.EntrySimplicity,
	}, nil
}

func (s *Store) SavePreferences(ctx context.Context, userID string, preferences user.Preferences) error {
	value, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdatePreferences(ctx, queries.UpdatePreferencesParams{
		UserID: value, TravelStyle: tags(preferences.TravelStyle), WeatherPreference: preferences.Weather,
		CrowdPreference: preferences.Crowds, MaxFlightMinutes: preferences.MaxFlightMinutes, TripPace: preferences.Pace,
		Interests: tags(preferences.Interests), TravelingWithChildren: preferences.TravelingWithChildren,
		Scope: preferences.Scope, EntrySimplicity: preferences.EntrySimplicity,
	}))
}
