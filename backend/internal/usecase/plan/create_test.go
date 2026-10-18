package plan_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

type txKey struct{}
type writer struct {
	inserted trip.Trip
	owner    string
	calls    int
	fail     bool
}

func (w *writer) Insert(ctx context.Context, value trip.Trip) error {
	if ctx.Value(txKey{}) != true {
		return errors.New("outside transaction")
	}
	w.inserted = value
	w.calls++
	return nil
}
func (w *writer) AddOwner(ctx context.Context, _, owner string) error {
	if ctx.Value(txKey{}) != true {
		return errors.New("outside transaction")
	}
	if w.fail {
		return errors.New("write failed")
	}
	w.owner = owner
	return nil
}

type transaction struct{}

func (transaction) Within(ctx context.Context, operation func(context.Context) error) error {
	return operation(context.WithValue(ctx, txKey{}, true))
}

func TestCreateUsesPrincipalAndAtomicBoundary(t *testing.T) {
	store := &writer{}
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	creator := plan.NewCreator(store, transaction{}, id.NewRandom(), clock.NewFixed(now))
	value, err := creator.Create(t.Context(), user.Principal{UserID: "actual"}, trip.Trip{OwnerID: "spoofed", Status: trip.Active, Adults: 1, Currency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if value.OwnerID != "actual" || store.owner != "actual" || value.Status != trip.Draft || value.ID == "" || value.CreatedAt != now {
		t.Fatalf("%+v", value)
	}
	store.fail = true
	if _, err := creator.Create(t.Context(), user.Principal{UserID: "actual"}, trip.Trip{Adults: 1, Currency: "USD"}); err == nil {
		t.Fatal("lost membership failure")
	}
}

func TestCreateRejectsMissingPrincipalAndInvalidTripBeforeWrites(t *testing.T) {
	store := &writer{}
	creator := plan.NewCreator(store, transaction{}, id.NewRandom(), clock.NewSystem())
	for _, principal := range []user.Principal{{}, {UserID: "user"}} {
		if _, err := creator.Create(t.Context(), principal, trip.Trip{}); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
	if store.calls != 0 {
		t.Fatal("invalid request reached repository")
	}
}
