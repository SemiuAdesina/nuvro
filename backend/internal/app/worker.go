package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/worker/runner"
)

const workerShutdownTimeout = 10 * time.Second

type Worker struct {
	logger       *slog.Logger
	runner       *runner.Runner
	dependencies *dependencies
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.InfoContext(ctx, "worker starting", slog.Any("jobs", w.runner.JobNames()))

	runErr := w.runner.Run(ctx)
	if w.dependencies != nil {
		shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), workerShutdownTimeout)
		defer cancel()
		_ = w.dependencies.close(shutdown)
	}
	if runErr != nil {
		return fmt.Errorf("app: worker: %w", runErr)
	}

	w.logger.Info("worker stopped")
	return nil
}
