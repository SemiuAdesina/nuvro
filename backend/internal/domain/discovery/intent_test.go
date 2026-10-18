package discovery_test

import (
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
)

func TestNormalizeTagsIsStableAndBounded(t *testing.T) {
	normalized := discovery.NormalizeTags([]string{"Beach", " food ", "beach", "", "FOOD", strings.Repeat("x", 41)})
	if len(normalized) != 2 || normalized[0] != "beach" || normalized[1] != "food" {
		t.Fatalf("normalised = %v", normalized)
	}
	if len(discovery.NormalizeTags(nil)) != 0 {
		t.Fatal("nil produced tags")
	}
}

func FuzzNormalizeTags(f *testing.F) {
	f.Add("Beach")
	f.Add("")
	f.Add(strings.Repeat("x", 60))
	f.Fuzz(func(t *testing.T, value string) {
		normalized := discovery.NormalizeTags([]string{value, value})
		if len(normalized) > 1 {
			t.Fatalf("duplicate survived: %v", normalized)
		}
		for _, tag := range normalized {
			if tag == "" || len(tag) > 40 || tag != strings.ToLower(tag) || strings.TrimSpace(tag) != tag {
				t.Fatalf("unsafe tag %q", tag)
			}
		}
	})
}
