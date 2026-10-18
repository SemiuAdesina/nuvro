package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

const validationContract = `openapi: 3.1.0
info: {title: Validation test, version: '1'}
paths:
  /items/{itemId}:
    post:
      operationId: createItem
      parameters:
        - {name: itemId, in: path, required: true, schema: {type: integer, minimum: 1}}
        - {name: limit, in: query, required: true, schema: {type: integer, minimum: 1, maximum: 10}}
        - {name: X-Client, in: header, required: true, schema: {type: string, minLength: 3}}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              additionalProperties: false
              required: [name]
              properties:
                name: {type: string, minLength: 2}
                note: {type: [string, 'null']}
      responses:
        '204': {description: accepted}
`

func TestOpenAPIValidationBeforeHandler(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromData([]byte(validationContract))
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(t.Context()); err != nil {
		t.Fatal(err)
	}
	calls := 0
	handler := middleware.BodyLimit(128)(middleware.Validate(document, authentication.Validator(nil))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(204) })))
	tests := []struct {
		path, body, content, header string
		success                     bool
	}{
		{"/items/1?limit=3", `{"name":"ok"}`, "application/json", "ios", true},
		{"/items/1?limit=3", `{"name":"ok","note":null}`, "application/json", "ios", true},
		{"/items/1?limit=3", `{"name":"ok","note":"value"}`, "application/json", "ios", true},
		{"/items/0?limit=3", `{"name":"ok"}`, "application/json", "ios", false},
		{"/items/1", `{"name":"ok"}`, "application/json", "ios", false},
		{"/items/1?limit=11", `{"name":"ok"}`, "application/json", "ios", false},
		{"/items/1?limit=3", `{"name":"ok"}`, "application/json", "", false},
		{"/items/1?limit=3", `{"name":1}`, "application/json", "ios", false},
		{"/items/1?limit=3", `{"name":"ok","unexpected":true}`, "application/json", "ios", false},
		{"/items/1?limit=3", `{}`, "application/json", "ios", false},
		{"/items/1?limit=3", `{"name":"ok"}`, "text/plain", "ios", false},
		{"/items/1?limit=3", strings.Repeat("a", 129), "application/json", "ios", false},
	}
	for _, test := range tests {
		before := calls
		request := httptest.NewRequestWithContext(t.Context(), "POST", test.path, strings.NewReader(test.body))
		request.Header.Set("Content-Type", test.content)
		request.Header.Set("X-Client", test.header)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if (response.Code == 204) != test.success {
			t.Fatalf("%+v status %d body %s", test, response.Code, response.Body.String())
		}
		if !test.success && (calls != before || response.Header().Get("Content-Type") != "application/problem+json") {
			t.Fatal("invalid request reached handler or leaked raw error")
		}
	}
}

func TestSecurityRequirementsFailClosed(t *testing.T) {
	source := validationContract + `security:
  - bearerAuth: []
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer}
`
	document, err := openapi3.NewLoader().LoadFromData([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	handler := middleware.Validate(document, authentication.Validator(nil))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("unauthenticated request reached use case") }))
	request := httptest.NewRequestWithContext(t.Context(), "POST", "/items/1?limit=3", strings.NewReader(`{"name":"ok"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Client", "ios")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
}
