package health

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/system"
)

type Probe = system.Probe
type Handler struct{ health *system.Health }

func NewHandler(timer clock.Clock, probes ...Probe) *Handler {
	return &Handler{health: system.NewHealth(timer, 2*time.Second, probes...)}
}

func (h *Handler) GetLiveness(_ context.Context, _ api.GetLivenessRequestObject) (api.GetLivenessResponseObject, error) {
	return api.GetLiveness200JSONResponse{Status: "live", CheckedAt: h.health.Live()}, nil
}

func (h *Handler) GetReadiness(ctx context.Context, _ api.GetReadinessRequestObject) (api.GetReadinessResponseObject, error) {
	result, err := h.health.Ready(ctx)
	if err != nil {
		return nil, err
	}
	response := api.GetReadiness200JSONResponse{Status: "ready", CheckedAt: result.CheckedAt}
	response.Components = make([]struct {
		Name   string                        `json:"name"`
		Status api.ReadinessComponentsStatus `json:"status"`
	}, 0, len(result.Components))
	for _, name := range result.Components {
		response.Components = append(response.Components, struct {
			Name   string                        `json:"name"`
			Status api.ReadinessComponentsStatus `json:"status"`
		}{Name: name, Status: "up"})
	}
	return response, nil
}
