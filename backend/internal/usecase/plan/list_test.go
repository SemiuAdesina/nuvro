package plan_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

type lister struct {
	trips     []trip.Trip
	requested int32
	after     *cursor.Position
	member    string
}

func (l *lister) ListForMember(_ context.Context, userID string, after *cursor.Position, limit int32) ([]trip.Trip, error) {
	l.member = userID
	l.after = after
	l.requested = limit
	if int(limit) < len(l.trips) {
		return l.trips[:limit], nil
	}
	return l.trips, nil
}

func trips(count int) []trip.Trip {
	created := make([]trip.Trip, 0, count)
	for index := range count {
		created = append(created, trip.Trip{
			ID:        string(rune('a'+index)) + "0000000-0000-4000-8000-000000000000",
			CreatedAt: time.Date(2026, 9, 9, 12, 0, index, 0, time.UTC),
		})
	}
	return created
}

func codec(t *testing.T) cursor.Codec {
	t.Helper()
	value, err := cursor.New([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestListRequiresAPrincipal(t *testing.T) {
	repository := &lister{}
	if _, err := plan.NewLister(repository, codec(t)).List(t.Context(), user.Principal{}, "", 10); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous listing")
	}
	if repository.requested != 0 {
		t.Fatal("an anonymous request reached the repository")
	}
}

func TestListReportsMorePagesWithoutRevealingTheExtraTrip(t *testing.T) {
	repository := &lister{trips: trips(4)}
	page, err := plan.NewLister(repository, codec(t)).List(t.Context(), user.Principal{UserID: "traveller"}, "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Trips) != 3 || !page.HasMore || page.NextCursor == "" {
		t.Fatalf("%+v", page)
	}
	if repository.requested != 4 {
		t.Fatalf("requested %d rows; the page probe is wrong", repository.requested)
	}
	if repository.member != "traveller" {
		t.Fatal("listing was not scoped to the principal")
	}
}

func TestFinalPageCarriesNoCursor(t *testing.T) {
	page, err := plan.NewLister(&lister{trips: trips(2)}, codec(t)).List(t.Context(), user.Principal{UserID: "traveller"}, "", 5)
	if err != nil || page.HasMore || page.NextCursor != "" || len(page.Trips) != 2 {
		t.Fatalf("%+v %v", page, err)
	}
}

func TestPageSizeIsBounded(t *testing.T) {
	for requested, expected := range map[int]int32{0: 20, -5: 20, 1: 1, 250: 100} {
		repository := &lister{}
		if _, err := plan.NewLister(repository, codec(t)).List(t.Context(), user.Principal{UserID: "traveller"}, "", requested); err != nil {
			t.Fatal(err)
		}
		if repository.requested != expected+1 {
			t.Fatalf("limit %d produced %d rows, want %d", requested, repository.requested, expected+1)
		}
	}
}

func TestCursorsAreAcceptedOnlyFromTheSameTraveller(t *testing.T) {
	repository := &lister{trips: trips(4)}
	subject := plan.NewLister(repository, codec(t))
	first, err := subject.List(t.Context(), user.Principal{UserID: "traveller"}, "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := subject.List(t.Context(), user.Principal{UserID: "traveller"}, first.NextCursor, 3); err != nil {
		t.Fatal(err)
	}
	if repository.after == nil || repository.after.ID != repository.trips[2].ID {
		t.Fatalf("cursor did not resume after the last returned trip: %+v", repository.after)
	}
	if _, err := subject.List(t.Context(), user.Principal{UserID: "another"}, first.NextCursor, 3); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("another traveller replayed the cursor")
	}
	if _, err := subject.List(t.Context(), user.Principal{UserID: "traveller"}, "forged", 3); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a forged cursor was accepted")
	}
}
