//go:build integration

package integration_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

func randomSuffix(t *testing.T) string {
	t.Helper()
	value, err := id.NewRandom().Generate()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
