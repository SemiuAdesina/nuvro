package pricewatch_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestMeaningfulDropsAreDetected(t *testing.T) {
	threshold := pricewatch.Threshold{PercentageDrop: 5, AbsoluteMinor: 1000}
	change, err := pricewatch.Compare(amount(t, 40_000, "EUR"), amount(t, 34_000, "EUR"), threshold, now)
	if err != nil {
		t.Fatal(err)
	}
	if change.Direction != pricewatch.Dropped || !change.Meaningful {
		t.Fatalf("change = %+v", change)
	}
	if change.AbsoluteMinor != 6_000 || change.PercentagePoints != 15 {
		t.Fatalf("delta = %+v", change)
	}
}

func TestSmallMovementsAreNotMeaningful(t *testing.T) {
	threshold := pricewatch.Threshold{PercentageDrop: 5, AbsoluteMinor: 1000}
	for name, current := range map[string]int64{
		"tiny percentage": 39_500,
		"tiny absolute":   39_990,
		"unchanged":       40_000,
	} {
		change, err := pricewatch.Compare(amount(t, 40_000, "EUR"), amount(t, current, "EUR"), threshold, now)
		if err != nil {
			t.Fatal(err)
		}
		if change.Meaningful {
			t.Fatalf("%s was treated as meaningful: %+v", name, change)
		}
	}
}

func TestIncreasesAreReportedSeparately(t *testing.T) {
	threshold := pricewatch.Threshold{PercentageDrop: 5, AbsoluteMinor: 1000}
	change, err := pricewatch.Compare(amount(t, 40_000, "EUR"), amount(t, 50_000, "EUR"), threshold, now)
	if err != nil {
		t.Fatal(err)
	}
	if change.Direction != pricewatch.Increased || !change.Meaningful {
		t.Fatalf("change = %+v", change)
	}
	if change.AbsoluteMinor != 10_000 {
		t.Fatalf("absolute = %d", change.AbsoluteMinor)
	}
}

func TestMixedCurrenciesAreRefused(t *testing.T) {
	if _, err := pricewatch.Compare(amount(t, 40_000, "EUR"), amount(t, 40_000, "NGN"),
		pricewatch.DefaultThreshold(), now); err == nil {
		t.Fatal("a cross-currency comparison was accepted")
	}
}

func TestUnavailabilityIsAlwaysMeaningful(t *testing.T) {
	change := pricewatch.UnavailableChange(amount(t, 40_000, "EUR"), now)
	if change.Direction != pricewatch.Unavailable || !change.Meaningful {
		t.Fatalf("change = %+v", change)
	}
}

func TestWatchValidationPerProduct(t *testing.T) {
	valid := pricewatch.Watch{
		ID: "watch", UserID: "user", Product: commerce.Flight,
		Target: pricewatch.Target{
			Origin: "LIS", Destination: "FNC", StartDate: now, EndDate: now.AddDate(0, 0, 6), Adults: 2,
		},
		Baseline: amount(t, 40_000, "EUR"), Threshold: pricewatch.DefaultThreshold(),
		Status: pricewatch.Active, CreatedAt: now, ExpiresAt: now.AddDate(0, 0, 30),
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*pricewatch.Watch){
		"no user":           func(w *pricewatch.Watch) { w.UserID = "" },
		"invented product":  func(w *pricewatch.Watch) { w.Product = "yacht" },
		"unwatched product": func(w *pricewatch.Watch) { w.Product = commerce.Insurance },
		"no route":          func(w *pricewatch.Watch) { w.Target.Origin = "" },
		"zero baseline":     func(w *pricewatch.Watch) { w.Baseline = amount(t, 0, "EUR") },
		"reversed dates":    func(w *pricewatch.Watch) { w.Target.EndDate = w.Target.StartDate.AddDate(0, 0, -2) },
		"no travellers":     func(w *pricewatch.Watch) { w.Target.Adults = 0 },
		"absurd threshold":  func(w *pricewatch.Watch) { w.Threshold.PercentageDrop = 99 },
		"expiry first":      func(w *pricewatch.Watch) { w.ExpiresAt = w.CreatedAt.Add(-time.Hour) },
	} {
		watch := valid
		mutate(&watch)
		if watch.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	stay := valid
	stay.Product = commerce.Stay
	stay.Target = pricewatch.Target{PropertyID: "hotel-1", StartDate: now, EndDate: now.AddDate(0, 0, 4), Adults: 2}
	if err := stay.Validate(); err != nil {
		t.Fatalf("a valid stay watch was refused: %v", err)
	}
}

func TestWatchLifecycle(t *testing.T) {
	watch := pricewatch.Watch{
		Status: pricewatch.Active, ExpiresAt: now.Add(time.Hour), NextCheckAt: now.Add(-time.Minute),
	}
	if !watch.Watching(now) || !watch.Due(now) {
		t.Fatal("an active due watch was not scheduled")
	}
	future := watch
	future.NextCheckAt = now.Add(time.Hour)
	if future.Due(now) {
		t.Fatal("a watch was checked before it was due")
	}
	expired := watch
	expired.ExpiresAt = now.Add(-time.Hour)
	if expired.Watching(now) || expired.Due(now) {
		t.Fatal("an expired watch is still being monitored")
	}
	ended := watch
	ended.Status = pricewatch.Ended
	if ended.Watching(now) {
		t.Fatal("an ended watch is still being monitored")
	}
}
