package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	identifiers "github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

type deletions struct {
	stored    map[string]retention.Deletion
	requested int
}

func (d *deletions) RequestDeletion(_ context.Context, deletion retention.Deletion) (retention.Deletion, error) {
	d.requested++
	if d.stored == nil {
		d.stored = map[string]retention.Deletion{}
	}
	if existing, found := d.stored[deletion.UserID]; found {
		return existing, nil
	}
	d.stored[deletion.UserID] = deletion
	return deletion, nil
}

func (d *deletions) Deletion(_ context.Context, userID string) (retention.Deletion, error) {
	found, ok := d.stored[userID]
	if !ok {
		return retention.Deletion{}, apperror.New(apperror.CodeNotFound, "No deletion is pending.")
	}
	return found, nil
}

func now() time.Time { return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC) }

func generator() account.Identifiers { return identifiers.NewUUID() }

func newEraser(store *deletions) *account.Eraser {
	return account.NewEraser(store, generator(), clock.NewFixed(now()))
}

func TestRequestingDeletionSchedulesErasure(t *testing.T) {
	store := &deletions{}
	accepted, err := newEraser(store).Request(t.Context(), user.Principal{UserID: traveller}, "no_longer_travelling")
	if err != nil {
		t.Fatal(err)
	}
	if accepted.Status != retention.Pending {
		t.Fatalf("status = %q, want pending", accepted.Status)
	}
	if !accepted.PurgeAfter.Equal(now().Add(retention.GracePeriod)) {
		t.Fatalf("purge after = %s, want the grace period from now", accepted.PurgeAfter)
	}
}

func TestAskingTwiceDoesNotRestartTheClock(t *testing.T) {
	store := &deletions{}
	eraser := newEraser(store)
	first, err := eraser.Request(t.Context(), user.Principal{UserID: traveller}, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := eraser.Request(t.Context(), user.Principal{UserID: traveller}, "changed_my_mind")
	if err != nil {
		t.Fatal(err)
	}
	if !second.PurgeAfter.Equal(first.PurgeAfter) || second.ID != first.ID {
		t.Fatalf("a second request moved erasure from %s to %s", first.PurgeAfter, second.PurgeAfter)
	}
}

func TestDeletionRequiresAnAuthenticatedTraveller(t *testing.T) {
	store := &deletions{}
	_, err := newEraser(store).Request(t.Context(), user.Principal{}, "")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("err = %v, want unauthenticated", err)
	}
	if store.requested != 0 {
		t.Fatal("an unauthenticated caller reached the store")
	}
}

func TestAReasonLongerThanTheContractIsRefused(t *testing.T) {
	store := &deletions{}
	_, err := newEraser(store).Request(t.Context(), user.Principal{UserID: traveller}, string(make([]byte, 65)))
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("err = %v, want invalid input", err)
	}
	if store.requested != 0 {
		t.Fatal("an over-long reason reached the store")
	}
}

func TestStatusReportsAPendingDeletion(t *testing.T) {
	store := &deletions{}
	eraser := newEraser(store)
	if _, err := eraser.Request(t.Context(), user.Principal{UserID: traveller}, ""); err != nil {
		t.Fatal(err)
	}
	found, err := eraser.Status(t.Context(), user.Principal{UserID: traveller})
	if err != nil {
		t.Fatal(err)
	}
	if found.Status != retention.Pending {
		t.Fatalf("status = %q, want pending", found.Status)
	}
	if _, err := eraser.Status(t.Context(), user.Principal{}); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("err = %v, want unauthenticated", err)
	}
}
