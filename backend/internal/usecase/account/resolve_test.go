package account_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

type directory struct {
	userID      string
	status      string
	created     []user.Account
	createError error
	lookups     int
	appearAfter int
}

func (d *directory) FindUserByIdentity(context.Context, string, string) (string, string, error) {
	d.lookups++
	if d.appearAfter > 0 && d.lookups < d.appearAfter {
		return "", "", nil
	}
	return d.userID, d.status, nil
}

func (d *directory) CreateAccount(_ context.Context, account user.Account) error {
	if d.createError != nil {
		return d.createError
	}
	d.created = append(d.created, account)
	return nil
}

type transaction struct{ calls int }

func (t *transaction) Within(ctx context.Context, operation func(context.Context) error) error {
	t.calls++
	return operation(ctx)
}

func TestExistingIdentityResolvesWithoutEnrolment(t *testing.T) {
	store := &directory{userID: "internal-uuid", status: user.StatusActive}
	boundary := &transaction{}
	principal, err := account.NewResolver(store, boundary, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: "001234.abc"})
	if err != nil || principal.UserID != "internal-uuid" || principal.Subject != "001234.abc" {
		t.Fatalf("%+v %v", principal, err)
	}
	if len(store.created) != 0 || boundary.calls != 0 {
		t.Fatal("an existing traveller was enrolled again")
	}
}

func TestFirstSignInEnrolsAtomicallyWithDefaults(t *testing.T) {
	store := &directory{userID: "internal-uuid", status: user.StatusActive, appearAfter: 2}
	boundary := &transaction{}
	principal, err := account.NewResolver(store, boundary, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: "001234.abc"})
	if err != nil || principal.UserID != "internal-uuid" {
		t.Fatalf("%+v %v", principal, err)
	}
	if len(store.created) != 1 || boundary.calls != 1 {
		t.Fatalf("enrolment was not a single atomic operation: %d accounts, %d transactions", len(store.created), boundary.calls)
	}
	created := store.created[0]
	if created.Provider != "apple" || created.Subject != "001234.abc" || created.UserID == created.Subject {
		t.Fatalf("provider subject leaked into the internal identifier: %+v", created)
	}
	if err := created.Profile.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := created.Preferences.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentFirstSignInRereadsInsteadOfFailing(t *testing.T) {
	store := &directory{
		userID: "internal-uuid", status: user.StatusActive, appearAfter: 2,
		createError: apperror.New(apperror.CodeConflict, "The resource already exists."),
	}
	principal, err := account.NewResolver(store, &transaction{}, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: "001234.abc"})
	if err != nil || principal.UserID != "internal-uuid" {
		t.Fatalf("a racing enrolment was not recovered: %+v %v", principal, err)
	}
}

func TestSuspendedAndAnonymousIdentitiesAreRefused(t *testing.T) {
	suspended := &directory{userID: "internal-uuid", status: user.StatusSuspended}
	if _, err := account.NewResolver(suspended, &transaction{}, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: "001234.abc"}); apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("a suspended account signed in: %v", err)
	}
	empty := &directory{}
	for _, identity := range []user.Identity{{}, {Provider: "apple"}, {Subject: "001234.abc"}} {
		if _, err := account.NewResolver(empty, &transaction{}, id.NewUUID()).Resolve(t.Context(), identity); err == nil {
			t.Fatalf("accepted %+v", identity)
		}
	}
}

func TestEnrolmentFailureIsReportedSafely(t *testing.T) {
	private := errors.New("duplicate key value violates unique constraint")
	store := &directory{createError: apperror.Wrap(apperror.CodeInternal, "", private), appearAfter: 99}
	_, err := account.NewResolver(store, &transaction{}, id.NewUUID()).
		Resolve(t.Context(), user.Identity{Provider: "apple", Subject: "001234.abc"})
	if err == nil {
		t.Fatal("enrolment failure was swallowed")
	}
	if apperror.SafeMessage(err) == private.Error() {
		t.Fatal("database detail reached the caller")
	}
}
