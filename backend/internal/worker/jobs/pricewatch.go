package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultWatchInterval = 15 * time.Minute
const watchBatchSize = 100

type WatchStore interface {
	DuePriceWatches(ctx context.Context, now time.Time, limit int) ([]pricewatch.Watch, error)
	RecordObservation(ctx context.Context, observationID string, observation pricewatch.Observation) error
	ReschedulePriceWatch(ctx context.Context, watchID string, checkedAt, nextCheckAt time.Time, baselineMinor int64) error
	ExpirePriceWatches(ctx context.Context, now time.Time) error
	EndPriceWatch(ctx context.Context, watchID string, status pricewatch.Status) error
}

type Repricer interface {
	Reprice(context.Context, pricewatch.Watch) (money.Amount, bool, error)
}

type Scheduler interface {
	NextCheck(now, departure time.Time) time.Time
}

type Identifiers interface{ Generate() (string, error) }

type PriceWatchEvaluation struct {
	store       WatchStore
	repricer    Repricer
	recorder    Recorder
	scheduler   Scheduler
	identifiers Identifiers
	logger      *slog.Logger
	timer       clock.Clock
	interval    time.Duration
}

func NewPriceWatchEvaluation(store WatchStore, repricer Repricer, recorder Recorder, scheduler Scheduler,
	identifiers Identifiers, logger *slog.Logger, timer clock.Clock, interval time.Duration) *PriceWatchEvaluation {
	if interval <= 0 {
		interval = DefaultWatchInterval
	}
	return &PriceWatchEvaluation{
		store: store, repricer: repricer, recorder: recorder, scheduler: scheduler,
		identifiers: identifiers, logger: logger, timer: timer, interval: interval,
	}
}

func (p *PriceWatchEvaluation) Name() string { return "price-watch-evaluation" }

func (p *PriceWatchEvaluation) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		if err := p.Once(ctx); err != nil && ctx.Err() == nil {
			p.logger.WarnContext(ctx, "price watch evaluation failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (p *PriceWatchEvaluation) Once(ctx context.Context) error {
	now := p.timer.Now().UTC()
	if err := p.store.ExpirePriceWatches(ctx, now); err != nil {
		return err
	}
	if p.repricer == nil {
		return nil
	}
	watches, err := p.store.DuePriceWatches(ctx, now, watchBatchSize)
	if err != nil {
		return err
	}
	for _, watch := range watches {
		if err := p.evaluate(ctx, watch, now); err != nil {
			return err
		}
	}
	return nil
}

func (p *PriceWatchEvaluation) evaluate(ctx context.Context, watch pricewatch.Watch, now time.Time) error {
	price, available, err := p.repricer.Reprice(ctx, watch)
	if err != nil {
		return p.reschedule(ctx, watch, now, watch.Baseline.Minor())
	}
	change, err := p.change(watch, price, available, now)
	if err != nil {
		return p.reschedule(ctx, watch, now, watch.Baseline.Minor())
	}
	if err := p.observe(ctx, watch, price, available, now); err != nil {
		return err
	}
	if change.Meaningful {
		if err := p.raise(ctx, watch, change); err != nil {
			return err
		}
	}
	baseline := watch.Baseline.Minor()
	if change.Direction == pricewatch.Dropped && change.Meaningful {
		baseline = price.Minor()
	}
	return p.reschedule(ctx, watch, now, baseline)
}

func (p *PriceWatchEvaluation) change(watch pricewatch.Watch, price money.Amount, available bool, now time.Time) (pricewatch.Change, error) {
	if !available {
		return pricewatch.UnavailableChange(watch.Baseline, now), nil
	}
	return pricewatch.Compare(watch.Baseline, price, watch.Threshold, now)
}

func (p *PriceWatchEvaluation) observe(ctx context.Context, watch pricewatch.Watch, price money.Amount, available bool, now time.Time) error {
	identifier, err := p.identifiers.Generate()
	if err != nil {
		return err
	}
	return p.store.RecordObservation(ctx, identifier, pricewatch.Observation{
		WatchID: watch.ID, Price: price, Available: available, ObservedAt: now,
	})
}

func (p *PriceWatchEvaluation) raise(ctx context.Context, watch pricewatch.Watch, change pricewatch.Change) error {
	if p.recorder == nil || watch.TripID == "" {
		return nil
	}
	_, err := p.recorder.Record(ctx, []alert.Alert{priceAlert(watch, change)})
	return err
}

func (p *PriceWatchEvaluation) reschedule(ctx context.Context, watch pricewatch.Watch, now time.Time, baselineMinor int64) error {
	next := now.Add(DefaultWatchInterval)
	if p.scheduler != nil {
		next = p.scheduler.NextCheck(now, watch.Target.StartDate)
	}
	return p.store.ReschedulePriceWatch(ctx, watch.ID, now, next, baselineMinor)
}
