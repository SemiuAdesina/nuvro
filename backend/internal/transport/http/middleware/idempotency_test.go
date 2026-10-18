package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

type replayStore struct {
	recorded map[string]replay.Response
	writes   int
}

func newReplayStore() *replayStore { return &replayStore{recorded: map[string]replay.Response{}} }

func (s *replayStore) Recorded(_ context.Context, subject, key string) (replay.Response, bool) {
	found, ok := s.recorded[subject+"|"+key]
	return found, ok
}

func (s *replayStore) Record(_ context.Context, subject, key string, recorded replay.Response, _ time.Duration) {
	s.writes++
	s.recorded[subject+"|"+key] = recorded
}

func traveller(*http.Request) (string, bool) { return "traveller-1", true }

func creation(t *testing.T, store replay.Store, body string) http.Handler {
	t.Helper()
	created := 0
	return middleware.Idempotent(store, traveller)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		created++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(body))
	}))
}

func post(t *testing.T, handler http.Handler, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/trips", strings.NewReader(body))
	if key != "" {
		request.Header.Set(middleware.IdempotencyKeyHeader, key)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestOnlyWellFormedKeysAreAccepted(t *testing.T) {
	handler := creation(t, newReplayStore(), `{"id":"trip-1"}`)
	for _, test := range []struct {
		header string
		status int
	}{
		{"order-2026.09:A_b", http.StatusCreated},
		{strings.Repeat("k", 255), http.StatusCreated},
		{"", http.StatusBadRequest},
		{strings.Repeat("k", 256), http.StatusBadRequest},
		{"has space", http.StatusBadRequest},
		{"emoji-✈", http.StatusBadRequest},
		{"semi;colon", http.StatusBadRequest},
	} {
		response := post(t, handler, test.header, `{"trip":1}`)
		if response.Code != test.status {
			t.Fatalf("header %q returned %d, want %d", test.header, response.Code, test.status)
		}
		if test.status == http.StatusBadRequest && response.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatalf("header %q was rejected without a problem document", test.header)
		}
	}
}

func TestTheSameKeyAndBodyReplaysTheFirstResponse(t *testing.T) {
	store := newReplayStore()
	handler := creation(t, store, `{"id":"trip-1"}`)
	first := post(t, handler, "same", `{"trip":1}`)
	second := post(t, handler, "same", `{"trip":1}`)
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("statuses %d then %d", first.Code, second.Code)
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("replay returned %q, want %q", second.Body.String(), first.Body.String())
	}
	if second.Header().Get("Idempotent-Replay") != "true" || store.writes != 1 {
		t.Fatalf("the retry was served fresh: replay header %q, %d writes",
			second.Header().Get("Idempotent-Replay"), store.writes)
	}
}

func TestTheSameKeyWithADifferentBodyConflicts(t *testing.T) {
	handler := creation(t, newReplayStore(), `{"id":"trip-1"}`)
	if response := post(t, handler, "same", `{"trip":1}`); response.Code != http.StatusCreated {
		t.Fatalf("first request returned %d", response.Code)
	}
	response := post(t, handler, "same", `{"trip":2}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("returned %d, want a refused key reuse", response.Code)
	}
}

func TestTheKeyReachesTheHandlerAndAServerErrorIsNotRecorded(t *testing.T) {
	store := newReplayStore()
	var received string
	handler := middleware.Idempotent(store, traveller)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = middleware.IdempotencyKeyFrom(r.Context())
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if response := post(t, handler, "boom", `{}`); response.Code != http.StatusInternalServerError {
		t.Fatalf("returned %d", response.Code)
	}
	if received != "boom" {
		t.Fatalf("the handler saw key %q", received)
	}
	if store.writes != 0 {
		t.Fatal("a server error was recorded for replay")
	}
	if _, ok := middleware.IdempotencyKeyFrom(t.Context()); ok {
		t.Fatal("key reported on a context without one")
	}
}

func TestRoutesThatDoNotCreateAreLeftAlone(t *testing.T) {
	store := newReplayStore()
	handler := middleware.Idempotent(store, traveller)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || store.writes != 0 {
		t.Fatalf("a read was intercepted: %d, %d writes", response.Code, store.writes)
	}
	if middleware.RequiresIdempotency(http.MethodGet, "/v1/me") ||
		!middleware.RequiresIdempotency(http.MethodPost, "/v1/commerce/handoffs") {
		t.Fatal("the replayed route list is wrong")
	}
}

func FuzzParseIdempotencyKey(f *testing.F) {
	f.Add("order-1")
	f.Add("")
	f.Add("with space")
	f.Add(strings.Repeat("a", 300))
	f.Fuzz(func(t *testing.T, raw string) {
		key, ok := middleware.ParseIdempotencyKey(raw)
		if !ok {
			if key != "" {
				t.Fatal("rejected key leaked a value")
			}
			return
		}
		if key != raw || len(key) == 0 || len(key) > 255 {
			t.Fatalf("accepted %q as %q", raw, key)
		}
		for _, r := range key {
			if r > 0x7f || strings.ContainsRune(" \t\r\n;,/\\", r) {
				t.Fatalf("accepted unsafe rune %q", r)
			}
		}
	})
}
