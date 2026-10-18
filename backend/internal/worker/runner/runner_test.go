package runner_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/worker/runner"
)

type stubJob struct {
	name    string
	started atomic.Bool
	failure error
	block   bool
}

func (j *stubJob) Name() string { return j.name }

func (j *stubJob) Run(ctx context.Context) error {
	j.started.Store(true)
	if j.failure != nil {
		return j.failure
	}
	if j.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
}

func TestJobNamesListsRegisteredJobs(t *testing.T) {
	worker := runner.New(discardLogger(), &stubJob{name: "price-refresh"}, &stubJob{name: "alert-sweep"})

	names := worker.JobNames()
	if len(names) != 2 || names[0] != "price-refresh" || names[1] != "alert-sweep" {
		t.Fatalf("JobNames = %v", names)
	}
}

func TestJobNamesIsEmptyWithoutJobs(t *testing.T) {
	if names := runner.New(discardLogger()).JobNames(); len(names) != 0 {
		t.Fatalf("JobNames = %v, want empty", names)
	}
}

func TestRunWithoutJobsWaitsForCancellation(t *testing.T) {
	worker := runner.New(discardLogger())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	select {
	case <-done:
		t.Fatal("Run returned before the context was cancelled")
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

func TestRunStartsEveryJobAndReturnsWhenTheyFinish(t *testing.T) {
	first := &stubJob{name: "price-refresh"}
	second := &stubJob{name: "alert-sweep"}

	if err := runner.New(discardLogger(), first, second).Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !first.started.Load() || !second.started.Load() {
		t.Fatal("not every job was started")
	}
}

func TestRunStopsAllJobsWhenTheContextIsCancelled(t *testing.T) {
	blocking := &stubJob{name: "alert-sweep", block: true}
	worker := runner.New(discardLogger(), blocking)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v, want nil for a cancelled job", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after cancellation")
	}
}

func TestRunReportsJobFailureAndStopsSiblings(t *testing.T) {
	sentinel := errors.New("price provider unreachable")
	failing := &stubJob{name: "price-refresh", failure: sentinel}
	blocking := &stubJob{name: "alert-sweep", block: true}

	err := runner.New(discardLogger(), failing, blocking).Run(context.Background())

	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want the job failure", err)
	}
	if !strings.Contains(err.Error(), "price-refresh") {
		t.Fatalf("error %q does not name the failing job", err)
	}
	if !blocking.started.Load() {
		t.Fatal("the sibling job never started")
	}
}

func TestRunLogsJobLifecycle(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	if err := runner.New(logger, &stubJob{name: "price-refresh"}).Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	recorded := logs.String()
	for _, fragment := range []string{"job started", "job stopped", "price-refresh"} {
		if !strings.Contains(recorded, fragment) {
			t.Errorf("log %q does not contain %q", recorded, fragment)
		}
	}
}

type countingJob struct {
	runs atomic.Int64
}

func (c *countingJob) Name() string { return "counting" }

func (c *countingJob) Run(ctx context.Context) error {
	c.runs.Add(1)
	<-ctx.Done()
	return nil
}

func (c *countingJob) Runs() int64 { return c.runs.Load() }

type lease struct {
	held     bool
	failure  error
	acquired int
	released int
}

func (l *lease) Acquire(context.Context) (bool, error) {
	l.acquired++
	return l.held, l.failure
}

func (l *lease) Release(context.Context) { l.released++ }

func TestAJobRunsOnlyOnTheReplicaHoldingItsLease(t *testing.T) {
	holder := &lease{held: true}
	work := &countingJob{}
	ctx, cancel := context.WithCancel(t.Context())
	group := runner.New(discardLogger(), work).WithLeases(func(string) runner.Lease { return holder })
	go func() {
		for work.Runs() == 0 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()
	if err := group.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if holder.acquired != 1 || holder.released != 1 {
		t.Fatalf("lease acquired %d released %d, want one of each", holder.acquired, holder.released)
	}
}

func TestAReplicaWithoutTheLeaseStaysIdleRatherThanDoublingTheWork(t *testing.T) {
	loser := &lease{held: false}
	work := &countingJob{}
	ctx, cancel := context.WithCancel(t.Context())
	group := runner.New(discardLogger(), work).WithLeases(func(string) runner.Lease { return loser })
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	if err := group.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if work.Runs() != 0 {
		t.Fatalf("the job ran %d times without holding its lease", work.Runs())
	}
	if loser.released != 0 {
		t.Fatal("a lease that was never held was released")
	}
}

func TestALeaseFailureStopsTheWorkerRatherThanRunningUnprotected(t *testing.T) {
	broken := &lease{failure: errors.New("database unavailable")}
	work := &countingJob{}
	group := runner.New(discardLogger(), work).WithLeases(func(string) runner.Lease { return broken })
	if err := group.Run(t.Context()); err == nil {
		t.Fatal("the worker ran without being able to check its lease")
	}
	if work.Runs() != 0 {
		t.Fatal("the job ran while its lease was unknown")
	}
}

func TestWithoutLeasesEveryJobRuns(t *testing.T) {
	work := &countingJob{}
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		for work.Runs() == 0 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()
	if err := runner.New(discardLogger(), work).Run(ctx); err != nil {
		t.Fatal(err)
	}
	if work.Runs() == 0 {
		t.Fatal("an unleased job never ran")
	}
}
