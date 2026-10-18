//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

func startSession(t *testing.T, store *postgres.Store, userID string) (session.Session, []byte) {
	t.Helper()
	sessionID, err := id.NewUUID().Generate()
	if err != nil {
		t.Fatal(err)
	}
	secret, err := id.NewRandom().Generate()
	if err != nil {
		t.Fatal(err)
	}
	started := session.Start(sessionID, sessionID, userID, time.Now().UTC())
	hash := session.HashRefreshToken(secret)
	if err := store.InsertSession(t.Context(), started, hash); err != nil {
		t.Fatal(err)
	}
	return started, hash
}

func TestASessionRoundTripsThroughItsRefreshHash(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	started, hash := startSession(t, store, principal.UserID)

	found, err := store.SessionByRefreshHash(t.Context(), hash)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != started.ID || found.UserID != principal.UserID || found.RotatedAt != nil || found.RevokedAt != nil {
		t.Fatalf("found = %+v", found)
	}
	usable, err := store.SessionAccess(t.Context(), started.ID)
	if err != nil || !usable {
		t.Fatalf("usable = %v %v, want a fresh session to grant access", usable, err)
	}
	if _, err := store.SessionByRefreshHash(t.Context(), session.HashRefreshToken("unknown")); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("an unknown hash returned %v", err)
	}
}

func TestRotationAndRevocationArePersisted(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	started, hash := startSession(t, store, principal.UserID)
	now := time.Now().UTC()

	if err := store.MarkSessionRotated(t.Context(), started.ID, now); err != nil {
		t.Fatal(err)
	}
	rotated, err := store.SessionByRefreshHash(t.Context(), hash)
	if err != nil || rotated.RotatedAt == nil || !rotated.Reused() {
		t.Fatalf("rotated = %+v %v, want the rotation recorded", rotated, err)
	}
	if err := store.RevokeSessionFamily(t.Context(), started.FamilyID, now); err != nil {
		t.Fatal(err)
	}
	usable, err := store.SessionAccess(t.Context(), started.ID)
	if err != nil || usable {
		t.Fatalf("usable = %v %v, want a revoked family to lose access", usable, err)
	}
}

func TestClosingAnAccountEndsSessionAccess(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	started, _ := startSession(t, store, principal.UserID)

	requestDeletion(t, store, principal, time.Now().UTC())
	usable, err := store.SessionAccess(t.Context(), started.ID)
	if err != nil || usable {
		t.Fatalf("usable = %v %v, want a closed account to lose session access", usable, err)
	}
}

func TestPurgingAnAccountRemovesItsSessions(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	started, _ := startSession(t, store, principal.UserID)
	deletion := requestDeletion(t, store, principal, time.Now().Add(-2*retention.GracePeriod).UTC())
	if _, err := store.Purge(t.Context(), deletion, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := database.QueryRowContext(t.Context(), "SELECT count(*) FROM sessions WHERE id = $1", started.ID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatal("a session outlived its account")
	}
}
