package contract_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/SemiuAdesina/nuvro/backend/internal/contract"
)

func (e *environment) call(t *testing.T, method, path, body string, authenticated bool) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	return e.callWithKey(t, method, path, body, authenticated, "key-"+strconv.Itoa(e.calls))
}

func (e *environment) callWithKey(t *testing.T, method, path, body string, authenticated bool, key string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	e.calls++
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequestWithContext(t.Context(), method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		request.Header.Set("Authorization", "Bearer "+credential)
	}
	if method == http.MethodPost && key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	e.handler.ServeHTTP(response, request)
	decoded := map[string]any{}
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("%s %s returned unparseable body: %s", method, path, response.Body.String())
		}
	}
	e.validate(t, request, response)
	return response, decoded
}

func (e *environment) validate(t *testing.T, request *http.Request, response *httptest.ResponseRecorder) {
	t.Helper()
	document, err := contract.Load(t.Context(), "../../../contracts/openapi/v1/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	router, err := gorillamux.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	route, params, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route},
		Status:                 response.Code, Header: response.Header(),
		Options: &openapi3filter.Options{IncludeResponseStatus: true, AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	}
	input.SetBodyBytes(response.Body.Bytes())
	if err := openapi3filter.ValidateResponse(t.Context(), input); err != nil {
		t.Fatalf("%s %s response violates the contract: %v", request.Method, request.URL.Path, err)
	}
}

func text(t *testing.T, body map[string]any, field string) string {
	t.Helper()
	value, ok := body[field].(string)
	if !ok {
		t.Fatalf("%s is not a string in %v", field, body)
	}
	return value
}

func list(t *testing.T, body map[string]any, field string) []any {
	t.Helper()
	value, ok := body[field].([]any)
	if !ok {
		t.Fatalf("%s is not an array in %v", field, body)
	}
	return value
}

func member(t *testing.T, value any, field string) string {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%v is not an object", value)
	}
	return text(t, object, field)
}

func TestProductRoutesRequireAVerifiedPrincipal(t *testing.T) {
	environment := newEnvironment(t)
	for _, route := range [][2]string{
		{http.MethodGet, "/v1/me"},
		{http.MethodGet, "/v1/me/preferences"},
		{http.MethodGet, "/v1/trips"},
		{http.MethodGet, "/v1/trips/" + travellerID},
	} {
		response, body := environment.call(t, route[0], route[1], "", false)
		if response.Code != http.StatusUnauthorized || body["code"] != "authentication_required" {
			t.Fatalf("%s %s returned %d %v", route[0], route[1], response.Code, body)
		}
		if response.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatalf("%s %s did not answer with Problem Details", route[0], route[1])
		}
	}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil)
	request.Header.Set("Authorization", "Bearer not-the-configured-credential")
	response := httptest.NewRecorder()
	environment.handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("a wrong credential returned %d", response.Code)
	}
}

func TestProfilePatchDistinguishesAbsentNullAndValue(t *testing.T) {
	environment := newEnvironment(t)
	_, profile := environment.call(t, http.MethodGet, "/v1/me", "", true)
	if profile["userId"] != travellerID || profile["preferredCurrency"] != "USD" || profile["homeCountry"] != nil {
		t.Fatalf("unexpected profile: %v", profile)
	}
	_, updated := environment.call(t, http.MethodPatch, "/v1/me", `{"preferredCurrency":"NGN","homeCountry":"NG","units":"imperial"}`, true)
	if updated["preferredCurrency"] != "NGN" || updated["homeCountry"] != "NG" || updated["units"] != "imperial" {
		t.Fatalf("patch was not applied: %v", updated)
	}
	if updated["locale"] != "en" {
		t.Fatalf("omitted field changed: %v", updated)
	}
	_, cleared := environment.call(t, http.MethodPatch, "/v1/me", `{"homeCountry":null}`, true)
	if cleared["homeCountry"] != nil || cleared["preferredCurrency"] != "NGN" {
		t.Fatalf("explicit null did not clear only the named field: %v", cleared)
	}
	response, problem := environment.call(t, http.MethodPatch, "/v1/me", `{"preferredCurrency":null}`, true)
	if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("clearing a required field returned %d %v", response.Code, problem)
	}
}

func TestPreferencePatchHandlesNullableFlightTolerance(t *testing.T) {
	environment := newEnvironment(t)
	_, preferences := environment.call(t, http.MethodPatch, "/v1/me/preferences", `{"interests":["food","nature"],"maxFlightMinutes":600}`, true)
	if len(list(t, preferences, "interests")) != 2 || preferences["maxFlightMinutes"] != float64(600) {
		t.Fatalf("preferences patch failed: %v", preferences)
	}
	_, unset := environment.call(t, http.MethodPatch, "/v1/me/preferences", `{"maxFlightMinutes":null}`, true)
	if unset["maxFlightMinutes"] != nil || len(list(t, unset, "interests")) != 2 {
		t.Fatalf("preference null handling failed: %v", unset)
	}
}

