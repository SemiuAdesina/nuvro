package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

func TestBodylessOperationsRejectChunkedBodies(t *testing.T) {
	calls := 0
	handler := middleware.BodyLimit(10)(middleware.NoBody(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(204) })))
	request := httptest.NewRequestWithContext(t.Context(), "GET", "/health/live", strings.NewReader("hidden body"))
	request.ContentLength = -1
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 400 || calls != 0 {
		t.Fatal("body accepted")
	}
	request = httptest.NewRequestWithContext(t.Context(), "GET", "/health/live", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 204 {
		t.Fatal(response)
	}
}
