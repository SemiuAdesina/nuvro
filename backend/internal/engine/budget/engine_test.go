package budget_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var calculatedAt = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

type estimates struct {
	values []engine.Estimate
	err    error
}

func (e estimates) Estimates(context.Context, engine.Query) ([]engine.Estimate, error) {
	return e.values, e.err
}

type booked struct {
	values []engine.Booked
	err    error
}

func (b booked) Booked(context.Context, string) ([]engine.Booked, error) { return b.values, b.err }

type rates struct {
	factor int64
	err    error
	calls  int
}

func (r *rates) Convert(_ context.Context, amount money.Amount, target money.Currency, _ time.Time) (money.Amount, error) {
	r.calls++
	if r.err != nil {
		return money.Amount{}, r.err
	}
	return money.New(amount.Minor()*r.factor, target)
}

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func estimate(t *testing.T, category domain.Category, minimum, expected, maximum int64, currency string) engine.Estimate {
	t.Helper()
	return engine.Estimate{
		Category: category, Minimum: amount(t, minimum, currency), Expected: amount(t, expected, currency),
		Maximum: amount(t, maximum, currency), Confidence: domain.Medium, Source: "fixture", ObservedAt: calculatedAt,
	}
}

func input(t *testing.T, currency string, budgetMinor int64) engine.Input {
	t.Helper()
	return engine.Input{
		Query:  engine.Query{TripID: "trip", Nights: 7, Travelers: 2, Currency: money.Currency(currency), Country: "PT"},
		Budget: amount(t, budgetMinor, currency), HasBudget: true, On: calculatedAt,
	}
}

