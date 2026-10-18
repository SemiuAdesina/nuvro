package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

func TestRequestIDSetsHeaderAndContext(t *testing.T) {
	var observed string
	handler := middleware.RequestID(id.NewRandom())(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		observed = middleware.RequestIDFrom(r.Context())
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	header := recorder.Header().Get(httperror.RequestIDHeader)
	if header == "" {
		t.Fatal("response carries no request identifier header")
	}
	if observed != header {
		t.Fatalf("context identifier %q does not match header %q", observed, header)
	}
}

func TestRequestIDIgnoresClientSuppliedIdentifier(t *testing.T) {
	handler := middleware.RequestID(id.NewRandom())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil)
	request.Header.Set(httperror.RequestIDHeader, "spoofed-by-caller")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get(httperror.RequestIDHeader) == "spoofed-by-caller" {
		t.Fatal("the caller supplied identifier was echoed back")
	}
}

func TestRequestIDIssuesDistinctIdentifiersPerRequest(t *testing.T) {
	handler := middleware.RequestID(id.NewRandom())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	first := httptest.NewRecorder()
	second := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))
	handler.ServeHTTP(second, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if first.Header().Get(httperror.RequestIDHeader) == second.Header().Get(httperror.RequestIDHeader) {
		t.Fatal("two requests received the same identifier")
	}
}

func TestRequestIDStillServesWhenGenerationFails(t *testing.T) {
	failing := id.NewRandomFrom(func([]byte) (int, error) { return 0, errors.New("entropy unavailable") })

	served := false
	handler := middleware.RequestID(failing)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if !served {
		t.Fatal("the request was not served when identifier generation failed")
	}
	if recorder.Header().Get(httperror.RequestIDHeader) == "" {
		t.Fatal("fallback identifier is missing")
	}
}

func TestRequestIDFromReturnsEmptyStringWithoutMiddleware(t *testing.T) {
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil)

	if got := middleware.RequestIDFrom(request.Context()); got != "" {
		t.Fatalf("RequestIDFrom = %q, want an empty string", got)
	}
}
