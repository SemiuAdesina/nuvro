package http_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	nuvrohttp "github.com/SemiuAdesina/nuvro/backend/internal/transport/http"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/health"
)

func newTestRouter() http.Handler {
	timer := clock.NewFixed(time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC))
	return nuvrohttp.NewRouter(nuvrohttp.RouterDependencies{
		Logger:      slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
		Clock:       timer,
		Identifiers: id.NewRandom(),
		Health:      health.NewHandler(timer),
	})
}

func TestRouterServesLivenessProbe(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"live"`) {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestRouterServesReadinessProbe(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ready"`) {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestRouterStampsEveryResponseWithARequestIdentifier(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("no request identifier on the response")
	}
}

func TestRouterReturnsStructuredNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/destinations", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if code := decodeErrorCode(t, recorder); code != "not_found" {
		t.Fatalf("code = %q, want not_found", code)
	}
}

func TestRouterReturnsStructuredMethodNotAllowed(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/health/live", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	if code := decodeErrorCode(t, recorder); code != "method_not_allowed" {
		t.Fatalf("code = %q, want method_not_allowed", code)
	}
}

func TestRouterErrorsCarryTheRequestIdentifier(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/nowhere", nil))

	var decoded struct {
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, recorder.Body.String())
	}
	if decoded.RequestID != recorder.Header().Get("X-Request-Id") {
		t.Fatalf("body identifier %q does not match header %q", decoded.RequestID, recorder.Header().Get("X-Request-Id"))
	}
}

func decodeErrorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var decoded struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, recorder.Body.String())
	}
	return decoded.Code
}
