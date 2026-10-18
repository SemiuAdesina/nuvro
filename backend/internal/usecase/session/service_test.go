package session_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/sessiontoken"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/session"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

type identity struct{ accepted string }

func (i identity) Verify(_ context.Context, credential string) (user.Identity, error) {
	if credential != i.accepted {
		return user.Identity{}, errors.New("rejected")
	}
	return user.Identity{Provider: "apple", Subject: "001234.abc"}, nil
}

type accounts struct{ closed bool }

func (a accounts) Resolve(_ context.Context, value user.Identity) (user.Principal, error) {
	if a.closed {
		return user.Principal{}, apperror.New(apperror.CodeForbidden, "This account cannot be used.")
	}
	return user.Principal{UserID: traveller, Provider: value.Provider, Subject: value.Subject}, nil
}

type linker struct{ codes []string }

func (l *linker) Link(_ context.Context, _ user.Principal, code string) error {
	l.codes = append(l.codes, code)
	return nil
}

type memory struct {
	byHash map[string]domain.Session
	byID   map[string]domain.Session
	active bool
}

func newMemory() *memory {
	return &memory{byHash: map[string]domain.Session{}, byID: map[string]domain.Session{}, active: true}
}

func (m *memory) InsertSession(_ context.Context, value domain.Session, hash []byte) error {
	m.byHash[string(hash)] = value
	m.byID[value.ID] = value
	return nil
}

func (m *memory) SessionByRefreshHash(_ context.Context, hash []byte) (domain.Session, error) {
	found, ok := m.byHash[string(hash)]
	if !ok {
		return domain.Session{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return m.byID[found.ID], nil
}

func (m *memory) SessionAccess(_ context.Context, sessionID string) (bool, error) {
	found, ok := m.byID[sessionID]
	return ok && found.RevokedAt == nil && m.active, nil
}

func (m *memory) MarkSessionRotated(_ context.Context, sessionID string, at time.Time) error {
	found := m.byID[sessionID]
	found.RotatedAt = &at
	m.byID[sessionID] = found
	return nil
}

func (m *memory) RevokeSession(_ context.Context, sessionID string, at time.Time) error {
	found := m.byID[sessionID]
	found.RevokedAt = &at
	m.byID[sessionID] = found
	return nil
}

func (m *memory) RevokeSessionFamily(_ context.Context, familyID string, at time.Time) error {
	for key, found := range m.byID {
		if found.FamilyID == familyID {
			found.RevokedAt = &at
			m.byID[key] = found
		}
	}
	return nil
}

func (m *memory) RevokeUserSessions(_ context.Context, userID string, at time.Time) error {
	for key, found := range m.byID {
		if found.UserID == userID {
			found.RevokedAt = &at
			m.byID[key] = found
		}
	}
	return nil
}

func now() time.Time { return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC) }

func service(t *testing.T, store *memory, link *linker, timer clock.Clock) *session.Service {
	t.Helper()
	signer, err := sessiontoken.NewSigner(bytes.Repeat([]byte{'k'}, sessiontoken.MinimumKeyLength))
	if err != nil {
		t.Fatal(err)
	}
	dependencies := session.Dependencies{
		Identity: identity{accepted: "apple-identity-token"}, Accounts: accounts{}, Store: store,
		Signer: signer, Identifiers: id.NewUUID(), Secrets: id.NewRandom(),
	}
	if link != nil {
		dependencies.Linker = link
	}
	return session.NewService(dependencies, timer)
}

func TestAVerifiedIdentityIsExchangedForASession(t *testing.T) {
	store := newMemory()
	granted, err := service(t, store, nil, clock.NewFixed(now())).Issue(t.Context(), "apple-identity-token", "")
	if err != nil {
		t.Fatal(err)
	}
	if granted.AccessToken == "" || granted.RefreshToken == "" || granted.UserID != traveller {
		t.Fatalf("grant = %+v", granted)
	}
	if !granted.ExpiresAt.Equal(now().Add(domain.AccessLifetime)) {
		t.Fatalf("expires = %s, want the access lifetime from now", granted.ExpiresAt)
	}
	if _, stored := store.byHash[string(domain.HashRefreshToken(granted.RefreshToken))]; !stored {
		t.Fatal("the refresh token hash was not stored")
	}
	for hash := range store.byHash {
		if hash == granted.RefreshToken {
			t.Fatal("the refresh token was stored in the clear")
		}
	}
}

func TestTheAuthorizationCodeIsLinkedWhenPresent(t *testing.T) {
	link := &linker{}
	if _, err := service(t, newMemory(), link, clock.NewFixed(now())).Issue(t.Context(), "apple-identity-token", "code-1"); err != nil {
		t.Fatal(err)
	}
	if len(link.codes) != 1 || link.codes[0] != "code-1" {
		t.Fatalf("linked = %v, want the authorization code handed over once", link.codes)
	}
}

func TestARejectedCredentialNeverBecomesASession(t *testing.T) {
	store := newMemory()
	for name, credential := range map[string]string{"wrong": "not-the-token", "empty": ""} {
		_, err := service(t, store, nil, clock.NewFixed(now())).Issue(t.Context(), credential, "")
		if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
			t.Fatalf("%s: err = %v, want unauthenticated", name, err)
		}
	}
	if len(store.byID) != 0 {
		t.Fatal("a session was created for a rejected credential")
	}
}

