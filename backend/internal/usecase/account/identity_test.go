package account_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/identity"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

const appleProvider = "apple"

type provider struct {
	refreshToken   string
	exchangeFailed bool
	revokeFailed   bool
	revoked        []string
	exchanged      []string
}

func (p *provider) Exchange(_ context.Context, code string) (string, error) {
	p.exchanged = append(p.exchanged, code)
	if p.exchangeFailed {
		return "", errors.New("rejected")
	}
	return p.refreshToken, nil
}

func (p *provider) Revoke(_ context.Context, token string) error {
	if p.revokeFailed {
		return errors.New("unavailable")
	}
	p.revoked = append(p.revoked, token)
	return nil
}

type identities struct {
	linked   map[string][]byte
	held     []identity.Sealed
	revoked  []string
	failure  error
	linkedAt time.Time
}

func (i *identities) LinkIdentityToken(_ context.Context, userID, _ string, sealed []byte, linkedAt time.Time) error {
	if i.failure != nil {
		return i.failure
	}
	if i.linked == nil {
		i.linked = map[string][]byte{}
	}
	i.linked[userID] = sealed
	i.linkedAt = linkedAt
	return nil
}

func (i *identities) RevocableIdentities(context.Context, int) ([]identity.Sealed, error) {
	return i.held, i.failure
}

func (i *identities) IdentitiesForUser(context.Context, string) ([]identity.Sealed, error) {
	return i.held, i.failure
}

func (i *identities) MarkIdentityRevoked(_ context.Context, identityID string, _ time.Time) error {
	i.revoked = append(i.revoked, identityID)
	return nil
}

func sealing(t *testing.T) *secret.Sealer {
	t.Helper()
	value, err := secret.NewSealer(bytes.Repeat([]byte{'k'}, secret.KeyLength))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func linker(t *testing.T, source *provider, store *identities) *account.Linker {
	t.Helper()
	return account.NewLinker(source, store, sealing(t), appleProvider, clock.NewFixed(now()))
}

func TestLinkingStoresTheRefreshTokenSealed(t *testing.T) {
	source := &provider{refreshToken: "apple-refresh-token"}
	store := &identities{}
	if err := linker(t, source, store).Link(t.Context(), user.Principal{UserID: traveller}, "code-1"); err != nil {
		t.Fatal(err)
	}
	sealed, found := store.linked[traveller]
	if !found {
		t.Fatal("no identity was linked")
	}
	if bytes.Contains(sealed, []byte("apple-refresh-token")) {
		t.Fatal("the refresh token was stored in the clear")
	}
	opened, err := sealing(t).Open(sealed, []byte(traveller))
	if err != nil || opened != "apple-refresh-token" {
		t.Fatalf("the stored token does not open back: %q %v", opened, err)
	}
}

func TestLinkingRefusesAnUnusableAuthorizationCode(t *testing.T) {
	source := &provider{refreshToken: "token"}
	store := &identities{}
	for name, code := range map[string]string{
		"empty": "",
		"long":  string(make([]byte, 513)),
	} {
		err := linker(t, source, store).Link(t.Context(), user.Principal{UserID: traveller}, code)
		if apperror.CodeOf(err) != apperror.CodeInvalidInput {
			t.Fatalf("%s: err = %v, want invalid input", name, err)
		}
	}
	if len(source.exchanged) != 0 {
		t.Fatal("an unusable code was sent to the provider")
	}
}

func TestLinkingRefusesACodeTheProviderRejects(t *testing.T) {
	store := &identities{}
	err := linker(t, &provider{exchangeFailed: true}, store).
		Link(t.Context(), user.Principal{UserID: traveller}, "code-1")
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("err = %v, want invalid input", err)
	}
	if len(store.linked) != 0 {
		t.Fatal("a rejected code produced a stored identity")
	}
}

func TestLinkingRequiresAuthentication(t *testing.T) {
	source := &provider{refreshToken: "token"}
	err := linker(t, source, &identities{}).Link(t.Context(), user.Principal{}, "code-1")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("err = %v, want unauthenticated", err)
	}
}

