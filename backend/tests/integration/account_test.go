//go:build integration

package integration_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/profile"
)

func TestEnrolmentAndProfilePersistence(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	identity := user.Identity{Provider: "apple", Subject: "001234." + randomSuffix(t)}
	resolver := account.NewResolver(store, store, id.NewUUID())

	principal, err := resolver.Resolve(t.Context(), identity)
	if err != nil || principal.UserID == "" {
		t.Fatalf("%+v %v", principal, err)
	}
	again, err := resolver.Resolve(t.Context(), identity)
	if err != nil || again.UserID != principal.UserID {
		t.Fatalf("a second sign-in created a second account: %+v %v", again, err)
	}

	service := profile.NewService(store, store)
	stored, err := service.Profile(t.Context(), principal)
	if err != nil || stored.UserID != principal.UserID || stored.PreferredCurrency != "USD" {
		t.Fatalf("%+v %v", stored, err)
	}
	updated, err := service.UpdateProfile(t.Context(), principal, user.ProfilePatch{
		PreferredCurrency: change.Set("NGN"), HomeCountry: change.Set("NG"), DepartureAirport: change.Set("LOS"),
	})
	if err != nil || updated.PreferredCurrency != "NGN" || *updated.HomeCountry != "NG" {
		t.Fatalf("%+v %v", updated, err)
	}
	reread, err := service.Profile(t.Context(), principal)
	if err != nil || reread.PreferredCurrency != "NGN" || *reread.DepartureAirport != "LOS" || reread.Locale != "en" {
		t.Fatalf("the update did not persist: %+v %v", reread, err)
	}
	cleared, err := service.UpdateProfile(t.Context(), principal, user.ProfilePatch{HomeCountry: change.Clear[string]()})
	if err != nil || cleared.HomeCountry != nil {
		t.Fatalf("clear failed: %+v %v", cleared, err)
	}

	minutes := int32(420)
	preferences, err := service.UpdatePreferences(t.Context(), principal, user.PreferencesPatch{
		Interests: change.Set([]string{"food", "nature"}), MaxFlightMinutes: change.Set(minutes), Pace: change.Set("relaxed"),
	})
	if err != nil || len(preferences.Interests) != 2 || *preferences.MaxFlightMinutes != minutes {
		t.Fatalf("%+v %v", preferences, err)
	}
	storedPreferences, err := service.Preferences(t.Context(), principal)
	if err != nil || storedPreferences.Pace != "relaxed" || len(storedPreferences.Interests) != 2 {
		t.Fatalf("preferences did not persist: %+v %v", storedPreferences, err)
	}
	emptied, err := service.UpdatePreferences(t.Context(), principal, user.PreferencesPatch{MaxFlightMinutes: change.Clear[int32]()})
	if err != nil || emptied.MaxFlightMinutes != nil {
		t.Fatalf("clearing the flight tolerance failed: %v", err)
	}
}
