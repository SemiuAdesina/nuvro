package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

type Job interface {
	Name() string
	Run(ctx context.Context) error
}

type Lease interface {
	Acquire(context.Context) (bool, error)
	Release(context.Context)
}

type Leases func(name string) Lease

type Runner struct {
	logger *slog.Logger
	leases Leases
	jobs   []Job
}

func New(logger *slog.Logger, jobs ...Job) *Runner {
	return &Runner{logger: logger, jobs: jobs}
}

func (r *Runner) WithLeases(leases Leases) *Runner {
	r.leases = leases
	return r
}

func (r *Runner) JobNames() []string {
	names := make([]string, 0, len(r.jobs))
	for _, job := range r.jobs {
		names = append(names, job.Name())
	}
	return names
}

func (r *Runner) Run(ctx context.Context) error {
	if len(r.jobs) == 0 {
		r.logger.InfoContext(ctx, "worker idle, no jobs registered")
		<-ctx.Done()
		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	for _, job := range r.jobs {
		group.Go(func() error { return r.run(groupCtx, job) })
	}
	return group.Wait()
}

func (r *Runner) run(ctx context.Context, job Job) error {
	release, held, err := r.claim(ctx, job)
	if err != nil {
		return fmt.Errorf("job %s: %w", job.Name(), err)
	}
	if !held {
		r.logger.InfoContext(ctx, "job runs on another replica", slog.String("job", job.Name()))
		<-ctx.Done()
		return nil
	}
	defer release()
	r.logger.InfoContext(ctx, "job started", slog.String("job", job.Name()))
	if err := job.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("job %s: %w", job.Name(), err)
	}
	r.logger.InfoContext(ctx, "job stopped", slog.String("job", job.Name()))
	return nil
}

func (r *Runner) claim(ctx context.Context, job Job) (func(), bool, error) {
	if r.leases == nil {
		return func() {}, true, nil
	}
	lease := r.leases(job.Name())
	if lease == nil {
		return func() {}, true, nil
	}
	held, err := lease.Acquire(ctx)
	if err != nil {
		return func() {}, false, err
	}
	if !held {
		return func() {}, false, nil
	}
	return func() { lease.Release(context.WithoutCancel(ctx)) }, true, nil
}
