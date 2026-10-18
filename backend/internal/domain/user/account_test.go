package user_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

func TestNewAccountSeedsValidDefaults(t *testing.T) {
	account, err := user.NewAccount(identifier, "11112233-4455-6677-8899-aabbccddeeff", user.Identity{Provider: "apple", Subject: "001234.abc"})
	if err != nil {
		t.Fatal(err)
	}
	if account.Profile.UserID != identifier || account.Subject != "001234.abc" || account.Provider != "apple" {
		t.Fatalf("%+v", account)
	}
	if err := account.Profile.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := account.Preferences.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewAccountRequiresCompleteIdentity(t *testing.T) {
	for name, attempt := range map[string][3]string{
		"no user":     {"", "identity", "subject"},
		"no identity": {identifier, "", "subject"},
		"no subject":  {identifier, "identity", ""},
	} {
		if _, err := user.NewAccount(attempt[0], attempt[1], user.Identity{Provider: "apple", Subject: attempt[2]}); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if _, err := user.NewAccount(identifier, "identity", user.Identity{Subject: "001234.abc"}); err == nil {
		t.Fatal("an identity without a provider was accepted")
	}
}

func TestOnlyActiveAccountsMaySignIn(t *testing.T) {
	if !user.IsActive(user.StatusActive) {
		t.Fatal("active account refused")
	}
	for _, status := range []string{user.StatusSuspended, user.StatusDeleted, "", "unknown"} {
		if user.IsActive(status) {
			t.Fatalf("%q was treated as active", status)
		}
	}
}
