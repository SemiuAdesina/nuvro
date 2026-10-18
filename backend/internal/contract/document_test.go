package contract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/SemiuAdesina/nuvro/backend/internal/contract"
)

func TestLoadAndBundle(t *testing.T) {
	source := "../../../contracts/openapi/v1/openapi.yaml"
	raw, err := contract.Bundle(t.Context(), source)
	if err != nil {
		t.Fatal(err)
	}
	loader := openapi3.NewLoader()
	document, err := loader.LoadFromData(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(t.Context()); err != nil {
		t.Fatal(err)
	}
	if document.Paths.Len() != 35 {
		t.Fatalf("released paths = %d, want 35", document.Paths.Len())
	}
	if len(document.Components.SecuritySchemes) != 1 || document.Components.SecuritySchemes["bearerAuth"].Value.Scheme != "bearer" {
		t.Fatal("bearer security scheme missing from the bundle")
	}
}

func TestProductOperationsInheritAuthenticationAndHealthOptsOut(t *testing.T) {
	document, err := contract.Load(t.Context(), "../../../contracts/openapi/v1/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Security) == 0 {
		t.Fatal("the document declares no security requirement")
	}
	for _, path := range []string{"/v1/me", "/v1/me/preferences", "/v1/trips", "/v1/trips/{tripId}", "/v1/discover/matches", "/v1/destinations/{destinationId}", "/v1/trips/{tripId}/itinerary", "/v1/trips/{tripId}/budget", "/v1/trips/{tripId}/nearby/search", "/v1/trips/{tripId}/active", "/v1/trips/{tripId}/replan", "/v1/trips/{tripId}/alerts", "/v1/trips/{tripId}/translate", "/v1/trips/{tripId}/offline-pack", "/v1/trips/{tripId}/flights/search", "/v1/trips/{tripId}/stays/search", "/v1/commerce/handoffs", "/v1/me/price-watches"} {
		item := document.Paths.Find(path)
		if item == nil {
			t.Fatalf("%s is not released", path)
		}
		for method, operation := range item.Operations() {
			if operation.Security != nil {
				t.Fatalf("%s %s overrides the document security requirement", method, path)
			}
		}
	}
	for _, path := range []string{"/health/live", "/health/ready"} {
		for method, operation := range document.Paths.Find(path).Operations() {
			if operation.Security == nil || len(*operation.Security) != 0 {
				t.Fatalf("%s %s does not opt out of authentication", method, path)
			}
		}
	}
	assertSessionEndpoints(t, document)
}

func assertSessionEndpoints(t *testing.T, document *openapi3.T) {
	t.Helper()
	for path, method := range map[string]string{"/v1/auth/sessions": "POST", "/v1/auth/sessions/refresh": "POST"} {
		operation := document.Paths.Find(path).GetOperation(method)
		if operation.Security == nil || len(*operation.Security) != 0 {
			t.Fatalf("%s %s must be reachable before a session exists", method, path)
		}
	}
	if signOut := document.Paths.Find("/v1/auth/sessions").Delete; signOut.Security != nil {
		t.Fatal("DELETE /v1/auth/sessions must inherit authentication")
	}
}

func TestLoadRejectsInvalidContract(t *testing.T) {
	for _, source := range []string{
		"openapi: 3.0.3\ninfo: {title: test, version: '1'}\npaths: {}\n",
		"openapi: 3.1.0\ninfo: {title: test, version: '1'}\npaths:\n  /broken:\n    $ref: './missing.yaml'\n",
		"openapi: 3.1.0\ninfo: {title: test, version: '1'}\npaths:\n  /broken:\n    get: {}\n",
	} {
		path := filepath.Join(t.TempDir(), "openapi.yaml")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := contract.Load(t.Context(), path); err == nil {
			t.Fatal("accepted invalid source")
		}
	}
}
