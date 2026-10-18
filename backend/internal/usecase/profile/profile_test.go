package profile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/profile"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

type repository struct {
	profile     user.Profile
	preferences user.Preferences
	saved       int
	readError   error
	insideTx    bool
	inTx        bool
}

type key struct{}

func (r *repository) Profile(ctx context.Context, _ string) (user.Profile, error) {
	r.insideTx = ctx.Value(key{}) == true
	return r.profile, r.readError
}

func (r *repository) SaveProfile(ctx context.Context, value user.Profile) error {
	r.inTx = ctx.Value(key{}) == true
	r.profile = value
	r.saved++
	return nil
}

func (r *repository) Preferences(context.Context, string) (user.Preferences, error) {
	return r.preferences, r.readError
}

func (r *repository) SavePreferences(_ context.Context, _ string, value user.Preferences) error {
	r.preferences = value
	r.saved++
	return nil
}

type transaction struct{ calls int }

func (t *transaction) Within(ctx context.Context, operation func(context.Context) error) error {
	t.calls++
	return operation(context.WithValue(ctx, key{}, true))
}

func newService(t *testing.T) (*profile.Service, *repository, *transaction) {
	t.Helper()
	store := &repository{profile: user.DefaultProfile(traveller), preferences: user.DefaultPreferences()}
	boundary := &transaction{}
	return profile.NewService(store, boundary), store, boundary
}

func TestReadsAndUpdatesRequireAPrincipal(t *testing.T) {
	service, store, _ := newService(t)
	anonymous := user.Principal{}
	if _, err := service.Profile(t.Context(), anonymous); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous profile read")
	}
	if _, err := service.UpdateProfile(t.Context(), anonymous, user.ProfilePatch{}); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous profile write")
	}
	if _, err := service.Preferences(t.Context(), anonymous); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous preference read")
	}
	if _, err := service.UpdatePreferences(t.Context(), anonymous, user.PreferencesPatch{}); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous preference write")
	}
	if store.saved != 0 {
		t.Fatal("an anonymous request reached the repository")
	}
}

func TestUpdateReadsAndWritesInsideOneTransaction(t *testing.T) {
	service, store, boundary := newService(t)
	principal := user.Principal{UserID: traveller}
	updated, err := service.UpdateProfile(t.Context(), principal, user.ProfilePatch{PreferredCurrency: change.Set("NGN")})
	if err != nil || updated.PreferredCurrency != "NGN" {
		t.Fatalf("%+v %v", updated, err)
	}
	if boundary.calls != 1 || !store.insideTx || !store.inTx {
		t.Fatal("the read and write did not share one transaction")
	}
	if store.profile.PreferredCurrency != "NGN" {
		t.Fatal("the update was not persisted")
	}
}

func TestInvalidPatchesAreRejectedWithoutWriting(t *testing.T) {
	service, store, _ := newService(t)
	principal := user.Principal{UserID: traveller}
	if _, err := service.UpdateProfile(t.Context(), principal, user.ProfilePatch{Units: change.Set("parsecs")}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("invalid units accepted")
	}
	if _, err := service.UpdatePreferences(t.Context(), principal, user.PreferencesPatch{Pace: change.Set("frantic")}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("invalid pace accepted")
	}
	if store.saved != 0 {
		t.Fatal("an invalid patch was written")
	}
}

func TestRepositoryFailuresPropagate(t *testing.T) {
	service, store, _ := newService(t)
	store.readError = errors.New("private database detail")
	principal := user.Principal{UserID: traveller}
	if _, err := service.Profile(t.Context(), principal); err == nil {
		t.Fatal("read failure was hidden")
	}
	if _, err := service.UpdatePreferences(t.Context(), principal, user.PreferencesPatch{}); err == nil {
		t.Fatal("write path ignored the read failure")
	}
	if store.saved != 0 {
		t.Fatal("wrote after a failed read")
	}
}
