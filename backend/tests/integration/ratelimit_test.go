//go:build integration

package integration_test

import (
	"testing"
	"time"

	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

func newLimiter(t *testing.T, allowance admission.Allowance, timer clock.Clock) *redisadapter.RateLimiter {
	t.Helper()
	client, err := redisadapter.Open(redisURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	limiter, err := redisadapter.NewRateLimiter(client,
		map[admission.Class]admission.Allowance{admission.ExpensiveSearch: allowance}, timer)
	if err != nil {
		t.Fatal(err)
	}
	return limiter
}

func allow(t *testing.T, limiter *redisadapter.RateLimiter, subject string) bool {
	t.Helper()
	allowed, err := limiter.Allow(t.Context(), admission.ExpensiveSearch, subject)
	if err != nil {
		t.Fatal(err)
	}
	return allowed
}

func TestTheBucketDrainsThenRefillsOverTime(t *testing.T) {
	start := time.Now().UTC()
	timer := clock.NewFixed(start)
	limiter := newLimiter(t, admission.Allowance{Burst: 3, PerSecond: 1}, timer)
	subject := "traveller-" + randomSuffix(t)

	for attempt := range 3 {
		if !allow(t, limiter, subject) {
			t.Fatalf("attempt %d was refused inside the burst", attempt+1)
		}
	}
	if allow(t, limiter, subject) {
		t.Fatal("the bucket admitted a fourth request inside the burst")
	}

	later := newLimiter(t, admission.Allowance{Burst: 3, PerSecond: 1}, clock.NewFixed(start.Add(2*time.Second)))
	for attempt := range 2 {
		if !allow(t, later, subject) {
			t.Fatalf("refill attempt %d was refused after two seconds", attempt+1)
		}
	}
	if allow(t, later, subject) {
		t.Fatal("the bucket refilled beyond what elapsed time earns")
	}
}

func TestBudgetsAreHeldPerSubjectAndPerClass(t *testing.T) {
	timer := clock.NewFixed(time.Now().UTC())
	limiter := newLimiter(t, admission.Allowance{Burst: 1, PerSecond: 1}, timer)
	first := "traveller-" + randomSuffix(t)
	second := "traveller-" + randomSuffix(t)

	if !allow(t, limiter, first) || allow(t, limiter, first) {
		t.Fatal("the first traveller's budget did not drain exactly once")
	}
	if !allow(t, limiter, second) {
		t.Fatal("one traveller's spending drained another traveller's budget")
	}
	unbudgeted, err := limiter.Allow(t.Context(), admission.CommerceWrite, first)
	if err != nil {
		t.Fatal(err)
	}
	if !unbudgeted {
		t.Fatal("a class with no configured allowance was refused")
	}
}

func TestTheBucketExpiresSoTheKeyspaceDoesNotGrow(t *testing.T) {
	timer := clock.NewFixed(time.Now().UTC())
	limiter := newLimiter(t, admission.Allowance{Burst: 2, PerSecond: 1}, timer)
	subject := "traveller-" + randomSuffix(t)
	if !allow(t, limiter, subject) {
		t.Fatal("the first request was refused")
	}
	keys, err := inspector(t).Keys(t.Context(), "nuvro:v1:rate_limit:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("no bucket was written under the rate limit namespace")
	}
	for _, key := range keys {
		ttl, ttlErr := inspector(t).TTL(t.Context(), key).Result()
		if ttlErr != nil {
			t.Fatal(ttlErr)
		}
		if ttl <= 0 {
			t.Fatalf("%s has no expiry", key)
		}
	}
}
