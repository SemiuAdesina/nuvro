package httperror_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

func TestStatusMapsEveryApplicationCode(t *testing.T) {
	cases := map[apperror.Code]int{
		apperror.CodeInvalidInput:    http.StatusBadRequest,
		apperror.CodeUnauthenticated: http.StatusUnauthorized,
		apperror.CodeForbidden:       http.StatusForbidden,
		apperror.CodeNotFound:        http.StatusNotFound,
		apperror.CodeConflict:        http.StatusConflict,
		apperror.CodeRateLimited:     http.StatusTooManyRequests,
		apperror.CodeUnavailable:     http.StatusServiceUnavailable,
		apperror.CodeInternal:        http.StatusInternalServerError,
	}

	for code, want := range cases {
		if got := httperror.Status(code); got != want {
			t.Errorf("Status(%q) = %d, want %d", code, got, want)
		}
	}
}

func TestStatusTreatsUnknownCodeAsInternal(t *testing.T) {
	if got := httperror.Status(apperror.Code("teapot")); got != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want 500", got)
	}
}

func TestWriteRendersClientErrorEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set(httperror.RequestIDHeader, "abc123")

	httperror.Write(recorder, apperror.New(apperror.CodeInvalidInput, "Departure date must be in the future."))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/problem+json") {
		t.Fatalf("Content-Type = %q", contentType)
	}

	decoded := decodeEnvelope(t, recorder)
	if decoded.Code != "validation_error" {
		t.Errorf("code = %q, want validation_error", decoded.Code)
	}
	if decoded.Detail != "Departure date must be in the future." {
		t.Errorf("message = %q", decoded.Detail)
	}
	if decoded.RequestID != "abc123" {
		t.Errorf("requestId = %q, want abc123", decoded.RequestID)
	}
}

func TestWriteHidesInternalDetailFromClients(t *testing.T) {
	recorder := httptest.NewRecorder()
	cause := errors.New("pq: password authentication failed for user \"nuvro\"")

	httperror.Write(recorder, apperror.Wrap(apperror.CodeInternal, "trip repository unavailable", cause))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	rendered := recorder.Body.String()
	for _, leak := range []string{"pq:", "password", "nuvro\\\"", "repository"} {
		if strings.Contains(rendered, leak) {
			t.Fatalf("body %q leaks %q", rendered, leak)
		}
	}
}

func TestWriteOmitsRequestIDWhenAbsent(t *testing.T) {
	recorder := httptest.NewRecorder()

	httperror.Write(recorder, apperror.New(apperror.CodeNotFound, "Trip not found."))

	if strings.Contains(recorder.Body.String(), "requestId") {
		t.Fatalf("body %q should omit an empty requestId", recorder.Body.String())
	}
}

func TestNotFoundHandlerReturnsStructuredEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()

	httperror.NotFoundHandler().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/unknown", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if decodeEnvelope(t, recorder).Code != "not_found" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestMethodNotAllowedHandlerReturnsStructuredEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()

	httperror.MethodNotAllowedHandler().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/health/live", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	if decodeEnvelope(t, recorder).Code != "method_not_allowed" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

type envelope struct {
	Code      string `json:"code"`
	Detail    string `json:"detail"`
	RequestID string `json:"requestId"`
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) envelope {
	t.Helper()
	var decoded envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, recorder.Body.String())
	}
	return decoded
}
