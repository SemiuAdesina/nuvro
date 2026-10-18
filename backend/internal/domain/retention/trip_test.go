package retention_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
)

func joined(userID string, days int) retention.Membership {
	return retention.Membership{UserID: userID, JoinedAt: requestedAt().Add(time.Duration(days) * 24 * time.Hour)}
}

func TestTheLongestStandingMemberInheritsASharedTrip(t *testing.T) {
	succession, found := retention.Succeed("trip-1", []retention.Membership{
		joined("late", 5), joined("early", 1), joined("middle", 3),
	})
	if !found {
		t.Fatal("a shared trip found no successor")
	}
	if succession.Successor != "early" {
		t.Fatalf("successor = %q, want the longest-standing member", succession.Successor)
	}
	if succession.TripID != "trip-1" {
		t.Fatalf("trip = %q, want the trip being transferred", succession.TripID)
	}
}

func TestSuccessionIsSettledDeterministicallyWhenMembersJoinedTogether(t *testing.T) {
	first, _ := retention.Succeed("trip-1", []retention.Membership{joined("bravo", 1), joined("alpha", 1)})
	second, _ := retention.Succeed("trip-1", []retention.Membership{joined("alpha", 1), joined("bravo", 1)})
	if first.Successor != second.Successor {
		t.Fatalf("succession depended on ordering: %q then %q", first.Successor, second.Successor)
	}
	if first.Successor != "alpha" {
		t.Fatalf("successor = %q, want the lower identifier as the tie-break", first.Successor)
	}
}

func TestATripWithNoOneLeftHasNoSuccessor(t *testing.T) {
	if _, found := retention.Succeed("trip-1", nil); found {
		t.Fatal("an empty trip found a successor")
	}
	if _, found := retention.Succeed("trip-1", []retention.Membership{{UserID: ""}}); found {
		t.Fatal("a blank membership was treated as a successor")
	}
}
