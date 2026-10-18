package redis_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
)

func TestClientConfiguration(t *testing.T) {
	if _, err := redis.Open("https://invalid"); err == nil {
		t.Fatal("invalid protocol accepted")
	}
	client, err := redis.Open("redis://localhost:1/0")
	if err != nil {
		t.Fatal(err)
	}
	if client.Name() != "redis" {
		t.Fatal("wrong dependency name")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}
