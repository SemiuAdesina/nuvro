package authentication_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
)

type authenticator struct{ fail bool }

func (a authenticator) Authenticate(context.Context, string) (user.Principal, error) {
	if a.fail {
		return user.Principal{}, errors.New("secret provider failure")
	}
	return user.Principal{UserID: "internal", Subject: "external", Provider: "oidc"}, nil
}

func TestAuthenticationFailsClosed(t *testing.T) {
	for _, test := range []struct {
		auth    authentication.Authenticator
		header  string
		success bool
	}{
		{nil, "Bearer token", false}, {authenticator{}, "", false}, {authenticator{}, "Basic token", false},
		{authenticator{fail: true}, "Bearer token", false}, {authenticator{}, "Bearer token", true},
	} {
		request := httptest.NewRequestWithContext(t.Context(), "GET", "/protected", nil)
		request.Header.Set("Authorization", test.header)
		input := &openapi3filter.AuthenticationInput{RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request}}
		err := authentication.Validator(test.auth)(request.Context(), input)
		if (err == nil) != test.success {
			t.Fatalf("header %q: %v", test.header, err)
		}
		principal, ok := authentication.From(request.Context())
		if ok != test.success {
			t.Fatal("principal context mismatch")
		}
		if ok && principal.UserID != "internal" {
			t.Fatal("external subject used as user id")
		}
	}
}
