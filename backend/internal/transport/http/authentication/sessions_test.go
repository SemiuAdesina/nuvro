package authentication_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
)

type verifier struct {
	identity user.Identity
	err      error
	seen     string
}

func (v *verifier) Verify(_ context.Context, credential string) (user.Identity, error) {
	v.seen = credential
	return v.identity, v.err
}

type accounts struct {
	principal user.Principal
	err       error
	seen      user.Identity
}

func (a *accounts) Resolve(_ context.Context, identity user.Identity) (user.Principal, error) {
	a.seen = identity
	return a.principal, a.err
}

func TestSessionsMapVerifiedIdentityToInternalPrincipal(t *testing.T) {
	checker := &verifier{identity: user.Identity{Provider: "apple", Subject: "001234.abc"}}
	directory := &accounts{principal: user.Principal{UserID: "internal-uuid", Subject: "001234.abc", Provider: "apple"}}
	principal, err := authentication.NewSessions(checker, directory).Authenticate(t.Context(), "token")
	if err != nil || principal.UserID != "internal-uuid" {
		t.Fatalf("%+v %v", principal, err)
	}
	if checker.seen != "token" || directory.seen.Subject != "001234.abc" {
		t.Fatal("credential or identity was not passed through")
	}
	if principal.UserID == principal.Subject {
		t.Fatal("provider subject was used as the internal identifier")
	}
}

func TestSessionsFailClosedAndHideProviderDetail(t *testing.T) {
	rejected := errors.New("apple: private diagnostic")
	_, err := authentication.NewSessions(&verifier{err: rejected}, &accounts{}).Authenticate(t.Context(), "token")
	if err == nil || errors.Is(err, rejected) {
		t.Fatal("verification failure leaked to the caller")
	}
	suspended := errors.New("account suspended")
	if _, err := authentication.NewSessions(&verifier{}, &accounts{err: suspended}).Authenticate(t.Context(), "token"); err == nil {
		t.Fatal("account failure was ignored")
	}
	if authentication.NewSessions(nil, &accounts{}) != nil || authentication.NewSessions(&verifier{}, nil) != nil {
		t.Fatal("incomplete session dependencies produced a usable authenticator")
	}
}
