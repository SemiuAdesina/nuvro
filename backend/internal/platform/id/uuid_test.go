package id_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

var canonicalVersionFour = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUIDGeneratesCanonicalVersionFourIdentifiers(t *testing.T) {
	generator := id.NewUUID()
	seen := make(map[string]struct{}, 256)
	for range 256 {
		value, err := generator.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if !canonicalVersionFour.MatchString(value) {
			t.Fatalf("Generate() = %q, want canonical RFC 4122 version 4", value)
		}
		if _, duplicate := seen[value]; duplicate {
			t.Fatalf("Generate() repeated %q", value)
		}
		seen[value] = struct{}{}
	}
}

func TestUUIDSetsVersionAndVariantOverAnySource(t *testing.T) {
	generator := id.NewUUIDFrom(func(raw []byte) (int, error) {
		for index := range raw {
			raw[index] = 0xff
		}
		return len(raw), nil
	})
	value, err := generator.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if value != "ffffffff-ffff-4fff-bfff-ffffffffffff" {
		t.Fatalf("Generate() = %q", value)
	}
}

func TestUUIDPropagatesSourceFailure(t *testing.T) {
	sentinel := errors.New("entropy unavailable")
	value, err := id.NewUUIDFrom(func([]byte) (int, error) { return 0, sentinel }).Generate()
	if !errors.Is(err, sentinel) || value != "" {
		t.Fatalf("Generate() = %q, %v", value, err)
	}
	var _ id.Generator = id.NewUUID()
}
