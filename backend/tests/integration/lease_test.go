//go:build integration

package integration_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
)

func TestOnlyOneReplicaHoldsAJobLease(t *testing.T) {
	pool, _ := openDatabase(t)
	name := "nuvro:job:retention-sweep:" + randomSuffix(t)

	first := postgres.NewLease(pool.Pool, name)
	held, err := first.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !held {
		t.Fatal("the first replica could not take a free lease")
	}

	second := postgres.NewLease(pool.Pool, name)
	alsoHeld, err := second.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if alsoHeld {
		t.Fatal("a second replica took a lease that was already held, so the job would run twice")
	}

	first.Release(t.Context())
	regained, err := second.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !regained {
		t.Fatal("the lease was never freed, so the job would stop running after a restart")
	}
	second.Release(t.Context())
}

func TestDifferentJobsDoNotBlockEachOther(t *testing.T) {
	pool, _ := openDatabase(t)
	suffix := randomSuffix(t)
	alerts := postgres.NewLease(pool.Pool, "nuvro:job:alert-evaluation:"+suffix)
	retention := postgres.NewLease(pool.Pool, "nuvro:job:retention-sweep:"+suffix)

	for name, lease := range map[string]*postgres.Lease{"alerts": alerts, "retention": retention} {
		held, err := lease.Acquire(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if !held {
			t.Fatalf("the %s job could not take its own lease", name)
		}
	}
	alerts.Release(t.Context())
	retention.Release(t.Context())
}

func TestAnAbsentPoolLeavesTheJobRunning(t *testing.T) {
	var absent *postgres.Lease
	held, err := absent.Acquire(t.Context())
	if err != nil || !held {
		t.Fatalf("held = %v %v, want a single-process deployment to keep running", held, err)
	}
	absent.Release(t.Context())
}
