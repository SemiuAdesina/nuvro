package id_test

import (
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

func TestRandomGeneratesDistinctHexIdentifiers(t *testing.T) {
	generator := id.NewRandom()

	seen := make(map[string]struct{}, 256)
	for range 256 {
		value, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if len(value) != 32 {
			t.Fatalf("Generate() = %q, want 32 hex characters", value)
		}
		for _, r := range value {
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
				t.Fatalf("Generate() = %q, which is not lowercase hex", value)
			}
		}
		if _, duplicate := seen[value]; duplicate {
			t.Fatalf("Generate() repeated %q", value)
		}
		seen[value] = struct{}{}
	}
}

func TestRandomPropagatesSourceFailure(t *testing.T) {
	sentinel := errors.New("entropy unavailable")
	generator := id.NewRandomFrom(func([]byte) (int, error) { return 0, sentinel })

	value, err := generator.Generate()
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want the source failure", err)
	}
	if value != "" {
		t.Fatalf("Generate() = %q, want an empty identifier on failure", value)
	}
}

func TestRandomSatisfiesGenerator(t *testing.T) {
	var _ id.Generator = id.NewRandom()
}
