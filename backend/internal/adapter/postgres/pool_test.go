package postgres_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
)

func TestPoolRejectsMalformedConfiguration(t *testing.T) {
	if _, err := postgres.Open(t.Context(), "postgres://%zz"); err == nil {
		t.Fatal("malformed configuration accepted")
	}
}
