//go:build integration

package integration_test

import (
	"testing"
	"time"

	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
)

func TestRecordedResponsesSurviveARoundTripThroughRedis(t *testing.T) {
	store := redisadapter.NewIdempotencyStore(newCache(t))
	subject := "traveller-" + randomSuffix(t)
	if _, found := store.Recorded(t.Context(), subject, "fresh"); found {
		t.Fatal("a cold store reported a recorded response")
	}
	recorded := replay.Response{
		Status: 201, ContentType: "application/json",
		Body: []byte(`{"id":"c0ffee","status":"draft"}`), Fingerprint: "abc123",
	}
	store.Record(t.Context(), subject, "fresh", recorded, time.Minute)

	replayed, found := store.Recorded(t.Context(), subject, "fresh")
	if !found || replayed.Status != 201 || string(replayed.Body) != string(recorded.Body) {
		t.Fatalf("replayed = %+v found = %v", replayed, found)
	}
	if replayed.Fingerprint != "abc123" || replayed.ContentType != "application/json" {
		t.Fatalf("provenance was lost: %+v", replayed)
	}
	if _, leaked := store.Recorded(t.Context(), "traveller-"+randomSuffix(t), "fresh"); leaked {
		t.Fatal("one traveller could replay another traveller's response")
	}
}

func TestRecordedResponsesExpire(t *testing.T) {
	store := redisadapter.NewIdempotencyStore(newCache(t))
	subject := "traveller-" + randomSuffix(t)
	store.Record(t.Context(), subject, "expiring", replay.Response{Status: 201, Body: []byte("{}")}, time.Minute)
	keys, err := inspector(t).Keys(t.Context(), "nuvro:v1:idempotency:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("nothing was written under the idempotency namespace")
	}
	for _, key := range keys {
		ttl, ttlErr := inspector(t).TTL(t.Context(), key).Result()
		if ttlErr != nil {
			t.Fatal(ttlErr)
		}
		if ttl <= 0 || ttl > time.Minute {
			t.Fatalf("%s has ttl %s", key, ttl)
		}
	}
}
