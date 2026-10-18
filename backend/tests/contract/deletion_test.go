package contract_test

import (
	"net/http"
	"testing"
	"time"
)

func TestClosingAnAccountIsAcceptedAndSchedulesErasure(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.call(t, http.MethodDelete, "/v1/me", `{"reason":"no_longer_travelling"}`, true)
	if response.Code != http.StatusAccepted {
		t.Fatalf("deletion returned %d %v", response.Code, decoded)
	}
	if decoded["status"] != "pending" {
		t.Fatalf("status = %v, want pending", decoded["status"])
	}
	requested, err := time.Parse(time.RFC3339, text(t, decoded, "requestedAt"))
	if err != nil {
		t.Fatal(err)
	}
	purgeAfter, err := time.Parse(time.RFC3339, text(t, decoded, "purgeAfter"))
	if err != nil {
		t.Fatal(err)
	}
	if !purgeAfter.After(requested) {
		t.Fatalf("purge after %s is not later than the request at %s", purgeAfter, requested)
	}
	if _, present := decoded["purgedAt"]; present {
		t.Fatalf("a pending deletion reported a purge time: %v", decoded)
	}
}

func TestClosingAnAccountWorksWithoutAReason(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.call(t, http.MethodDelete, "/v1/me", "", true)
	if response.Code != http.StatusAccepted {
		t.Fatalf("deletion returned %d %v", response.Code, decoded)
	}
	if len(environment.store.closed) != 1 {
		t.Fatalf("closed %v, want the authenticated account", environment.store.closed)
	}
}

func TestClosingAnAccountTwiceKeepsTheOriginalSchedule(t *testing.T) {
	environment := newEnvironment(t)
	_, first := environment.call(t, http.MethodDelete, "/v1/me", "", true)
	_, second := environment.call(t, http.MethodDelete, "/v1/me", "", true)
	if first["purgeAfter"] != second["purgeAfter"] {
		t.Fatalf("a repeated request moved erasure from %v to %v", first["purgeAfter"], second["purgeAfter"])
	}
}

func TestClosingAnAccountRequiresAuthentication(t *testing.T) {
	environment := newEnvironment(t)
	response, _ := environment.call(t, http.MethodDelete, "/v1/me", "", false)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("an anonymous deletion returned %d, want 401", response.Code)
	}
}

func TestAnOverLongDeletionReasonIsRefusedByTheContract(t *testing.T) {
	environment := newEnvironment(t)
	body := `{"reason":"` + longReason() + `"}`
	response, _ := environment.call(t, http.MethodDelete, "/v1/me", body, true)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("an over-long reason returned %d, want 400", response.Code)
	}
}

func longReason() string {
	value := make([]byte, 65)
	for index := range value {
		value[index] = 'a'
	}
	return string(value)
}

func TestLinkingAnAppleIdentityIsUnavailableUntilCredentialsExist(t *testing.T) {
	environment := newEnvironment(t)
	body := `{"authorizationCode":"c-1234567890"}`
	response, decoded := environment.call(t, http.MethodPut, "/v1/me/identity/apple", body, true)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("linking returned %d %v, want 503 while Apple credentials are unset", response.Code, decoded)
	}
}

func TestAnEmptyAuthorizationCodeIsRefusedByTheContract(t *testing.T) {
	environment := newEnvironment(t)
	response, _ := environment.call(t, http.MethodPut, "/v1/me/identity/apple", `{"authorizationCode":""}`, true)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("an empty code returned %d, want 400", response.Code)
	}
}

func TestLinkingAnAppleIdentityRequiresAuthentication(t *testing.T) {
	environment := newEnvironment(t)
	body := `{"authorizationCode":"c-1234567890"}`
	response, _ := environment.call(t, http.MethodPut, "/v1/me/identity/apple", body, false)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("an anonymous link returned %d, want 401", response.Code)
	}
}
