package redis_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
)

func TestKeyBuildsNamespacedKeysOnly(t *testing.T) {
	key, err := redis.Key("weather", "lisbon-2026-09-10")
	if err != nil || key != "nuvro:v1:weather:lisbon-2026-09-10" {
		t.Fatalf("%q %v", key, err)
	}
	for _, test := range []struct{ capability, identifier string }{
		{"", "id"}, {"Weather", "id"}, {"rate-limit", "id"}, {"weather", ""},
		{"weather", "has space"}, {"weather", "colon:inside"}, {"weather", "glob*"},
		{"weather", "tab\there"}, {"weather", strings.Repeat("x", 520)},
	} {
		if _, err := redis.Key(test.capability, test.identifier); !errors.Is(err, redis.ErrInvalidKey) {
			t.Fatalf("accepted %q %q", test.capability, test.identifier)
		}
	}
}

func FuzzKey(f *testing.F) {
	f.Add("weather", "lisbon")
	f.Add("rate_limit", "principal:1")
	f.Fuzz(func(t *testing.T, capability, identifier string) {
		key, err := redis.Key(capability, identifier)
		if err != nil {
			return
		}
		if !strings.HasPrefix(key, "nuvro:v1:") || strings.Count(key, ":") != 3 || len(key) > 512 {
			t.Fatalf("built unsafe key %q", key)
		}
	})
}
