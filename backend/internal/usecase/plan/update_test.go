package plan_test

import (
	"context"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

type editable struct {
	stored    trip.Trip
	role      trip.Role
	roleError error
	saves     int
	insideTx  bool
}

type txMarker struct{}

func (e *editable) Get(context.Context, string) (trip.Trip, error) { return e.stored, nil }

func (e *editable) Role(context.Context, string, string) (trip.Role, error) {
	return e.role, e.roleError
}

func (e *editable) Save(ctx context.Context, value trip.Trip) error {
	e.insideTx = ctx.Value(txMarker{}) == true
	e.stored = value
	e.saves++
	return nil
}

type editTransaction struct{ calls int }

func (t *editTransaction) Within(ctx context.Context, operation func(context.Context) error) error {
	t.calls++
	return operation(context.WithValue(ctx, txMarker{}, true))
}

func stored() trip.Trip {
	return trip.Trip{
		ID: "trip", OwnerID: "traveller", Adults: 2, Children: 0, Currency: "EUR", Status: trip.Draft,
		CreatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestUpdateRequiresEditRights(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	for role, expected := range map[trip.Role]apperror.Code{
		trip.Owner:  "",
		trip.Editor: "",
		trip.Viewer: apperror.CodeForbidden,
	} {
		repository := &editable{stored: stored(), role: role}
		updater := plan.NewUpdater(repository, &editTransaction{}, clock.NewFixed(now))
		_, err := updater.Update(t.Context(), user.Principal{UserID: "traveller"}, "trip", trip.Patch{Adults: change.Set(int32(3))})
		if expected == "" {
			if err != nil || repository.stored.Adults != 3 {
				t.Fatalf("%s could not edit: %v", role, err)
			}
			continue
		}
		if apperror.CodeOf(err) != expected {
			t.Fatalf("%s received %v", role, err)
		}
		if repository.saves != 0 {
			t.Fatalf("%s wrote to the trip", role)
		}
	}
}

func TestUpdateHidesTripsTheTravellerCannotSee(t *testing.T) {
	repository := &editable{stored: stored(), roleError: apperror.New(apperror.CodeNotFound, "Trip not found.")}
	updater := plan.NewUpdater(repository, &editTransaction{}, clock.NewSystem())
	_, err := updater.Update(t.Context(), user.Principal{UserID: "stranger"}, "trip", trip.Patch{})
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("a non-member learned the trip exists: %v", err)
	}
	if _, err := updater.Update(t.Context(), user.Principal{}, "trip", trip.Patch{}); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous update")
	}
	if repository.saves != 0 {
		t.Fatal("an unauthorised update was written")
	}
}

func TestUpdateAppliesInsideOneTransactionAndStampsTheClock(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	repository := &editable{stored: stored(), role: trip.Owner}
	boundary := &editTransaction{}
	updated, err := plan.NewUpdater(repository, boundary, clock.NewFixed(now)).
		Update(t.Context(), user.Principal{UserID: "traveller"}, "trip", trip.Patch{Currency: change.Set("USD")})
	if err != nil || updated.Currency != "USD" || !updated.UpdatedAt.Equal(now) {
		t.Fatalf("%+v %v", updated, err)
	}
	if boundary.calls != 1 || !repository.insideTx {
		t.Fatal("the authorisation, read and write did not share one transaction")
	}
	if !updated.CreatedAt.Equal(stored().CreatedAt) {
		t.Fatal("createdAt was rewritten")
	}
}

func TestInvalidPatchesNeverReachTheRepository(t *testing.T) {
	repository := &editable{stored: stored(), role: trip.Owner}
	_, err := plan.NewUpdater(repository, &editTransaction{}, clock.NewSystem()).
		Update(t.Context(), user.Principal{UserID: "traveller"}, "trip", trip.Patch{Currency: change.Clear[string]()})
	if apperror.CodeOf(err) != apperror.CodeInvalidInput || repository.saves != 0 {
		t.Fatalf("clearing a required field was written: %v", err)
	}
}
