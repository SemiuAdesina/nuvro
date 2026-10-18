package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/SemiuAdesina/nuvro/backend/internal/contract"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	nuvrohttp "github.com/SemiuAdesina/nuvro/backend/internal/transport/http"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/health"
)

func TestReleasedContractAndResponses(t *testing.T) {
	document, err := contract.Load(t.Context(), "../../../contracts/openapi/v1/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	router, err := gorillamux.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	timer := clock.NewSystem()
	handler := nuvrohttp.NewRouter(nuvrohttp.RouterDependencies{
		Logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), Clock: timer, Identifiers: id.NewRandom(), Health: health.NewHandler(timer),
	})
	for _, path := range []string{"/health/live", "/health/ready"} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		route, params, err := router.FindRoute(request)
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		input := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route},
			Status:                 response.Code, Header: response.Header(), Options: &openapi3filter.Options{IncludeResponseStatus: true},
		}
		input.SetBodyBytes(response.Body.Bytes())
		if err := openapi3filter.ValidateResponse(t.Context(), input); err != nil {
			t.Fatal(err)
		}
	}
	if document.Paths.Len() != 35 {
		t.Fatalf("released paths = %d, want 35", document.Paths.Len())
	}
}

func TestProblemResponsesConform(t *testing.T) {
	document, err := api.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	timer := clock.NewSystem()
	handler := nuvrohttp.NewRouter(nuvrohttp.RouterDependencies{Logger: slog.Default(), Clock: timer, Identifiers: id.NewRandom(), Health: health.NewHandler(timer)})
	for _, test := range []struct {
		method, path string
		status       int
	}{{"GET", "/missing", 404}, {"POST", "/health/live", 405}} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), test.method, test.path, nil))
		var value any
		if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		if err := document.Components.Schemas["Problem"].Value.VisitJSON(value); err != nil {
			t.Fatal(err)
		}
		if response.Code != test.status || response.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatal(response)
		}
	}
}

func TestCommonSchemasRejectInvalidValues(t *testing.T) {
	document, err := api.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, payload string }{
		{"Money", `{"amountMinor":1.2,"currency":"USD"}`},
		{"GeoPoint", `{"latitude":91,"longitude":0}`},
		{"Freshness", `{"observedAt":"2026-09-08T00:00:00Z","validUntil":"2026-09-09T00:00:00Z","status":"unknown"}`},
		{"MatchScore", `{"score":101,"components":[],"summary":"","weightsVersion":"1","evidenceVersion":"1"}`},
	} {
		var value any
		if err := json.Unmarshal([]byte(test.payload), &value); err != nil {
			t.Fatal(err)
		}
		if err := document.Components.Schemas[test.name].Value.VisitJSON(value); err == nil {
			t.Errorf("%s accepted invalid data", test.name)
		}
	}
}

func TestBundledDocumentValidates(t *testing.T) {
	document, err := api.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}
