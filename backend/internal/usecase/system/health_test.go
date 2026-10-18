package system_test

import (
	"context"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/system"
)

type blockingProbe struct{}

func (blockingProbe) Name() string                    { return "dependency" }
func (blockingProbe) Check(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }

func TestReadinessDeadlineAndCancellation(t *testing.T) {
	health := system.NewHealth(clock.NewSystem(), 10*time.Millisecond, blockingProbe{})
	started := time.Now()
	if _, err := health.Ready(t.Context()); err == nil {
		t.Fatal("expected unavailable")
	}
	if time.Since(started) > time.Second {
		t.Fatal("readiness deadline was not bounded")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := health.Ready(ctx); err == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestNoDependenciesReturnsEmptyCollection(t *testing.T) {
	health := system.NewHealth(clock.NewSystem(), time.Second)
	result, err := health.Ready(t.Context())
	if err != nil || result.Components == nil || len(result.Components) != 0 {
		t.Fatalf("%+v %v", result, err)
	}
}
