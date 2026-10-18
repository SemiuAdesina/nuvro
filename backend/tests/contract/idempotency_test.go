package contract_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
)

type replays struct {
	recorded map[string]replay.Response
	writes   int
}

func newReplays() *replays { return &replays{recorded: map[string]replay.Response{}} }

func (r *replays) Recorded(_ context.Context, subject, key string) (replay.Response, bool) {
	found, ok := r.recorded[subject+"|"+key]
	return found, ok
}

func (r *replays) Record(_ context.Context, subject, key string, recorded replay.Response, _ time.Duration) {
	r.writes++
	r.recorded[subject+"|"+key] = recorded
}

func createTripWithKey(t *testing.T, environment *environment, key string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	body := `{"destinationId":"madeira-pt","originAirport":"LIS","adults":2,` +
		`"preferredCurrency":"EUR","startDate":"2026-09-12","endDate":"2026-09-18"}`
	return environment.callWithKey(t, http.MethodPost, "/v1/trips", body, true, key)
}

func TestARetriedCreationReturnsTheOriginalTrip(t *testing.T) {
	store := newReplays()
	environment := newEnvironment(t, withReplays(store))
	first, created := createTripWithKey(t, environment, "retry-once")
	if first.Code != http.StatusCreated {
		t.Fatalf("returned %d %v", first.Code, created)
	}
	second, replayed := createTripWithKey(t, environment, "retry-once")
	if second.Code != http.StatusCreated {
		t.Fatalf("the retry returned %d %v", second.Code, replayed)
	}
	if text(t, replayed, "id") != text(t, created, "id") {
		t.Fatalf("the retry created a second trip: %v then %v", created["id"], replayed["id"])
	}
	if second.Header().Get("Idempotent-Replay") != "true" {
		t.Fatal("the retry was served fresh rather than replayed")
	}
	if len(environment.store.trips) != 1 {
		t.Fatalf("%d trips exist after one creation and one retry", len(environment.store.trips))
	}
	if store.writes != 1 {
		t.Fatalf("the response was recorded %d times", store.writes)
	}
}

func TestReusingAKeyForADifferentRequestIsRefused(t *testing.T) {
	environment := newEnvironment(t, withReplays(newReplays()))
	if first, decoded := createTripWithKey(t, environment, "shared"); first.Code != http.StatusCreated {
		t.Fatalf("returned %d %v", first.Code, decoded)
	}
	different := `{"destinationId":"madeira-pt","originAirport":"OPO","adults":4,` +
		`"preferredCurrency":"EUR","startDate":"2026-09-12","endDate":"2026-09-18"}`
	response, problem := environment.callWithKey(t, http.MethodPost, "/v1/trips", different, true, "shared")
	if response.Code != http.StatusConflict || problem["code"] != "conflict" {
		t.Fatalf("returned %d %v, want a refused key reuse", response.Code, problem)
	}
	if len(environment.store.trips) != 1 {
		t.Fatalf("%d trips exist after a refused reuse", len(environment.store.trips))
	}
}

func TestDistinctKeysCreateDistinctTrips(t *testing.T) {
	environment := newEnvironment(t, withReplays(newReplays()))
	_, first := createTripWithKey(t, environment, "one")
	_, second := createTripWithKey(t, environment, "two")
	if text(t, first, "id") == text(t, second, "id") {
		t.Fatal("two distinct keys returned the same trip")
	}
	if len(environment.store.trips) != 2 {
		t.Fatalf("%d trips exist after two distinct creations", len(environment.store.trips))
	}
}

func TestCreationWithoutAKeyIsRefusedByTheContract(t *testing.T) {
	environment := newEnvironment(t, withReplays(newReplays()))
	body := `{"destinationId":"madeira-pt","originAirport":"LIS","adults":2,` +
		`"preferredCurrency":"EUR","startDate":"2026-09-12","endDate":"2026-09-18"}`
	response, problem := environment.callWithKey(t, http.MethodPost, "/v1/trips", body, true, "")
	if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestReadsAreNeverRecordedForReplay(t *testing.T) {
	store := newReplays()
	environment := newEnvironment(t, withReplays(store))
	if _, decoded := createTripWithKey(t, environment, "read-check"); decoded["id"] == nil {
		t.Fatal("the trip was not created")
	}
	before := store.writes
	if response, _ := environment.call(t, http.MethodGet, "/v1/me", "", true); response.Code != http.StatusOK {
		t.Fatalf("the profile read returned %d", response.Code)
	}
	if store.writes != before {
		t.Fatal("a read was recorded for replay")
	}
}
