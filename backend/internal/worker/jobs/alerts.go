package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultInterval = 5 * time.Minute
const evaluationHorizon = 12 * time.Hour

type TripSource interface {
	TripsNeedingEvaluation(ctx context.Context, from, until time.Time, limit int) ([]trip.Trip, error)
}

type Snapshots interface {
	Evaluate(ctx context.Context, tripID string) (alerting.Input, bool)
}

type Recorder interface {
	Record(context.Context, []alert.Alert) ([]alert.Alert, error)
}

type AlertEvaluation struct {
	trips     TripSource
	snapshots Snapshots
	engine    *alerting.Engine
	recorder  Recorder
	logger    *slog.Logger
	timer     clock.Clock
	interval  time.Duration
	limit     int
}

func NewAlertEvaluation(trips TripSource, snapshots Snapshots, engine *alerting.Engine, recorder Recorder,
	logger *slog.Logger, timer clock.Clock, interval time.Duration) *AlertEvaluation {
	if interval <= 0 {
		interval = DefaultInterval
	}
	return &AlertEvaluation{
		trips: trips, snapshots: snapshots, engine: engine, recorder: recorder,
		logger: logger, timer: timer, interval: interval, limit: 100,
	}
}

func (a *AlertEvaluation) Name() string { return "alert-evaluation" }

func (a *AlertEvaluation) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		if err := a.Once(ctx); err != nil && ctx.Err() == nil {
			a.logger.WarnContext(ctx, "alert evaluation failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (a *AlertEvaluation) Once(ctx context.Context) error {
	now := a.timer.Now().UTC()
	candidates, err := a.trips.TripsNeedingEvaluation(ctx, now, now.Add(evaluationHorizon), a.limit)
	if err != nil {
		return err
	}
	for _, record := range candidates {
		if err := a.evaluate(ctx, record.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *AlertEvaluation) evaluate(ctx context.Context, tripID string) error {
	input, ready := a.snapshots.Evaluate(ctx, tripID)
	if !ready {
		return nil
	}
	produced := a.engine.Evaluate(input)
	if len(produced) == 0 {
		return nil
	}
	_, err := a.recorder.Record(ctx, produced)
	return err
}