func sealedFor(t *testing.T, token string) identity.Sealed {
	t.Helper()
	value, err := sealing(t).Seal(token, []byte(traveller))
	if err != nil {
		t.Fatal(err)
	}
	return identity.Sealed{ID: "identity-" + traveller, UserID: traveller, Provider: appleProvider, Token: value}
}

func TestRevokingSendsTheStoredTokenToTheProvider(t *testing.T) {
	source := &provider{}
	store := &identities{held: []identity.Sealed{sealedFor(t, "apple-refresh-token")}}
	revoked, err := linker(t, source, store).RevokeFor(t.Context(), traveller)
	if err != nil {
		t.Fatal(err)
	}
	if revoked != 1 || len(source.revoked) != 1 || source.revoked[0] != "apple-refresh-token" {
		t.Fatalf("revoked %d with provider calls %v", revoked, source.revoked)
	}
	if len(store.revoked) != 1 {
		t.Fatal("the identity was not marked revoked, so it would be retried forever")
	}
}

func TestAProviderOutageLeavesTheIdentityForTheNextSweep(t *testing.T) {
	source := &provider{revokeFailed: true}
	store := &identities{held: []identity.Sealed{sealedFor(t, "apple-refresh-token")}}
	revoked, err := linker(t, source, store).RevokePending(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if revoked != 0 {
		t.Fatalf("revoked = %d, want none while the provider is down", revoked)
	}
	if len(store.revoked) != 0 {
		t.Fatal("an identity was marked revoked without the provider confirming it")
	}
}

func TestATokenThatCannotBeOpenedIsRetiredRatherThanRetriedForever(t *testing.T) {
	source := &provider{}
	store := &identities{held: []identity.Sealed{{
		ID: "identity-1", UserID: traveller, Provider: appleProvider, Token: []byte("not-a-sealed-value"),
	}}}
	revoked, err := linker(t, source, store).RevokePending(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if revoked != 1 || len(store.revoked) != 1 {
		t.Fatalf("revoked = %d marked = %v, want the unusable token retired", revoked, store.revoked)
	}
	if len(source.revoked) != 0 {
		t.Fatal("an unopenable token was sent to the provider")
	}
}

func TestIdentitiesFromOtherProvidersAreLeftAlone(t *testing.T) {
	source := &provider{}
	other := sealedFor(t, "google-refresh-token")
	other.Provider = "google"
	store := &identities{held: []identity.Sealed{other}}
	revoked, err := linker(t, source, store).RevokeFor(t.Context(), traveller)
	if err != nil || revoked != 0 {
		t.Fatalf("revoked = %d %v, want another provider's identity untouched", revoked, err)
	}
}

func TestAnUnconfiguredLinkerRevokesNothingAndSaysSo(t *testing.T) {
	var absent *account.Linker
	if revoked, err := absent.RevokePending(t.Context(), 10); revoked != 0 || err != nil {
		t.Fatalf("revoked = %d %v", revoked, err)
	}
	err := absent.Link(t.Context(), user.Principal{UserID: traveller}, "code-1")
	if apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatalf("err = %v, want unavailable", err)
	}
}

func TestDeletionRevokesSignInTokens(t *testing.T) {
	source := &provider{}
	store := &identities{held: []identity.Sealed{sealedFor(t, "apple-refresh-token")}}
	eraser := account.NewEraser(&deletions{}, generator(), clock.NewFixed(now())).
		WithLinker(linker(t, source, store))
	if _, err := eraser.Request(t.Context(), user.Principal{UserID: traveller}, ""); err != nil {
		t.Fatal(err)
	}
	if len(source.revoked) != 1 {
		t.Fatalf("provider calls = %v, want the token revoked when the account closed", source.revoked)
	}
}

func TestDeletionSucceedsEvenWhenRevocationFails(t *testing.T) {
	source := &provider{revokeFailed: true}
	store := &identities{held: []identity.Sealed{sealedFor(t, "apple-refresh-token")}}
	eraser := account.NewEraser(&deletions{}, generator(), clock.NewFixed(now())).
		WithLinker(linker(t, source, store))
	accepted, err := eraser.Request(t.Context(), user.Principal{UserID: traveller}, "")
	if err != nil {
		t.Fatalf("a provider outage blocked the traveller from closing their account: %v", err)
	}
	if accepted.ID == "" {
		t.Fatal("no deletion was recorded")
	}
}
