package budget_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

var calculatedAt = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func line(t *testing.T, category budget.Category, minimum, expected, maximum int64, currency string) budget.Line {
	t.Helper()
	amounts, err := budget.NewRange(amount(t, minimum, currency), amount(t, expected, currency), amount(t, maximum, currency))
	if err != nil {
		t.Fatal(err)
	}
	return budget.Line{Category: category, Estimated: amounts, HasEstimate: true, Confidence: budget.Medium}
}

func TestCategoriesAggregateInStableOrder(t *testing.T) {
	lines := []budget.Line{
		line(t, budget.Food, 20_000, 30_000, 40_000, "EUR"),
		line(t, budget.Flight, 200_000, 240_000, 300_000, "EUR"),
		line(t, budget.Stay, 90_000, 100_000, 120_000, "EUR"),
	}
	result, err := budget.Aggregate(money.Currency("EUR"), amount(t, 600_000, "EUR"), true, lines, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Lines[0].Category != budget.Flight || result.Lines[1].Category != budget.Stay || result.Lines[2].Category != budget.Food {
		t.Fatalf("categories are not in canonical order: %v", result.Lines)
	}
	if result.Estimated.Expected.Minor() != 370_000 {
		t.Fatalf("expected total = %d", result.Estimated.Expected.Minor())
	}
	if result.Estimated.Minimum.Minor() != 310_000 || result.Estimated.Maximum.Minor() != 460_000 {
		t.Fatalf("range did not aggregate: %+v", result.Estimated)
	}
	if !result.HasRemaining || result.Remaining.Minor() != 230_000 {
		t.Fatalf("remaining = %+v", result.Remaining)
	}
}

func TestTravellerCurrencyIsPreserved(t *testing.T) {
	result, err := budget.Aggregate(money.Currency("NGN"), amount(t, 500_000_000, "NGN"), true,
		[]budget.Line{line(t, budget.Stay, 100_000_000, 120_000_000, 150_000_000, "NGN")}, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Currency.String() != "NGN" || result.Estimated.Expected.Currency().String() != "NGN" {
		t.Fatal("the traveller's currency was replaced")
	}
	if result.Remaining.Currency().String() != "NGN" {
		t.Fatal("remaining is not in the traveller's currency")
	}
}

func TestMixedCurrenciesAreRefusedRatherThanGuessed(t *testing.T) {
	lines := []budget.Line{line(t, budget.Stay, 1, 2, 3, "EUR")}
	if _, err := budget.Aggregate(money.Currency("NGN"), amount(t, 100, "NGN"), true, lines, calculatedAt); err == nil {
		t.Fatal("a line in another currency was silently aggregated")
	}
}

func TestBookedAndEstimatedStaySeparate(t *testing.T) {
	stay := line(t, budget.Stay, 90_000, 98_000, 110_000, "EUR")
	stay.Booked = amount(t, 94_000, "EUR")
	stay.HasBooked = true
	stay.Confidence = budget.High
	result, err := budget.Aggregate(money.Currency("EUR"), amount(t, 300_000, "EUR"), true,
		[]budget.Line{stay, line(t, budget.Food, 20_000, 30_000, 40_000, "EUR")}, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	stored := result.Lines[0]
	if stored.Estimated.Expected.Minor() != 98_000 || stored.Booked.Minor() != 94_000 {
		t.Fatalf("booked overwrote the estimate: %+v", stored)
	}
	if result.Booked.Minor() != 94_000 {
		t.Fatalf("booked total = %d", result.Booked.Minor())
	}
	effective, known := stored.Effective()
	if !known || effective.Minor() != 94_000 {
		t.Fatal("the booked amount should take precedence when spending is computed")
	}
	if result.Remaining.Minor() != 300_000-94_000-30_000 {
		t.Fatalf("remaining did not use booked over estimated: %d", result.Remaining.Minor())
	}
}

func TestConfidenceDropsToTheWeakestLine(t *testing.T) {
	weak := line(t, budget.Food, 1, 2, 3, "EUR")
	weak.Confidence = budget.Low
	strong := line(t, budget.Stay, 1, 2, 3, "EUR")
	strong.Confidence = budget.High
	result, err := budget.Aggregate(money.Currency("EUR"), amount(t, 100, "EUR"), true, []budget.Line{strong, weak}, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Confidence != budget.Low {
		t.Fatalf("confidence = %q, want low", result.Confidence)
	}
}

func TestNoEvidenceProducesNoInventedTotal(t *testing.T) {
	result, err := budget.Aggregate(money.Currency("EUR"), amount(t, 100_000, "EUR"), true, nil, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if result.HasEstimate || result.HasRemaining {
		t.Fatalf("an estimate was invented: %+v", result)
	}
	if result.Confidence != budget.Unknown {
		t.Fatalf("confidence = %q, want unknown", result.Confidence)
	}
}

func TestRangeAndLineValidation(t *testing.T) {
	if _, err := budget.NewRange(amount(t, 300, "EUR"), amount(t, 200, "EUR"), amount(t, 100, "EUR")); err == nil {
		t.Fatal("a reversed range was accepted")
	}
	if _, err := budget.NewRange(amount(t, 1, "EUR"), amount(t, 2, "NGN"), amount(t, 3, "EUR")); err == nil {
		t.Fatal("a mixed-currency range was accepted")
	}
	invalid := budget.Line{Category: "sightseeing", Confidence: budget.High}
	if invalid.Validate() == nil {
		t.Fatal("an invented category was accepted")
	}
	if (budget.Line{Category: budget.Food, Confidence: "certain"}).Validate() == nil {
		t.Fatal("an invented confidence was accepted")
	}
}

func TestLargeAmountsDoNotOverflowOrLoseAccuracy(t *testing.T) {
	huge := line(t, budget.Stay, 900_000_000_000, 950_000_000_000, 999_000_000_000, "NGN")
	result, err := budget.Aggregate(money.Currency("NGN"), amount(t, 1_000_000_000_000, "NGN"), true, []budget.Line{huge}, calculatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Estimated.Expected.Minor() != 950_000_000_000 {
		t.Fatalf("large minor units lost accuracy: %d", result.Estimated.Expected.Minor())
	}
	if result.Remaining.Minor() != 50_000_000_000 {
		t.Fatalf("remaining = %d", result.Remaining.Minor())
	}
}
