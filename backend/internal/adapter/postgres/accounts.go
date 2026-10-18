package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

func (s *Store) FindUserByIdentity(ctx context.Context, provider, subject string) (string, string, error) {
	row, err := s.queries(ctx).FindIdentity(ctx, queries.FindIdentityParams{Provider: provider, ProviderSubject: subject})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", databaseError(err)
	}
	return row.ID.String(), row.Status, nil
}

func (s *Store) CreateAccount(ctx context.Context, account user.Account) error {
	userID, err := identifier(account.UserID)
	if err != nil {
		return err
	}
	identityID, err := identifier(account.IdentityID)
	if err != nil {
		return err
	}
	store := s.queries(ctx)
	if err := store.InsertUser(ctx, userID); err != nil {
		return databaseError(err)
	}
	identity := queries.InsertIdentityParams{
		ID: identityID, UserID: userID, Provider: account.Provider, ProviderSubject: account.Subject,
	}
	if err := store.InsertIdentity(ctx, identity); err != nil {
		return databaseError(err)
	}
	if err := store.InsertProfile(ctx, insertProfileParams(userID, account.Profile)); err != nil {
		return databaseError(err)
	}
	return databaseError(store.InsertPreferences(ctx, insertPreferencesParams(userID, account.Preferences)))
}

func insertProfileParams(userID pgtypeUUID, profile user.Profile) queries.InsertProfileParams {
	return queries.InsertProfileParams{
		UserID: userID, PreferredLanguage: profile.PreferredLanguage, Locale: profile.Locale,
		HomeCountry: profile.HomeCountry, HomeTimezone: profile.HomeTimezone, DepartureAirport: profile.DepartureAirport,
		PreferredCurrency: profile.PreferredCurrency, Units: profile.Units, HourCycle: profile.HourCycle,
		AccessibilityPreferences: tags(profile.AccessibilityPreferences),
	}
}

func insertPreferencesParams(userID pgtypeUUID, preferences user.Preferences) queries.InsertPreferencesParams {
	return queries.InsertPreferencesParams{
		UserID: userID, TravelStyle: tags(preferences.TravelStyle), WeatherPreference: preferences.Weather,
		CrowdPreference: preferences.Crowds, MaxFlightMinutes: preferences.MaxFlightMinutes, TripPace: preferences.Pace,
		Interests: tags(preferences.Interests), TravelingWithChildren: preferences.TravelingWithChildren,
		Scope: preferences.Scope, EntrySimplicity: preferences.EntrySimplicity,
	}
}