func TestTripLifecycleIsOwnedByThePrincipal(t *testing.T) {
	environment := newEnvironment(t)
	response, created := environment.call(t, http.MethodPost, "/v1/trips", `{"adults":2,"children":1,"preferredCurrency":"EUR","destinationId":"lisbon"}`, true)
	if response.Code != http.StatusCreated || created["ownerId"] != travellerID || created["status"] != "draft" {
		t.Fatalf("create returned %d %v", response.Code, created)
	}
	identifier := text(t, created, "id")
	_, found := environment.call(t, http.MethodGet, "/v1/trips/"+identifier, "", true)
	if found["id"] != identifier || found["destinationId"] != "lisbon" {
		t.Fatalf("read returned %v", found)
	}
	_, patched := environment.call(t, http.MethodPatch, "/v1/trips/"+identifier, `{"destinationId":null,"adults":3}`, true)
	if patched["destinationId"] != nil || patched["adults"] != float64(3) || patched["preferredCurrency"] != "EUR" {
		t.Fatalf("patch returned %v", patched)
	}
	missing, notFound := environment.call(t, http.MethodGet, "/v1/trips/"+otherTravellerID, "", true)
	if missing.Code != http.StatusNotFound || notFound["code"] != "not_found" {
		t.Fatalf("another traveller's trip returned %d %v", missing.Code, notFound)
	}
}

func TestTripDatesChangeAsOneUnit(t *testing.T) {
	environment := newEnvironment(t)
	_, created := environment.call(t, http.MethodPost, "/v1/trips", `{"adults":1,"preferredCurrency":"EUR"}`, true)
	identifier := text(t, created, "id")
	_, scheduled := environment.call(t, http.MethodPatch, "/v1/trips/"+identifier, `{"startDate":"2026-11-02","endDate":"2026-11-09"}`, true)
	if scheduled["startDate"] != "2026-11-02" || scheduled["endDate"] != "2026-11-09" {
		t.Fatalf("dates were not applied: %v", scheduled)
	}
	for name, body := range map[string]string{
		"start only": `{"startDate":"2026-12-01"}`,
		"end only":   `{"endDate":"2026-12-09"}`,
		"reversed":   `{"startDate":"2026-12-09","endDate":"2026-12-01"}`,
	} {
		response, problem := environment.call(t, http.MethodPatch, "/v1/trips/"+identifier, body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
	_, cleared := environment.call(t, http.MethodPatch, "/v1/trips/"+identifier, `{"startDate":null,"endDate":null}`, true)
	if cleared["startDate"] != nil || cleared["endDate"] != nil {
		t.Fatalf("the range was not cleared: %v", cleared)
	}
}

func TestTripListingPagesWithOpaqueCursors(t *testing.T) {
	environment := newEnvironment(t)
	for range 3 {
		environment.call(t, http.MethodPost, "/v1/trips", `{"adults":1,"preferredCurrency":"USD"}`, true)
	}
	_, first := environment.call(t, http.MethodGet, "/v1/trips?limit=2", "", true)
	items := list(t, first, "items")
	if len(items) != 2 || first["hasMore"] != true || first["nextCursor"] == nil {
		t.Fatalf("first page: %v", first)
	}
	token := text(t, first, "nextCursor")
	if strings.Contains(token, member(t, items[1], "id")) {
		t.Fatal("the cursor exposed a database identifier")
	}
	_, second := environment.call(t, http.MethodGet, "/v1/trips?limit=2&cursor="+token, "", true)
	if len(list(t, second, "items")) != 1 || second["hasMore"] != false {
		t.Fatalf("second page: %v", second)
	}
	response, problem := environment.call(t, http.MethodGet, "/v1/trips?cursor=tampered.cursor", "", true)
	if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("a forged cursor returned %d %v", response.Code, problem)
	}
}

func TestRequestValidationRunsBeforeTheUseCase(t *testing.T) {
	environment := newEnvironment(t)
	for name, body := range map[string]string{
		"unknown field":    `{"adults":1,"preferredCurrency":"USD","surprise":true}`,
		"missing required": `{"children":1}`,
		"wrong type":       `{"adults":"two","preferredCurrency":"USD"}`,
		"bad currency":     `{"adults":1,"preferredCurrency":"usd"}`,
		"too many adults":  `{"adults":101,"preferredCurrency":"USD"}`,
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/trips", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
		if len(environment.store.trips) != 0 {
			t.Fatalf("%s reached the use case", name)
		}
	}
}
