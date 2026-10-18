package retention_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

func requestedAt() time.Time { return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC) }

func TestARequestSchedulesErasureAfterTheGracePeriod(t *testing.T) {
	deletion, err := retention.Request("deletion-1", traveller, "no_longer_travelling", requestedAt())
	if err != nil {
		t.Fatal(err)
	}
	if deletion.Status != retention.Pending {
		t.Fatalf("status = %q, want pending", deletion.Status)
	}
	if !deletion.PurgeAfter.Equal(requestedAt().Add(retention.GracePeriod)) {
		t.Fatalf("purge after = %s, want the grace period from the request", deletion.PurgeAfter)
	}
	if deletion.PurgedAt != nil {
		t.Fatal("a fresh request claims to have been purged")
	}
}

func TestARequestNeedsAnAccountAndAShortReason(t *testing.T) {
	for name, test := range map[string]struct{ id, user, reason string }{
		"no identifier": {"", traveller, ""},
		"no account":    {"deletion-1", "", ""},
		"long reason":   {"deletion-1", traveller, string(make([]byte, 65))},
	} {
		_, err := retention.Request(test.id, test.user, test.reason, requestedAt())
		if apperror.CodeOf(err) != apperror.CodeInvalidInput {
			t.Fatalf("%s: err = %v, want invalid input", name, err)
		}
	}
}

func TestNothingIsDueBeforeTheGracePeriodEnds(t *testing.T) {
	deletion, err := retention.Request("deletion-1", traveller, "", requestedAt())
	if err != nil {
		t.Fatal(err)
	}
	if deletion.Due(requestedAt().Add(retention.GracePeriod - time.Second)) {
		t.Fatal("erasure came due a second early")
	}
	if !deletion.Due(deletion.PurgeAfter) {
		t.Fatal("erasure never came due")
	}
}

func TestAPurgedDeletionIsNeverDueAgain(t *testing.T) {
	deletion, err := retention.Request("deletion-1", traveller, "", requestedAt())
	if err != nil {
		t.Fatal(err)
	}
	completed := deletion.Complete(deletion.PurgeAfter, 2, 3)
	if completed.Status != retention.Purged || completed.PurgedAt == nil {
		t.Fatalf("completed = %+v, want a purged record with a timestamp", completed)
	}
	if completed.TripsTransferred != 2 || completed.TripsDeleted != 3 {
		t.Fatalf("trips = %d transferred and %d deleted, want 2 and 3",
			completed.TripsTransferred, completed.TripsDeleted)
	}
	if completed.Due(completed.PurgeAfter.Add(time.Hour)) {
		t.Fatal("a purged account was queued for purging again")
	}
}