func TestCategoriesAreAggregatedInTheTravellerCurrency(t *testing.T) {
	source := estimates{values: []engine.Estimate{
		estimate(t, domain.Stay, 90_000, 100_000, 120_000, "EUR"),
		estimate(t, domain.Food, 20_000, 28_000, 35_000, "EUR"),
	}}
	result, err := engine.New(source, nil, nil, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "EUR", 400_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.Currency.String() != "EUR" || len(result.Lines) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.Estimated.Expected.Minor() != 128_000 {
		t.Fatalf("expected total = %d", result.Estimated.Expected.Minor())
	}
	if !result.HasRemaining || result.Remaining.Minor() != 272_000 {
		t.Fatalf("remaining = %+v", result.Remaining)
	}
}

func TestForeignCurrencyEstimatesAreConvertedThroughThePort(t *testing.T) {
	source := estimates{values: []engine.Estimate{estimate(t, domain.Stay, 100, 200, 300, "USD")}}
	converter := &rates{factor: 1500}
	result, err := engine.New(source, nil, converter, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "NGN", 1_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if converter.calls == 0 {
		t.Fatal("the exchange rate port was never consulted")
	}
	if result.Lines[0].Estimated.Expected.Minor() != 300_000 {
		t.Fatalf("converted expected = %d", result.Lines[0].Estimated.Expected.Minor())
	}
	if result.Lines[0].Estimated.Expected.Currency().String() != "NGN" {
		t.Fatal("the traveller's currency was not used")
	}
}

func TestMissingExchangeRateProducesNoEstimateRatherThanAGuess(t *testing.T) {
	source := estimates{values: []engine.Estimate{estimate(t, domain.Stay, 100, 200, 300, "USD")}}
	result, err := engine.New(source, nil, nil, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "NGN", 1_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.Lines[0].HasEstimate {
		t.Fatal("a cross currency estimate was invented without a rate")
	}
	if result.Lines[0].Confidence != domain.Unknown {
		t.Fatalf("confidence = %q, want unknown", result.Lines[0].Confidence)
	}
	failing, err := engine.New(source, nil, &rates{err: errors.New("rates down")}, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "NGN", 1_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if failing.Lines[0].HasEstimate {
		t.Fatal("a failing rate provider produced an estimate")
	}
}

func TestBookedAmountsAreMergedWithoutReplacingEstimates(t *testing.T) {
	source := estimates{values: []engine.Estimate{estimate(t, domain.Stay, 90_000, 98_000, 110_000, "EUR")}}
	reservations := booked{values: []engine.Booked{
		{Category: domain.Stay, Amount: amount(t, 94_000, "EUR"), Source: "confirmation"},
		{Category: domain.Flight, Amount: amount(t, 210_000, "EUR"), Source: "confirmation"},
	}}
	result, err := engine.New(source, reservations, nil, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "EUR", 500_000))
	if err != nil {
		t.Fatal(err)
	}
	var stay, flight domain.Line
	for _, line := range result.Lines {
		switch line.Category {
		case domain.Stay:
			stay = line
		case domain.Flight:
			flight = line
		}
	}
	if stay.Estimated.Expected.Minor() != 98_000 || stay.Booked.Minor() != 94_000 {
		t.Fatalf("stay line = %+v", stay)
	}
	if !flight.HasBooked || flight.HasEstimate {
		t.Fatalf("a booked-only category was mishandled: %+v", flight)
	}
	if result.Booked.Minor() != 304_000 {
		t.Fatalf("booked total = %d", result.Booked.Minor())
	}
}

func TestItineraryActivitiesBecomeTheirOwnCategory(t *testing.T) {
	values := input(t, "EUR", 400_000)
	values.Activities = []money.Amount{amount(t, 2_500, "EUR"), amount(t, 4_000, "EUR")}
	result, err := engine.New(estimates{}, nil, nil, clock.NewFixed(calculatedAt)).Calculate(t.Context(), values)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 1 || result.Lines[0].Category != domain.Activities {
		t.Fatalf("lines = %+v", result.Lines)
	}
	if result.Lines[0].Estimated.Expected.Minor() != 6_500 {
		t.Fatalf("activity total = %d", result.Lines[0].Estimated.Expected.Minor())
	}
	if result.Lines[0].Source != "itinerary" {
		t.Fatal("activity costs lost their source")
	}
}

func TestNoProvidersProducesAnHonestEmptyBudget(t *testing.T) {
	result, err := engine.New(nil, nil, nil, clock.NewFixed(calculatedAt)).Calculate(t.Context(), input(t, "NGN", 500_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.HasEstimate || result.HasRemaining || len(result.Lines) != 0 {
		t.Fatalf("an empty budget was invented: %+v", result)
	}
	if result.Confidence != domain.Unknown {
		t.Fatalf("confidence = %q", result.Confidence)
	}
	if result.Currency.String() != "NGN" || result.Budget.Minor() != 500_000_000 {
		t.Fatal("the traveller's own budget was lost")
	}
}

func TestFailingProvidersDegradeInsteadOfFailing(t *testing.T) {
	result, err := engine.New(estimates{err: errors.New("cost provider down")},
		booked{err: errors.New("bookings down")}, nil, clock.NewFixed(calculatedAt)).
		Calculate(t.Context(), input(t, "EUR", 400_000))
	if err != nil {
		t.Fatalf("a failing provider broke the budget: %v", err)
	}
	if result.HasEstimate {
		t.Fatal("a failing provider produced an estimate")
	}
}

func TestInventedCategoriesAreDropped(t *testing.T) {
	source := estimates{values: []engine.Estimate{
		{Category: "souvenirs", Minimum: amount(t, 1, "EUR"), Expected: amount(t, 2, "EUR"), Maximum: amount(t, 3, "EUR")},
		estimate(t, domain.Food, 10, 20, 30, "EUR"),
	}}
	result, err := engine.New(source, nil, nil, clock.NewFixed(calculatedAt)).Calculate(t.Context(), input(t, "EUR", 1000))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 1 || result.Lines[0].Category != domain.Food {
		t.Fatalf("lines = %+v", result.Lines)
	}
}
