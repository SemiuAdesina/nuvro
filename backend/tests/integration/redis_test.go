//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
)

func TestRealRedisReadinessAndCancellation(t *testing.T) {
	url := os.Getenv("NUVRO_TEST_REDIS_URL")
	if url == "" {
		t.Fatal("NUVRO_TEST_REDIS_URL is required")
	}
	client, err := redis.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Check(t.Context()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := client.Check(ctx); err == nil {
		t.Fatal("Redis ignored cancellation")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Check(t.Context()); err == nil {
		t.Fatal("closed Redis is ready")
	}
}
