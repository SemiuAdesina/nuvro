//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/identity"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

type appleStub struct {
	refreshToken string
	revoked      []string
	failing      bool
}

func (a *appleStub) Exchange(context.Context, string) (string, error) { return a.refreshToken, nil }

func (a *appleStub) Revoke(_ context.Context, token string) error {
	if a.failing {
		return apple.ErrRevokeFailed
	}
	a.revoked = append(a.revoked, token)
	return nil
}

func linkerFor(t *testing.T, store *postgres.Store, stub *appleStub) *account.Linker {
	t.Helper()
	sealer, err := secret.NewSealer(bytes.Repeat([]byte{'k'}, secret.KeyLength))
	if err != nil {
		t.Fatal(err)
	}
	return account.NewLinker(stub, store, sealer, apple.ProviderName, clock.NewFixed(time.Now().UTC()))
}

func TestALinkedRefreshTokenIsStoredSealedAndNeverInTheClear(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	stub := &appleStub{refreshToken: "apple-refresh-token"}

	if err := linkerFor(t, store, stub).Link(t.Context(), principal, "authorization-code"); err != nil {
		t.Fatal(err)
	}
	var stored []byte
	err := database.QueryRowContext(t.Context(),
		"SELECT sealed_refresh_token FROM user_identities WHERE user_id = $1", principal.UserID).Scan(&stored)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) == 0 {
		t.Fatal("no sealed token was stored")
	}
	if bytes.Contains(stored, []byte("apple-refresh-token")) {
		t.Fatal("the refresh token is readable in the database")
	}
}

func TestClosingAnAccountRevokesItsSignInTokens(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	stub := &appleStub{refreshToken: "apple-refresh-token"}
	linker := linkerFor(t, store, stub)
	if err := linker.Link(t.Context(), principal, "authorization-code"); err != nil {
		t.Fatal(err)
	}

	eraser := account.NewEraser(store, generatorFor(t), clock.NewFixed(time.Now().UTC())).WithLinker(linker)
	if _, err := eraser.Request(t.Context(), principal, "no_longer_travelling"); err != nil {
		t.Fatal(err)
	}
	if len(stub.revoked) != 1 || stub.revoked[0] != "apple-refresh-token" {
		t.Fatalf("provider calls = %v, want the stored token revoked", stub.revoked)
	}
	var sealed []byte
	var revokedAt *time.Time
	err := database.QueryRowContext(t.Context(),
		"SELECT sealed_refresh_token, revoked_at FROM user_identities WHERE user_id = $1",
		principal.UserID).Scan(&sealed, &revokedAt)
	if err != nil {
		t.Fatal(err)
	}
	if sealed != nil || revokedAt == nil {
		t.Fatalf("the identity still holds a token (%v) or was never marked revoked (%v)", sealed, revokedAt)
	}
}

func TestAProviderOutageLeavesTheTokenForTheSweepToRetry(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	failing := &appleStub{refreshToken: "apple-refresh-token", failing: true}
	linker := linkerFor(t, store, failing)
	if err := linker.Link(t.Context(), principal, "authorization-code"); err != nil {
		t.Fatal(err)
	}
	eraser := account.NewEraser(store, generatorFor(t), clock.NewFixed(time.Now().UTC())).WithLinker(linker)
	if _, err := eraser.Request(t.Context(), principal, ""); err != nil {
		t.Fatalf("a provider outage blocked account closure: %v", err)
	}

	pending, err := store.RevocableIdentities(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if !holdsIdentityFor(pending, principal) {
		t.Fatal("the unrevoked identity was not queued for the next sweep")
	}

	recovered := &appleStub{refreshToken: "apple-refresh-token"}
	revoked, err := linkerFor(t, store, recovered).RevokePending(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if revoked == 0 || len(recovered.revoked) == 0 {
		t.Fatal("the sweep did not revoke the identity once the provider recovered")
	}
}

func holdsIdentityFor(pending []identity.Sealed, principal user.Principal) bool {
	for _, sealed := range pending {
		if sealed.UserID == principal.UserID {
			return true
		}
	}
	return false
}

func generatorFor(t *testing.T) account.Identifiers {
	t.Helper()
	return id.NewUUID()
}
