package plan_test

import (
	"context"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

type reader struct {
	role  trip.Role
	reads int
}

func (r *reader) Get(context.Context, string) (trip.Trip, error) {
	r.reads++
	return trip.Trip{ID: "trip"}, nil
}
func (r *reader) Role(context.Context, string, string) (trip.Role, error) { return r.role, nil }
func TestReadAuthorizesBeforeLoading(t *testing.T) {
	store := &reader{}
	operation := plan.NewReader(store)
	if _, err := operation.Read(t.Context(), user.Principal{}, "trip"); err == nil {
		t.Fatal("anonymous read")
	}
	if _, err := operation.Read(t.Context(), user.Principal{UserID: "user"}, "trip"); err == nil {
		t.Fatal("nonmember read")
	}
	if store.reads != 0 {
		t.Fatal("loaded unauthorized trip")
	}
	for _, role := range []trip.Role{trip.Owner, trip.Editor, trip.Viewer} {
		store.role = role
		if _, err := operation.Read(t.Context(), user.Principal{UserID: "user"}, "trip"); err != nil {
			t.Fatal(err)
		}
	}
}
