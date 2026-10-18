package config_test

import (
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

func TestConnectionValidationDoesNotExposeCredentials(t *testing.T) {
	for _, key := range []string{"NUVRO_DATABASE_URL", "NUVRO_REDIS_URL"} {
		_, err := config.Load(lookupFrom(map[string]string{key: "https://synthetic:private@host"}))
		if err == nil || strings.Contains(err.Error(), "private") {
			t.Fatalf("unsafe validation: %v", err)
		}
	}
}
