package jobs_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/worker/jobs"
)

type purger struct {
	due     []retention.Deletion
	purged  []string
	failure error
}

func (p *purger) DueDeletions(context.Context, time.Time, int) ([]retention.Deletion, error) {
	return p.due, p.failure
}

func (p *purger) Purge(_ context.Context, deletion retention.Deletion, now time.Time) (retention.Deletion, error) {
	p.purged = append(p.purged, deletion.UserID)
	return deletion.Complete(now, 1, 2), nil
}

type sweeper struct {
	swept   []retention.Class
	cutoffs map[retention.Class]time.Time
	failure error
}

func (s *sweeper) Sweep(_ context.Context, rule retention.Rule, now time.Time) (int64, error) {
	if s.failure != nil {
		return 0, s.failure
	}
	s.swept = append(s.swept, rule.Class)
	if s.cutoffs == nil {
		s.cutoffs = map[retention.Class]time.Time{}
	}
	s.cutoffs[rule.Class] = rule.CutoffFrom(now)
	return 1, nil
}

func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)) }

func instant() time.Time { return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC) }

func sweep(purge *purger, sweeps *sweeper) *jobs.RetentionSweep {
	return jobs.NewRetentionSweep(purge, sweeps, nil, quiet(), clock.NewFixed(instant()), time.Hour)
}

type revoker struct {
	revoked int
	failure error
}

func (r *revoker) RevokePending(context.Context, int) (int, error) {
	if r.failure != nil {
		return 0, r.failure
	}
	r.revoked++
	return r.revoked, nil
}

func TestTheSweepRevokesSignInTokensBeforeErasing(t *testing.T) {
	revocations := &revoker{}
	job := jobs.NewRetentionSweep(&purger{}, &sweeper{}, revocations, quiet(), clock.NewFixed(instant()), time.Hour)
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if revocations.revoked != 1 {
		t.Fatalf("revoked %d batches, want the pending identities revoked", revocations.revoked)
	}
}

func TestAFailedRevocationDoesNotStopErasure(t *testing.T) {
	revocations := &revoker{failure: errors.New("apple unavailable")}
	sweeps := &sweeper{}
	purge := &purger{due: []retention.Deletion{{ID: "one", UserID: "traveller", Status: retention.Pending}}}
	job := jobs.NewRetentionSweep(purge, sweeps, revocations, quiet(), clock.NewFixed(instant()), time.Hour)
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(purge.purged) != 1 || len(sweeps.swept) == 0 {
		t.Fatal("a provider outage blocked erasure that does not depend on it")
	}
}

func TestTheSweepPurgesEveryAccountThatCameDue(t *testing.T) {
	purge := &purger{due: []retention.Deletion{
		{ID: "one", UserID: "traveller-one", Status: retention.Pending},
		{ID: "two", UserID: "traveller-two", Status: retention.Pending},
	}}
	if err := sweep(purge, &sweeper{}).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(purge.purged) != 2 {
		t.Fatalf("purged %v, want both due accounts", purge.purged)
	}
}

func TestTheSweepAppliesEveryRuleInThePolicy(t *testing.T) {
	sweeps := &sweeper{}
	if err := sweep(&purger{}, sweeps).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(sweeps.swept) != len(retention.Policy()) {
		t.Fatalf("swept %v, want every class in the policy", sweeps.swept)
	}
	rule, _ := retention.RuleFor(retention.PersonalizationSignals)
	if cutoff := sweeps.cutoffs[retention.PersonalizationSignals]; !cutoff.Equal(rule.CutoffFrom(instant())) {
		t.Fatalf("cutoff = %s, want the policy age before now", cutoff)
	}
}

func TestAFailureToListDueAccountsStopsTheSweep(t *testing.T) {
	purge := &purger{failure: errors.New("database unavailable")}
	sweeps := &sweeper{}
	if err := sweep(purge, sweeps).Once(t.Context()); err == nil {
		t.Fatal("a failed purge was reported as a clean sweep")
	}
	if len(sweeps.swept) != 0 {
		t.Fatal("expired data was swept while deletions were failing")
	}
}

func TestASweepFailureIsReported(t *testing.T) {
	sweeps := &sweeper{failure: errors.New("database unavailable")}
	if err := sweep(&purger{}, sweeps).Once(t.Context()); err == nil {
		t.Fatal("a failed sweep was reported as clean")
	}
}

func TestTheSweepRunsWithNothingConfigured(t *testing.T) {
	job := jobs.NewRetentionSweep(nil, nil, nil, quiet(), clock.NewFixed(instant()), 0)
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if job.Name() != "retention-sweep" {
		t.Fatalf("name = %q", job.Name())
	}
}
