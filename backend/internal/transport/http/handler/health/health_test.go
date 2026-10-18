package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/health"
)

type probe struct {
	calls   int
	failure error
}

func (*probe) Name() string                  { return "postgres" }
func (p *probe) Check(context.Context) error { p.calls++; return p.failure }

func TestStrictHealthMapping(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	dependency := &probe{}
	handler := health.NewHandler(clock.NewFixed(now), dependency)
	live, err := handler.GetLiveness(t.Context(), api.GetLivenessRequestObject{})
	liveResponse, ok := live.(api.GetLiveness200JSONResponse)
	if err != nil || !ok || liveResponse.CheckedAt != now || dependency.calls != 0 {
		t.Fatalf("live = %v, %v", live, err)
	}
	ready, err := handler.GetReadiness(t.Context(), api.GetReadinessRequestObject{})
	if err != nil {
		t.Fatal(err)
	}
	response, ok := ready.(api.GetReadiness200JSONResponse)
	if !ok {
		t.Fatal("unexpected response type")
	}
	if len(response.Components) != 1 || response.Components[0].Name != "postgres" || response.Components[0].Status != "up" {
		t.Fatalf("ready = %+v", response)
	}
	dependency.failure = errors.New("private connection detail")
	if _, err := handler.GetReadiness(t.Context(), api.GetReadinessRequestObject{}); err == nil {
		t.Fatal("unavailable dependency returned success")
	}
}
