package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

func TestEveryResponseCarriesTheSecurityHeaders(t *testing.T) {
	response := httptest.NewRecorder()
	middleware.SecurityHeaders(ok()).
		ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil))

	for header, want := range map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"Cache-Control":             "no-store",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'none'; frame-ancestors 'none'",
		"X-Frame-Options":           "DENY",
	} {
		if got := response.Header().Get(header); got != want {
			t.Fatalf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestNoCrossOriginHeadersAreOffered(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil)
	request.Header.Set("Origin", "https://example.com")
	middleware.SecurityHeaders(ok()).ServeHTTP(response, request)

	for _, header := range []string{
		"Access-Control-Allow-Origin", "Access-Control-Allow-Credentials",
		"Access-Control-Allow-Methods", "Access-Control-Allow-Headers",
	} {
		if value := response.Header().Get(header); value != "" {
			t.Fatalf("%s = %q, want no cross-origin access for a native-client API", header, value)
		}
	}
}
