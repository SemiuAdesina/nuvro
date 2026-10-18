package httpclient_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var probeTime = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

type counter struct {
	requests atomic.Int64
	failures atomic.Int64
}

func (c *counter) Provider(_ context.Context, _ string, failed bool, _ time.Duration) {
	c.requests.Add(1)
	if failed {
		c.failures.Add(1)
	}
}

func responder(t *testing.T, statuses ...int) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		index := int(calls.Add(1)) - 1
		if index >= len(statuses) {
			index = len(statuses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statuses[index])
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func newClient(t *testing.T, url string, metrics httpclient.Recorder, timer clock.Clock) *httpclient.Client {
	t.Helper()
	return httpclient.New(httpclient.Options{
		Provider: "probe", BaseURL: url, Timeout: 2 * time.Second,
		Retry:   httpclient.RetryPolicy{Attempts: 3, BaseDelay: time.Millisecond, MaximumGap: 2 * time.Millisecond},
		Breaker: httpclient.BreakerPolicy{FailureThreshold: 2, Cooldown: 30 * time.Second},
		Metrics: metrics, Clock: timer,
	})
}

func TestATransientFailureIsRetriedUntilItSucceeds(t *testing.T) {
	server, calls := responder(t, http.StatusServiceUnavailable, http.StatusTooManyRequests, http.StatusOK)
	metrics := &counter{}
	var into map[string]any
	if err := newClient(t, server.URL, metrics, clock.NewFixed(probeTime)).
		Get(t.Context(), httpclient.Request{Path: "/probe"}, &into); err != nil {
		t.Fatalf("a recoverable failure was not retried: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("the provider was called %d times, want three attempts", calls.Load())
	}
	if metrics.requests.Load() != 3 || metrics.failures.Load() != 2 {
		t.Fatalf("metrics recorded %d requests and %d failures",
			metrics.requests.Load(), metrics.failures.Load())
	}
}

func TestAClientErrorIsNotRetried(t *testing.T) {
	server, calls := responder(t, http.StatusBadRequest)
	err := newClient(t, server.URL, nil, clock.NewFixed(probeTime)).
		Get(t.Context(), httpclient.Request{Path: "/probe"}, nil)
	if !errors.Is(err, httpclient.ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("a client error was retried %d times", calls.Load()-1)
	}
}

func TestAPostIsRetriedOnlyWhenTheCallerSaysItIsSafe(t *testing.T) {
	server, calls := responder(t, http.StatusServiceUnavailable)
	client := newClient(t, server.URL, nil, clock.NewFixed(probeTime))
	if err := client.Post(t.Context(), httpclient.Request{Path: "/probe", Body: map[string]int{"a": 1}}, nil); err == nil {
		t.Fatal("an unavailable provider reported success")
	}
	if calls.Load() != 1 {
		t.Fatalf("an unmarked post was retried %d times", calls.Load()-1)
	}
	before := calls.Load()
	_ = client.Post(t.Context(), httpclient.Request{Path: "/probe", Body: map[string]int{"a": 1}, Retryable: true}, nil)
	if calls.Load()-before != 3 {
		t.Fatalf("a retryable post made %d attempts, want three", calls.Load()-before)
	}
}

func TestTheCircuitOpensAndStopsCallingTheProvider(t *testing.T) {
	server, calls := responder(t, http.StatusServiceUnavailable)
	timer := clock.NewFixed(probeTime)
	client := newClient(t, server.URL, nil, timer)
	for range 2 {
		_ = client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil)
	}
	attempted := calls.Load()
	if err := client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil); !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("err = %v, want an open circuit", err)
	}
	if calls.Load() != attempted {
		t.Fatal("the provider was called while the circuit was open")
	}
}

func TestTheCircuitProbesOnceAfterTheCooldown(t *testing.T) {
	server, calls := responder(t, http.StatusServiceUnavailable, http.StatusServiceUnavailable,
		http.StatusServiceUnavailable, http.StatusServiceUnavailable, http.StatusServiceUnavailable,
		http.StatusServiceUnavailable, http.StatusOK)
	moving := &movingClock{now: probeTime}
	client := newClient(t, server.URL, nil, moving)
	for range 2 {
		_ = client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil)
	}
	if err := client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil); !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("err = %v, want an open circuit", err)
	}
	moving.advance(31 * time.Second)
	before := calls.Load()
	if err := client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil); err != nil {
		t.Fatalf("the probe after cooldown failed: %v", err)
	}
	if calls.Load() <= before {
		t.Fatal("the cooldown did not let a probe through")
	}
	if err := client.Get(t.Context(), httpclient.Request{Path: "/probe"}, nil); err != nil {
		t.Fatalf("the circuit did not close after a successful probe: %v", err)
	}
}

func TestAMalformedBodyIsNeverRetried(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte("not json"))
	}))
	t.Cleanup(server.Close)
	var into map[string]any
	err := newClient(t, server.URL, nil, clock.NewFixed(probeTime)).
		Get(t.Context(), httpclient.Request{Path: "/probe"}, &into)
	if !errors.Is(err, httpclient.ErrMalformed) {
		t.Fatalf("err = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("a malformed response was retried %d times", calls.Load()-1)
	}
}

type movingClock struct{ now time.Time }

func (m *movingClock) Now() time.Time          { return m.now }
func (m *movingClock) advance(d time.Duration) { m.now = m.now.Add(d) }
