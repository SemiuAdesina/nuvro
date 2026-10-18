package budget

import (
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Trip struct {
	Currency     money.Currency
	Budget       money.Amount
	HasBudget    bool
	Lines        []Line
	Estimated    Range
	HasEstimate  bool
	Booked       money.Amount
	Remaining    money.Amount
	HasRemaining bool
	Confidence   Confidence
	CalculatedAt time.Time
}

func Aggregate(currency money.Currency, budget money.Amount, hasBudget bool, lines []Line, calculatedAt time.Time) (Trip, error) {
	ordered := slices.Clone(lines)
	slices.SortStableFunc(ordered, func(left, right Line) int {
		return slices.Index(Categories(), left.Category) - slices.Index(Categories(), right.Category)
	})
	result := Trip{Currency: currency, Budget: budget, HasBudget: hasBudget, Lines: ordered, CalculatedAt: calculatedAt.UTC()}
	zero, err := money.New(0, currency)
	if err != nil {
		return Trip{}, err
	}
	total := Range{Minimum: zero, Expected: zero, Maximum: zero}
	booked := zero
	confidence := High
	for _, line := range ordered {
		if validationErr := line.Validate(); validationErr != nil {
			return Trip{}, validationErr
		}
		if currencyErr := assertCurrency(line, currency); currencyErr != nil {
			return Trip{}, currencyErr
		}
		total, booked, confidence, err = accumulate(total, booked, confidence, line)
		if err != nil {
			return Trip{}, err
		}
	}
	result.Estimated = total
	result.HasEstimate = len(ordered) > 0
	result.Booked = booked
	result.Confidence = confidence
	if !result.HasEstimate {
		result.Confidence = Unknown
	}
	return withRemaining(result)
}

func accumulate(total Range, booked money.Amount, confidence Confidence, line Line) (Range, money.Amount, Confidence, error) {
	if line.HasEstimate {
		combined, err := total.Add(line.Estimated)
		if err != nil {
			return Range{}, money.Amount{}, confidence, err
		}
		total = combined
	}
	if line.HasBooked {
		combined, err := booked.Add(line.Booked)
		if err != nil {
			return Range{}, money.Amount{}, confidence, err
		}
		booked = combined
	}
	if line.Confidence.Rank() < confidence.Rank() {
		confidence = line.Confidence
	}
	return total, booked, confidence, nil
}

func withRemaining(result Trip) (Trip, error) {
	if !result.HasBudget || !result.HasEstimate {
		return result, nil
	}
	spend := result.Estimated.Expected
	if result.Booked.Minor() > 0 {
		combined, err := committedSpend(result)
		if err != nil {
			return Trip{}, err
		}
		spend = combined
	}
	remaining, err := result.Budget.Subtract(spend)
	if err != nil {
		return Trip{}, err
	}
	result.Remaining = remaining
	result.HasRemaining = true
	return result, nil
}

func committedSpend(result Trip) (money.Amount, error) {
	zero, err := money.New(0, result.Currency)
	if err != nil {
		return money.Amount{}, err
	}
	spend := zero
	for _, line := range result.Lines {
		amount, known := line.Effective()
		if !known {
			continue
		}
		spend, err = spend.Add(amount)
		if err != nil {
			return money.Amount{}, err
		}
	}
	return spend, nil
}

func assertCurrency(line Line, currency money.Currency) error {
	if line.HasEstimate && line.Estimated.Expected.Currency() != currency {
		return ErrCurrencyMismatch
	}
	if line.HasBooked && line.Booked.Currency() != currency {
		return ErrCurrencyMismatch
	}
	return nil
}
