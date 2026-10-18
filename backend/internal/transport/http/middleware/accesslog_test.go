package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

type slowClock struct {
	fixed *clock.Fixed
	step  time.Duration
}

func (c *slowClock) Now() time.Time {
	now := c.fixed.Now()
	c.fixed.Advance(c.step)
	return now
}

func TestAccessLogRecordsOutcomeOfHandledRequest(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	timer := &slowClock{fixed: clock.NewFixed(time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)), step: 250 * time.Millisecond}

	handler := middleware.AccessLog(logger, timer)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"t1"}`))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/trips", nil))

	record := decodeLogRecord(t, logs.Bytes())
	if record["msg"] != "http request" {
		t.Errorf("msg = %v", record["msg"])
	}
	if record["method"] != "POST" {
		t.Errorf("method = %v, want POST", record["method"])
	}
	if record["route"] != "unmatched" {
		t.Errorf("route = %v, want unmatched", record["route"])
	}
	if record["status"] != float64(http.StatusCreated) {
		t.Errorf("status = %v, want 201", record["status"])
	}
	if record["bytes"] != float64(len(`{"id":"t1"}`)) {
		t.Errorf("bytes = %v, want %d", record["bytes"], len(`{"id":"t1"}`))
	}
	if record["duration"] != float64(250*time.Millisecond) {
		t.Errorf("duration = %v, want 250ms", record["duration"])
	}
}

func TestAccessLogDefaultsStatusToOKWhenHandlerNeverWrites(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	handler := middleware.AccessLog(logger, clock.NewSystem())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if record := decodeLogRecord(t, logs.Bytes()); record["status"] != float64(http.StatusOK) {
		t.Fatalf("status = %v, want 200", record["status"])
	}
}

func TestAccessLogCarriesRequestIdentifier(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	handler := middleware.RequestID(id.NewRandom())(
		middleware.AccessLog(logger, clock.NewSystem())(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	record := decodeLogRecord(t, logs.Bytes())
	if record["request_id"] == "" || record["request_id"] == nil {
		t.Fatalf("request_id is missing from %v", record)
	}
	if record["request_id"] != recorder.Header().Get("X-Request-Id") {
		t.Fatalf("request_id %v does not match the response header %q", record["request_id"], recorder.Header().Get("X-Request-Id"))
	}
}

func TestAccessLogPreservesHandlerResponse(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	handler := middleware.AccessLog(logger, clock.NewSystem())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("brewing"))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))

	if recorder.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", recorder.Code)
	}
	if recorder.Body.String() != "brewing" {
		t.Fatalf("body = %q, want brewing", recorder.Body.String())
	}
}

func decodeLogRecord(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(raw), &record); err != nil {
		t.Fatalf("log output is not a single JSON record: %v (%q)", err, string(raw))
	}
	return record
}
