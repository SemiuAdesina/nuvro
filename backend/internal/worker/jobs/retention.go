package jobs

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

const DefaultRetentionInterval = 6 * time.Hour
const purgeBatchSize = 25

type Revoker interface {
	RevokePending(context.Context, int) (int, error)
}

type RetentionSweep struct {
	purger   account.Purger
	sweeper  account.Sweeper
	revoker  Revoker
	logger   *slog.Logger
	timer    clock.Clock
	interval time.Duration
}

func NewRetentionSweep(purger account.Purger, sweeper account.Sweeper, revoker Revoker,
	logger *slog.Logger, timer clock.Clock, interval time.Duration) *RetentionSweep {
	if interval <= 0 {
		interval = DefaultRetentionInterval
	}
	if revoker != nil && reflect.ValueOf(revoker).IsNil() {
		revoker = nil
	}
	return &RetentionSweep{
		purger: purger, sweeper: sweeper, revoker: revoker,
		logger: logger, timer: timer, interval: interval,
	}
}

func (r *RetentionSweep) Name() string { return "retention-sweep" }

func (r *RetentionSweep) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		if err := r.Once(ctx); err != nil && ctx.Err() == nil {
			r.logger.WarnContext(ctx, "retention sweep failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (r *RetentionSweep) Once(ctx context.Context) error {
	now := r.timer.Now().UTC()
	r.revokeIdentities(ctx)
	if err := r.purge(ctx, now); err != nil {
		return err
	}
	return r.sweep(ctx, now)
}

func (r *RetentionSweep) revokeIdentities(ctx context.Context) {
	if r.revoker == nil {
		return
	}
	revoked, err := r.revoker.RevokePending(ctx, purgeBatchSize)
	if err != nil {
		r.logger.WarnContext(ctx, "identity revocation failed", slog.String("error", err.Error()))
		return
	}
	if revoked > 0 {
		r.logger.InfoContext(ctx, "sign in tokens revoked", slog.Int("identities", revoked))
	}
}

func (r *RetentionSweep) purge(ctx context.Context, now time.Time) error {
	if r.purger == nil {
		return nil
	}
	due, err := r.purger.DueDeletions(ctx, now, purgeBatchSize)
	if err != nil {
		return err
	}
	for _, deletion := range due {
		completed, purgeErr := r.purger.Purge(ctx, deletion, now)
		if purgeErr != nil {
			return purgeErr
		}
		r.logger.InfoContext(ctx, "account purged",
			slog.String("deletion_id", completed.ID),
			slog.Int("trips_transferred", completed.TripsTransferred),
			slog.Int("trips_deleted", completed.TripsDeleted))
	}
	return nil
}

func (r *RetentionSweep) sweep(ctx context.Context, now time.Time) error {
	if r.sweeper == nil {
		return nil
	}
	for _, rule := range retention.Policy() {
		removed, err := r.sweeper.Sweep(ctx, rule, now)
		if err != nil {
			return err
		}
		if removed > 0 {
			r.logger.InfoContext(ctx, "retention sweep removed expired records",
				slog.String("class", string(rule.Class)), slog.Int64("removed", removed))
		}
	}
	return nil
}
