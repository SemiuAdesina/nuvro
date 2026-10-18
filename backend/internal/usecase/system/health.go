package system

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Probe interface {
	Name() string
	Check(context.Context) error
}

type Health struct {
	timer   clock.Clock
	probes  []Probe
	timeout time.Duration
}
type Ready struct {
	CheckedAt  time.Time
	Components []string
}

func NewHealth(timer clock.Clock, timeout time.Duration, probes ...Probe) *Health {
	return &Health{timer: timer, timeout: timeout, probes: probes}
}

func (h *Health) Live() time.Time { return h.timer.Now().UTC() }

func (h *Health) Ready(ctx context.Context) (Ready, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	result := Ready{CheckedAt: h.timer.Now().UTC(), Components: make([]string, 0, len(h.probes))}
	for _, probe := range h.probes {
		if err := probe.Check(ctx); err != nil {
			return Ready{}, apperror.New(apperror.CodeUnavailable, "A required dependency is unavailable.")
		}
		result.Components = append(result.Components, probe.Name())
	}
	return result, nil
}
