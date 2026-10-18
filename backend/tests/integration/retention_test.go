//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

func enrol(t *testing.T, store *postgres.Store) user.Principal {
	t.Helper()
	identity := user.Identity{Provider: "apple", Subject: "001234." + randomSuffix(t)}
	principal, err := account.NewResolver(store, store, id.NewUUID()).Resolve(t.Context(), identity)
	if err != nil {
		t.Fatal(err)
	}
	return principal
}

func ownTrip(t *testing.T, store *postgres.Store, owner user.Principal) string {
	t.Helper()
	identifiers := id.NewUUID()
	tripID, err := identifiers.Generate()
	if err != nil {
		t.Fatal(err)
	}
	planned := trip.Trip{
		ID: tripID, OwnerID: owner.UserID, Adults: 1, Currency: "USD",
		Status: trip.Draft, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	err = store.Within(t.Context(), func(ctx context.Context) error {
		if err := store.Insert(ctx, planned); err != nil {
			return err
		}
		return store.AddOwner(ctx, tripID, owner.UserID)
	})
	if err != nil {
		t.Fatal(err)
	}
	return tripID
}

func requestDeletion(t *testing.T, store *postgres.Store, principal user.Principal, now time.Time) retention.Deletion {
	t.Helper()
	eraser := account.NewEraser(store, id.NewUUID(), clock.NewFixed(now))
	accepted, err := eraser.Request(t.Context(), principal, "no_longer_travelling")
	if err != nil {
		t.Fatal(err)
	}
	return accepted
}

func TestRequestingDeletionClosesTheAccountAndSchedulesErasure(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	subject := "001234." + randomSuffix(t)
	principal, err := account.NewResolver(store, store, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: subject})
	if err != nil {
		t.Fatal(err)
	}

	accepted := requestDeletion(t, store, principal, time.Now().UTC())
	if accepted.Status != retention.Pending {
		t.Fatalf("status = %q, want pending", accepted.Status)
	}
	_, status, err := store.FindUserByIdentity(t.Context(), "apple", subject)
	if err != nil {
		t.Fatal(err)
	}
	if user.IsActive(status) {
		t.Fatalf("status = %q, want an account that is no longer active", status)
	}
	stored, err := store.Deletion(t.Context(), principal.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.PurgeAfter.Equal(accepted.PurgeAfter) {
		t.Fatalf("purge after = %s, want %s", stored.PurgeAfter, accepted.PurgeAfter)
	}
}

func TestASharedTripPassesToTheLongestStandingMember(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	owner := enrol(t, store)
	companion := enrol(t, store)
	tripID := ownTrip(t, store, owner)
	addMember(t, database, tripID, companion.UserID)

	deletion := requestDeletion(t, store, owner, time.Now().Add(-2*retention.GracePeriod).UTC())
	completed, err := store.Purge(t.Context(), deletion, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if completed.TripsTransferred != 1 || completed.TripsDeleted != 0 {
		t.Fatalf("completed = %+v, want one trip transferred and none deleted", completed)
	}

	var survivor string
	if err := database.QueryRowContext(t.Context(),
		"SELECT owner_id FROM trips WHERE id = $1", tripID).Scan(&survivor); err != nil {
		t.Fatal(err)
	}
	if survivor != companion.UserID {
		t.Fatalf("owner = %q, want the remaining member %q", survivor, companion.UserID)
	}
	var role string
	if err := database.QueryRowContext(t.Context(),
		"SELECT role FROM trip_members WHERE trip_id = $1 AND user_id = $2",
		tripID, companion.UserID).Scan(&role); err != nil {
		t.Fatal(err)
	}
	if role != "owner" {
		t.Fatalf("role = %q, want the successor promoted to owner", role)
	}
}

func TestASweepRemovesOnlyRecordsPastTheirRetentionAge(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	rule, found := retention.RuleFor(retention.PersonalizationSignals)
	if !found {
		t.Fatal("personalization signals are not covered by the policy")
	}
	now := time.Now().UTC()
	stale := recordSignal(t, store, principal.UserID, "expired-city", now.Add(-rule.MaxAge-time.Hour))
	fresh := recordSignal(t, store, principal.UserID, "current-city", now.Add(-time.Hour))

	if _, err := store.Sweep(t.Context(), rule, now); err != nil {
		t.Fatal(err)
	}
	if signalCount(t, database, stale) != 0 {
		t.Fatal("a signal past its retention age survived the sweep")
	}
	if signalCount(t, database, fresh) != 1 {
		t.Fatal("the sweep removed a signal that is still within its retention age")
	}
}

func addMember(t *testing.T, database *sql.DB, tripID, userID string) {
	t.Helper()
	_, err := database.ExecContext(t.Context(),
		"INSERT INTO trip_members (trip_id, user_id, role) VALUES ($1, $2, 'editor')", tripID, userID)
	if err != nil {
		t.Fatal(err)
	}
}

func recordSignal(t *testing.T, store *postgres.Store, userID, subjectID string, occurredAt time.Time) string {
	t.Helper()
	signalID, err := id.NewUUID().Generate()
	if err != nil {
		t.Fatal(err)
	}
	err = store.RecordSignal(t.Context(), signalID, domain.Signal{
		UserID: userID, Type: domain.DestinationSaved, Subject: domain.Destination,
		SubjectID: subjectID, Strength: domain.Weak, OccurredAt: occurredAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	return signalID
}

func signalCount(t *testing.T, database *sql.DB, signalID string) int {
	t.Helper()
	var found int
	err := database.QueryRowContext(t.Context(),
		"SELECT count(*) FROM personalization_signals WHERE id = $1", signalID).Scan(&found)
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestASignedInAccountIsRefusedOnceDeletionIsRequested(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	subject := "001234." + randomSuffix(t)
	identity := user.Identity{Provider: "apple", Subject: subject}
	resolver := account.NewResolver(store, store, id.NewUUID())
	principal, err := resolver.Resolve(t.Context(), identity)
	if err != nil {
		t.Fatal(err)
	}
	requestDeletion(t, store, principal, time.Now().UTC())

	if _, err := resolver.Resolve(t.Context(), identity); err == nil {
		t.Fatal("a closed account signed in again")
	}
}

func TestPurgingErasesEverythingTheAccountOwned(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	tripID := ownTrip(t, store, principal)

	signalID, err := id.NewUUID().Generate()
	if err != nil {
		t.Fatal(err)
	}
	err = store.RecordSignal(t.Context(), signalID, domain.Signal{
		UserID: principal.UserID, Type: domain.DestinationSaved, Subject: domain.Destination,
		SubjectID: "lisbon", Strength: domain.Weak, OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	deletion := requestDeletion(t, store, principal, time.Now().Add(-2*retention.GracePeriod).UTC())
	due, err := store.DueDeletions(t.Context(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if !containsDeletion(due, deletion.ID) {
		t.Fatalf("the account never came due for erasure: %+v", due)
	}

	completed, err := store.Purge(t.Context(), deletion, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != retention.Purged || completed.TripsDeleted != 1 {
		t.Fatalf("completed = %+v, want a purged record naming the deleted trip", completed)
	}

	for name, query := range map[string]string{
		"user":            "SELECT count(*) FROM users WHERE id = $1",
		"identity":        "SELECT count(*) FROM user_identities WHERE user_id = $1",
		"profile":         "SELECT count(*) FROM user_profiles WHERE user_id = $1",
		"preferences":     "SELECT count(*) FROM traveler_preferences WHERE user_id = $1",
		"signals":         "SELECT count(*) FROM personalization_signals WHERE user_id = $1",
		"trip membership": "SELECT count(*) FROM trip_members WHERE user_id = $1",
		"owned trips":     "SELECT count(*) FROM trips WHERE owner_id = $1",
	} {
		var remaining int
		if err := database.QueryRowContext(t.Context(), query, principal.UserID).Scan(&remaining); err != nil {
			t.Fatal(err)
		}
		if remaining != 0 {
			t.Fatalf("%s survived erasure: %d rows", name, remaining)
		}
	}
	var trips int
	if err := database.QueryRowContext(t.Context(), "SELECT count(*) FROM trips WHERE id = $1", tripID).Scan(&trips); err != nil {
		t.Fatal(err)
	}
	if trips != 0 {
		t.Fatal("the trip outlived the only account that could see it")
	}
}

func TestPurgingKeepsTheDeletionReceiptAfterTheAccountIsGone(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	principal := enrol(t, store)
	deletion := requestDeletion(t, store, principal, time.Now().Add(-2*retention.GracePeriod).UTC())

	if _, err := store.Purge(t.Context(), deletion, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var status string
	var purgedAt *time.Time
	err := database.QueryRowContext(t.Context(),
		"SELECT status, purged_at FROM account_deletions WHERE id = $1", deletion.ID).Scan(&status, &purgedAt)
	if err != nil {
		t.Fatal(err)
	}
	if status != string(retention.Purged) || purgedAt == nil {
		t.Fatalf("receipt = %q at %v, want a purged receipt with a timestamp", status, purgedAt)
	}
}

func containsDeletion(due []retention.Deletion, deletionID string) bool {
	for _, deletion := range due {
		if deletion.ID == deletionID {
			return true
		}
	}
	return false
}
