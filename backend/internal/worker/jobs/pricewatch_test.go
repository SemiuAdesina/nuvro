package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/worker/jobs"
)

const watchTripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"

type watchStore struct {
	due          []pricewatch.Watch
	observations []pricewatch.Observation
	rescheduled  map[string]int64
	expired      int
	err          error
}

func newWatchStore(watches ...pricewatch.Watch) *watchStore {
	return &watchStore{due: watches, rescheduled: map[string]int64{}}
}

func (w *watchStore) DuePriceWatches(context.Context, time.Time, int) ([]pricewatch.Watch, error) {
	return w.due, w.err
}

func (w *watchStore) RecordObservation(_ context.Context, _ string, observation pricewatch.Observation) error {
	w.observations = append(w.observations, observation)
	return nil
}

func (w *watchStore) ReschedulePriceWatch(_ context.Context, watchID string, _, _ time.Time, baselineMinor int64) error {
	w.rescheduled[watchID] = baselineMinor
	return nil
}

func (w *watchStore) ExpirePriceWatches(context.Context, time.Time) error {
	w.expired++
	return nil
}

func (w *watchStore) EndPriceWatch(context.Context, string, pricewatch.Status) error { return nil }

type repricer struct {
	price     int64
	available bool
	err       error
	calls     int
}

func (r *repricer) Reprice(context.Context, pricewatch.Watch) (money.Amount, bool, error) {
	r.calls++
	if r.err != nil {
		return money.Amount{}, false, r.err
	}
	amount, err := money.New(r.price, money.Currency("EUR"))
	return amount, r.available, err
}

type scheduler struct{}

func (scheduler) NextCheck(now, _ time.Time) time.Time { return now.Add(6 * time.Hour) }

func watchFixture(t *testing.T) pricewatch.Watch {
	t.Helper()
	baseline, err := money.New(40_000, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	return pricewatch.Watch{
		ID: "1c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11", UserID: "user", TripID: watchTripID,
		Product: commerce.Flight,
		Target: pricewatch.Target{
			Origin: "LIS", Destination: "FNC", StartDate: now.AddDate(0, 0, 20),
			EndDate: now.AddDate(0, 0, 27), Adults: 2,
		},
		Baseline: baseline, Threshold: pricewatch.DefaultThreshold(), Status: pricewatch.Active,
		CreatedAt: now.Add(-time.Hour), ExpiresAt: now.AddDate(0, 0, 20),
	}
}

func newWatchJob(t *testing.T, store *watchStore, prices *repricer, store2 *recorder) *jobs.PriceWatchEvaluation {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return jobs.NewPriceWatchEvaluation(store, prices, store2, scheduler{}, id.NewUUID(), logger, clock.NewFixed(now), 0)
}

func TestMeaningfulDropRaisesAnAlertAndMovesTheBaseline(t *testing.T) {
	watch := watchFixture(t)
	store := newWatchStore(watch)
	prices := &repricer{price: 32_000, available: true}
	recorded := &recorder{}
	if err := newWatchJob(t, store, prices, recorded).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(recorded.recorded) != 1 {
		t.Fatalf("alerts = %v", recorded.recorded)
	}
	produced := recorded.recorded[0]
	if produced.Type != alert.PriceDropped || produced.Action != alert.ViewOffer {
		t.Fatalf("alert = %+v", produced)
	}
	if produced.MessageData["oldAmount"] != "40000" || produced.MessageData["newAmount"] != "32000" {
		t.Fatalf("message data = %v", produced.MessageData)
	}
	if produced.MessageData["currency"] != "EUR" {
		t.Fatal("currency was lost")
	}
	if produced.Validate() != nil {
		t.Fatalf("an invalid alert was produced: %+v", produced)
	}
	if store.rescheduled[watch.ID] != 32_000 {
		t.Fatalf("baseline = %d, want the new lower price", store.rescheduled[watch.ID])
	}
	if len(store.observations) != 1 || store.observations[0].Price.Minor() != 32_000 {
		t.Fatalf("observations = %+v", store.observations)
	}
}

func TestSmallMovementRaisesNoAlert(t *testing.T) {
	watch := watchFixture(t)
	store := newWatchStore(watch)
	recorded := &recorder{}
	if err := newWatchJob(t, store, &repricer{price: 39_900, available: true}, recorded).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(recorded.recorded) != 0 {
		t.Fatalf("a tiny movement produced %v", recorded.recorded)
	}
	if store.rescheduled[watch.ID] != 40_000 {
		t.Fatal("the baseline moved without a meaningful change")
	}
}

func TestUnavailabilityRaisesItsOwnAlert(t *testing.T) {
	store := newWatchStore(watchFixture(t))
	recorded := &recorder{}
	if err := newWatchJob(t, store, &repricer{price: 40_000}, recorded).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(recorded.recorded) != 1 || recorded.recorded[0].Type != alert.AvailabilityChanged {
		t.Fatalf("alerts = %v", recorded.recorded)
	}
}

func TestRepeatedEvaluationsAtTheSamePriceDeduplicate(t *testing.T) {
	watch := watchFixture(t)
	store := newWatchStore(watch)
	recorded := &recorder{}
	job := newWatchJob(t, store, &repricer{price: 32_000, available: true}, recorded)
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	fingerprints := map[string]struct{}{}
	for _, produced := range recorded.recorded {
		fingerprints[produced.Fingerprint] = struct{}{}
	}
	if len(fingerprints) != 1 {
		t.Fatalf("the same drop produced %d distinct alerts", len(fingerprints))
	}
}

func TestProviderFailureReschedulesWithoutAlerting(t *testing.T) {
	watch := watchFixture(t)
	store := newWatchStore(watch)
	recorded := &recorder{}
	if err := newWatchJob(t, store, &repricer{err: errors.New("provider down")}, recorded).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(recorded.recorded) != 0 {
		t.Fatalf("a provider outage produced alerts: %v", recorded.recorded)
	}
	if _, rescheduled := store.rescheduled[watch.ID]; !rescheduled {
		t.Fatal("the watch was not rescheduled after a failure")
	}
	if len(store.observations) != 0 {
		t.Fatal("an observation was recorded without a price")
	}
}

func TestExpiryRunsEvenWithNoRepricer(t *testing.T) {
	store := newWatchStore(watchFixture(t))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	job := jobs.NewPriceWatchEvaluation(store, nil, &recorder{}, scheduler{}, id.NewUUID(), logger, clock.NewFixed(now), 0)
	if err := job.Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.expired != 1 {
		t.Fatal("expiry did not run")
	}
	if len(store.observations) != 0 {
		t.Fatal("prices were checked without a repricer")
	}
	if job.Name() != "price-watch-evaluation" {
		t.Fatalf("name = %q", job.Name())
	}
}

func TestWatchesWithoutATripRaiseNoAlert(t *testing.T) {
	watch := watchFixture(t)
	watch.TripID = ""
	store := newWatchStore(watch)
	recorded := &recorder{}
	if err := newWatchJob(t, store, &repricer{price: 30_000, available: true}, recorded).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(recorded.recorded) != 0 {
		t.Fatalf("an alert was raised with no trip to attach it to: %v", recorded.recorded)
	}
	if len(store.observations) != 1 {
		t.Fatal("the observation was still expected")
	}
}

func TestStoreFailuresSurface(t *testing.T) {
	store := newWatchStore()
	store.err = errors.New("database down")
	if err := newWatchJob(t, store, &repricer{price: 1, available: true}, &recorder{}).Once(t.Context()); err == nil {
		t.Fatal("a failing store reported success")
	}
}
