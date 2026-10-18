package contract_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (e *environment) callAs(t *testing.T, method, path, body, bearer string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	response := httptest.NewRecorder()
	e.handler.ServeHTTP(response, request)
	decoded := map[string]any{}
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("%s %s returned a non-JSON body: %s", method, path, response.Body.String())
		}
	}
	return response, decoded
}

func openSession(t *testing.T, environment *environment) map[string]any {
	t.Helper()
	body := `{"identityToken":"` + credential + `","authorizationCode":"c-1234567890"}`
	response, decoded := environment.callAs(t, http.MethodPost, "/v1/auth/sessions", body, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("opening a session returned %d %v", response.Code, decoded)
	}
	return decoded
}

func TestAVerifiedIdentityBecomesASessionThatReachesTheProduct(t *testing.T) {
	environment := newEnvironment(t)
	granted := openSession(t, environment)
	access, refresh := text(t, granted, "accessToken"), text(t, granted, "refreshToken")
	if access == "" || refresh == "" || text(t, granted, "accountId") != travellerID {
		t.Fatalf("session = %v", granted)
	}
	if strings.Count(access, ".") != 2 {
		t.Fatalf("access token = %q, want a compact JWT", access)
	}
	response, decoded := environment.callAs(t, http.MethodGet, "/v1/me", "", access)
	if response.Code != http.StatusOK {
		t.Fatalf("the access token was refused: %d %v", response.Code, decoded)
	}
}

func TestARejectedCredentialGetsNoSession(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.callAs(t, http.MethodPost, "/v1/auth/sessions", `{"identityToken":"not-the-credential"}`, "")
	if response.Code != http.StatusUnauthorized || decoded["code"] != "authentication_required" {
		t.Fatalf("a bad credential returned %d %v", response.Code, decoded)
	}
	empty, _ := environment.callAs(t, http.MethodPost, "/v1/auth/sessions", `{"identityToken":""}`, "")
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("an empty credential returned %d, want the contract to refuse it", empty.Code)
	}
}

func TestRefreshRotatesAndReuseRevokesTheFamily(t *testing.T) {
	environment := newEnvironment(t)
	first := openSession(t, environment)
	response, second := environment.callAs(t, http.MethodPost, "/v1/auth/sessions/refresh",
		`{"refreshToken":"`+text(t, first, "refreshToken")+`"}`, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("refresh returned %d %v", response.Code, second)
	}
	if second["refreshToken"] == first["refreshToken"] || second["accessToken"] == first["accessToken"] {
		t.Fatal("refresh returned the same tokens")
	}
	reused, _ := environment.callAs(t, http.MethodPost, "/v1/auth/sessions/refresh",
		`{"refreshToken":"`+text(t, first, "refreshToken")+`"}`, "")
	if reused.Code != http.StatusUnauthorized {
		t.Fatalf("a used refresh token was accepted again: %d", reused.Code)
	}
	after, _ := environment.callAs(t, http.MethodGet, "/v1/me", "", text(t, second, "accessToken"))
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("the family survived a replayed refresh token: %d", after.Code)
	}
}

func TestSigningOutEndsTheSessionImmediately(t *testing.T) {
	environment := newEnvironment(t)
	granted := openSession(t, environment)
	access := text(t, granted, "accessToken")
	response, _ := environment.callAs(t, http.MethodDelete, "/v1/auth/sessions", "", access)
	if response.Code != http.StatusNoContent {
		t.Fatalf("sign out returned %d", response.Code)
	}
	after, _ := environment.callAs(t, http.MethodGet, "/v1/me", "", access)
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("a signed-out session still reached the product: %d", after.Code)
	}
	anonymous, _ := environment.callAs(t, http.MethodDelete, "/v1/auth/sessions", "", "")
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("an anonymous sign out returned %d, want 401", anonymous.Code)
	}
}

func TestClosingTheAccountKillsItsSessions(t *testing.T) {
	environment := newEnvironment(t)
	granted := openSession(t, environment)
	access := text(t, granted, "accessToken")
	response, _ := environment.callAs(t, http.MethodDelete, "/v1/me", "", access)
	if response.Code != http.StatusAccepted {
		t.Fatalf("closing the account returned %d", response.Code)
	}
	after, _ := environment.callAs(t, http.MethodGet, "/v1/me", "", access)
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("a closed account still reached the product with its session: %d", after.Code)
	}
}