func TestTheAccessTokenAuthenticatesUntilTheSessionIsRevoked(t *testing.T) {
	store := newMemory()
	svc := service(t, store, nil, clock.NewFixed(now()))
	granted, err := svc.Issue(t.Context(), "apple-identity-token", "")
	if err != nil {
		t.Fatal(err)
	}
	principal, handled, err := svc.Authenticate(t.Context(), granted.AccessToken)
	if err != nil || !handled || principal.UserID != traveller || principal.Provider != domain.Provider {
		t.Fatalf("principal = %+v handled = %v err = %v", principal, handled, err)
	}
	if err := svc.Revoke(t.Context(), principal); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(t.Context(), granted.AccessToken); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("a revoked session still authenticated: %v", err)
	}
}

func TestATokenThatIsNotASessionTokenIsLeftToTheIdentityVerifier(t *testing.T) {
	svc := service(t, newMemory(), nil, clock.NewFixed(now()))
	_, handled, err := svc.Authenticate(t.Context(), "eyJhbGciOiJSUzI1NiJ9.apple.token")
	if handled || err != nil {
		t.Fatalf("handled = %v err = %v, want an Apple token passed through", handled, err)
	}
	var absent *session.Service
	if _, handled, _ := absent.Authenticate(t.Context(), "anything"); handled {
		t.Fatal("an absent service claimed a token")
	}
}

func TestRefreshRotatesAndTheOldTokenDies(t *testing.T) {
	store := newMemory()
	svc := service(t, store, nil, clock.NewFixed(now()))
	first, err := svc.Issue(t.Context(), "apple-identity-token", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Refresh(t.Context(), first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if second.RefreshToken == first.RefreshToken || second.SessionID == first.SessionID {
		t.Fatal("refresh did not rotate the session")
	}
	if _, _, err := svc.Authenticate(t.Context(), first.AccessToken); err != nil {
		t.Fatalf("the previous access token died before its expiry: %v", err)
	}
	if _, err := svc.Refresh(t.Context(), first.RefreshToken); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("a used refresh token was accepted again: %v", err)
	}
	if _, err := svc.Refresh(t.Context(), second.RefreshToken); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("reuse of an old token did not revoke the whole family: %v", err)
	}
	if _, _, err := svc.Authenticate(t.Context(), second.AccessToken); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("a family revoked for reuse still authenticated: %v", err)
	}
}

func TestAnExpiredRefreshTokenIsRefused(t *testing.T) {
	store := newMemory()
	granted, err := service(t, store, nil, clock.NewFixed(now())).Issue(t.Context(), "apple-identity-token", "")
	if err != nil {
		t.Fatal(err)
	}
	later := service(t, store, nil, clock.NewFixed(now().Add(domain.RefreshLifetime)))
	if _, err := later.Refresh(t.Context(), granted.RefreshToken); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("an expired refresh token was accepted: %v", err)
	}
	if _, err := later.Refresh(t.Context(), "unknown"); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("an unknown refresh token was accepted: %v", err)
	}
}

func TestClosingTheAccountKillsEverySession(t *testing.T) {
	store := newMemory()
	svc := service(t, store, nil, clock.NewFixed(now()))
	first, _ := svc.Issue(t.Context(), "apple-identity-token", "")
	second, _ := svc.Issue(t.Context(), "apple-identity-token", "")
	if err := svc.RevokeAll(t.Context(), traveller); err != nil {
		t.Fatal(err)
	}
	for name, token := range map[string]string{"first": first.AccessToken, "second": second.AccessToken} {
		if _, _, err := svc.Authenticate(t.Context(), token); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
			t.Fatalf("%s session survived account closure: %v", name, err)
		}
	}
}

func TestRevokeIgnoresPrincipalsThatAreNotSessions(t *testing.T) {
	store := newMemory()
	svc := service(t, store, nil, clock.NewFixed(now()))
	granted, _ := svc.Issue(t.Context(), "apple-identity-token", "")
	if err := svc.Revoke(t.Context(), user.Principal{UserID: traveller, Provider: "apple", Subject: "001234.abc"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(t.Context(), granted.AccessToken); err != nil {
		t.Fatalf("an Apple-authenticated sign-out revoked a session it did not name: %v", err)
	}
}
