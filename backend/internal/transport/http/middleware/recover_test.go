package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

func TestRecoverConvertsPanicIntoStructuredInternalError(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	handler := middleware.Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("nil map write in trip assembly")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/trips", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}

	var decoded struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, recorder.Body.String())
	}
	if decoded.Code != "internal_error" {
		t.Errorf("code = %q, want internal", decoded.Code)
	}
	if strings.Contains(recorder.Body.String(), "nil map write") {
		t.Error("the panic value leaked to the client")
	}
	if strings.Contains(recorder.Body.String(), "goroutine") {
		t.Error("a stack trace leaked to the client")
	}
}

func TestRecoverLogsPanicWithRouteContext(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	handler := middleware.Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("nil map write in trip assembly")
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/trips", nil))

	recorded := logs.String()
	for _, fragment := range []string{"panic recovered", "nil map write", "/v1/trips", "POST", "stack"} {
		if !strings.Contains(recorded, fragment) {
			t.Errorf("log %q does not contain %q", recorded, fragment)
		}
	}
}

func TestRecoverLeavesSuccessfulRequestsUntouched(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	handler := middleware.Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("queued"))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/trips", nil))

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", recorder.Code)
	}
	if recorder.Body.String() != "queued" {
		t.Fatalf("body = %q, want queued", recorder.Body.String())
	}
}

func TestRecoverRepanicsOnAbortHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	handler := middleware.Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		recovered := recover()
		failure, isError := recovered.(error)
		if !isError || !errors.Is(failure, http.ErrAbortHandler) {
			t.Fatalf("recovered = %v, want http.ErrAbortHandler", recovered)
		}
	}()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/trips", nil))
	t.Fatal("ErrAbortHandler was swallowed")
}
