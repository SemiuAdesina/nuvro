package money_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func TestNewRateRejectsNonPositiveValues(t *testing.T) {
	for _, micros := range []int64{0, -1, math.MinInt64} {
		if _, err := money.NewRate(micros); !errors.Is(err, money.ErrNonPositiveRate) {
			t.Errorf("NewRate(%d) error = %v, want ErrNonPositiveRate", micros, err)
		}
	}
}

func TestRateApplyRoundsHalfAwayFromZero(t *testing.T) {
	rate := mustRate(t, 1_500_000)

	converted, err := rate.Apply(mustAmount(t, 3, "USD"))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if converted != 5 {
		t.Fatalf("Apply = %d, want 5", converted)
	}

	convertedNegative, err := rate.Apply(mustAmount(t, -3, "USD"))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if convertedNegative != -5 {
		t.Fatalf("Apply = %d, want -5", convertedNegative)
	}
}

func TestRateApplyDetectsOverflow(t *testing.T) {
	rate := mustRate(t, 2_000_000)

	if _, err := rate.Apply(mustAmount(t, math.MaxInt64, "USD")); !errors.Is(err, money.ErrOverflow) {
		t.Fatalf("error = %v, want ErrOverflow", err)
	}
}

func TestNewConversionRetainsProvenance(t *testing.T) {
	observedAt := time.Date(2026, time.September, 7, 10, 30, 0, 0, time.FixedZone("WAT", 3600))

	conversion, err := money.NewConversion(
		mustAmount(t, 10000, "USD"),
		mustAmount(t, 9200, "EUR"),
		mustRate(t, 920_000),
		observedAt,
		"ecb",
	)
	if err != nil {
		t.Fatalf("NewConversion returned error: %v", err)
	}
	if conversion.Source.Currency() != "USD" || conversion.Target.Currency() != "EUR" {
		t.Fatalf("currencies = %s -> %s, want USD -> EUR", conversion.Source.Currency(), conversion.Target.Currency())
	}
	if conversion.Provider != "ecb" {
		t.Fatalf("Provider = %q, want %q", conversion.Provider, "ecb")
	}
	if conversion.ObservedAt.Location() != time.UTC {
		t.Fatalf("ObservedAt location = %v, want UTC", conversion.ObservedAt.Location())
	}
	if !conversion.ObservedAt.Equal(observedAt) {
		t.Fatalf("ObservedAt = %v, want the same instant as %v", conversion.ObservedAt, observedAt)
	}
}

func TestNewConversionRejectsIncompleteProvenance(t *testing.T) {
	source := mustAmount(t, 100, "USD")
	target := mustAmount(t, 90, "EUR")
	rate := mustRate(t, 900_000)
	observedAt := time.Date(2026, time.September, 7, 0, 0, 0, 0, time.UTC)

	if _, err := money.NewConversion(source, source, rate, observedAt, "ecb"); !errors.Is(err, money.ErrSameCurrency) {
		t.Errorf("same currency error = %v, want ErrSameCurrency", err)
	}
	if _, err := money.NewConversion(source, target, rate, time.Time{}, "ecb"); !errors.Is(err, money.ErrMissingRateTime) {
		t.Errorf("missing timestamp error = %v, want ErrMissingRateTime", err)
	}
	if _, err := money.NewConversion(source, target, rate, observedAt, ""); !errors.Is(err, money.ErrMissingRateOwner) {
		t.Errorf("missing provider error = %v, want ErrMissingRateOwner", err)
	}
}

func TestConversionAgeMeasuresStaleness(t *testing.T) {
	observedAt := time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)
	conversion, err := money.NewConversion(
		mustAmount(t, 100, "USD"),
		mustAmount(t, 90, "EUR"),
		mustRate(t, 900_000),
		observedAt,
		"ecb",
	)
	if err != nil {
		t.Fatalf("NewConversion returned error: %v", err)
	}

	if got := conversion.Age(observedAt.Add(90 * time.Minute)); got != 90*time.Minute {
		t.Fatalf("Age = %v, want 90m", got)
	}
}

func mustRate(t *testing.T, micros int64) money.Rate {
	t.Helper()
	rate, err := money.NewRate(micros)
	if err != nil {
		t.Fatalf("NewRate(%d) returned error: %v", micros, err)
	}
	return rate
}
