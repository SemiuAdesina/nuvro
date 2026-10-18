package authentication_test

import (
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
)

const credential = "synthetic-development-credential"

func TestDevelopmentVerifierRefusesToExistOutsideDevelopment(t *testing.T) {
	if _, err := authentication.NewDevelopmentVerifier(false, credential, "local-traveller"); !errors.Is(err, authentication.ErrDevelopmentUnavailable) {
		t.Fatal("development verifier was built outside development")
	}
}

func TestDevelopmentVerifierRequiresAConfiguredCredential(t *testing.T) {
	for _, test := range [][2]string{{"", "subject"}, {"short", "subject"}, {credential, ""}} {
		if _, err := authentication.NewDevelopmentVerifier(true, test[0], test[1]); err == nil {
			t.Fatalf("accepted credential %q subject %q", test[0], test[1])
		}
	}
}

func TestDevelopmentVerifierMatchesOnlyTheConfiguredCredential(t *testing.T) {
	verifier, err := authentication.NewDevelopmentVerifier(true, credential, "local-traveller")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := verifier.Verify(t.Context(), credential)
	if err != nil || identity.Provider != authentication.DevelopmentProvider || identity.Subject != "local-traveller" {
		t.Fatalf("%+v %v", identity, err)
	}
	for _, wrong := range []string{"", "another-credential-entirely", credential + "x"} {
		if _, err := verifier.Verify(t.Context(), wrong); err == nil {
			t.Fatalf("accepted %q", wrong)
		}
	}
}
