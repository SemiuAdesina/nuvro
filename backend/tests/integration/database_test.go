//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

func openDatabase(t *testing.T) (*postgres.Pool, *sql.DB) {
	t.Helper()
	url := os.Getenv("NUVRO_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("NUVRO_TEST_DATABASE_URL must name an isolated disposable database")
	}
	pool, err := postgres.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpContext(t.Context(), database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	return pool, database
}

func TestMigrationsRepositoriesAndTransactions(t *testing.T) {
	pool, database := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	owner := "00112233-4455-6677-8899-aabbccddeeff"
	if _, err := pool.Exec(t.Context(), "INSERT INTO users(id) VALUES($1)", owner); err != nil {
		t.Fatal(err)
	}
	creator := plan.NewCreator(store, store, id.NewUUID(), clock.NewSystem())
	created, err := creator.Create(t.Context(), user.Principal{UserID: owner}, trip.Trip{Adults: 2, Children: 1, Currency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := plan.NewReader(store).Read(t.Context(), user.Principal{UserID: owner}, created.ID)
	if err != nil || loaded.ID != created.ID || loaded.OwnerID != owner || loaded.DestinationID != nil || loaded.StartDate != nil || loaded.Adults != 2 {
		t.Fatalf("%+v %v", loaded, err)
	}
	assertRollback(t, store, owner)
	assertConstraints(t, pool, owner, created.ID)
	assertListingAndUpdates(t, store, owner, created)
	if err := goose.DownToContext(t.Context(), database, "../../db/migrations", 0); err != nil {
		t.Fatal(err)
	}
	var name *string
	if err := pool.QueryRow(t.Context(), "SELECT to_regclass('public.users')::text").Scan(&name); err != nil || name != nil {
		t.Fatalf("down did not remove tables: %v %v", name, err)
	}
	if err := goose.UpContext(t.Context(), database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
}

func assertListingAndUpdates(t *testing.T, store *postgres.Store, owner string, first trip.Trip) {
	t.Helper()
	principal := user.Principal{UserID: owner}
	creator := plan.NewCreator(store, store, id.NewUUID(), clock.NewSystem())
	second, err := creator.Create(t.Context(), principal, trip.Trip{Adults: 1, Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.New([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	lister := plan.NewLister(store, codec)
	page, err := lister.List(t.Context(), principal, "", 1)
	if err != nil || len(page.Trips) != 1 || !page.HasMore || page.NextCursor == "" {
		t.Fatalf("first page: %+v %v", page, err)
	}
	next, err := lister.List(t.Context(), principal, page.NextCursor, 1)
	if err != nil || len(next.Trips) != 1 || next.Trips[0].ID == page.Trips[0].ID {
		t.Fatalf("second page repeated a trip: %+v %v", next, err)
	}
	seen := map[string]bool{page.Trips[0].ID: true, next.Trips[0].ID: true}
	if !seen[first.ID] || !seen[second.ID] {
		t.Fatalf("pagination lost a trip: %v", seen)
	}
	if _, err := lister.List(t.Context(), user.Principal{UserID: "00000000-0000-4000-8000-000000000000"}, page.NextCursor, 1); err == nil {
		t.Fatal("another traveller replayed the cursor against the database")
	}
	updated, err := plan.NewUpdater(store, store, clock.NewSystem()).
		Update(t.Context(), principal, second.ID, trip.Patch{Currency: change.Set("GBP"), Adults: change.Set(int32(4))})
	if err != nil || updated.Currency != "GBP" || updated.Adults != 4 {
		t.Fatalf("%+v %v", updated, err)
	}
	reread, err := plan.NewReader(store).Read(t.Context(), principal, second.ID)
	if err != nil || reread.Currency != "GBP" || reread.Adults != 4 || !reread.UpdatedAt.After(second.UpdatedAt) {
		t.Fatalf("the update did not persist: %+v %v", reread, err)
	}
}

func assertRollback(t *testing.T, store *postgres.Store, owner string) {
	t.Helper()
	identifier := "11112233-4455-6677-8899-aabbccddeeff"
	failure := errors.New("second write failed")
	err := store.Within(t.Context(), func(ctx context.Context) error {
		err := store.Insert(ctx, trip.Trip{ID: identifier, OwnerID: owner, Adults: 1, Currency: "USD", Status: trip.Draft, CreatedAt: time.Now(), UpdatedAt: time.Now()})
		if err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("rollback: %v", err)
	}
	if _, err := store.Get(t.Context(), identifier); err == nil {
		t.Fatal("transaction committed after failure")
	}
	err = store.Within(t.Context(), func(ctx context.Context) error { return store.Within(ctx, func(context.Context) error { return nil }) })
	if err == nil {
		t.Fatal("nested transaction accepted")
	}
}

func assertConstraints(t *testing.T, pool *postgres.Pool, owner, tripID string) {
	t.Helper()
	for _, statement := range []string{
		"UPDATE trips SET adults=0 WHERE id=$1",
		"UPDATE trips SET preferred_currency='usd' WHERE id=$1",
		"UPDATE trips SET status='imaginary' WHERE id=$1",
		"UPDATE trips SET start_date='2026-09-10',end_date='2026-09-01' WHERE id=$1",
		"UPDATE trips SET start_date='2026-09-10',end_date=NULL WHERE id=$1",
	} {
		if _, err := pool.Exec(t.Context(), statement, tripID); err == nil {
			t.Fatal("database accepted invalid trip")
		}
	}
	query := "INSERT INTO user_identities(id,user_id,provider,provider_subject) VALUES($1,$2,'test','synthetic')"
	if _, err := pool.Exec(t.Context(), query, "22112233-4455-6677-8899-aabbccddeeff", owner); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), query, "33112233-4455-6677-8899-aabbccddeeff", owner); err == nil {
		t.Fatal("duplicate identity accepted")
	}
}
