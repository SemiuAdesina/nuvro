package feasibility

import (
	"context"
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

const percent = 100
const strongFitRatio = 70

var errRateUnavailable = errors.New("feasibility: exchange rate unavailable")

func (e *Engine) affordability(ctx context.Context, intent discovery.Intent, cost CostEvidence, available bool) feasibility.Affordability {
	if !available {
		return feasibility.UnknownAffordability(intent.Budget, "cost_evidence_unavailable")
	}
	target := intent.Budget.Currency()
	converted, conversion, err := e.convertRange(ctx, cost, target, intent.Window.Start)
	if err != nil {
		return feasibility.UnknownAffordability(intent.Budget, "currency_conversion_unavailable")
	}
	ratio := ratioOf(converted.Expected, intent.Budget)
	return feasibility.Affordability{
		Status:      affordabilityStatus(ratio, converted.Maximum, intent.Budget),
		Budget:      intent.Budget,
		Estimate:    converted,
		HasEstimate: true,
		Confidence:  confidenceOrUnknown(cost.Confidence),
		Reasons:     cost.Components,
		ObservedAt:  cost.ObservedAt,
		Ratio:       ratio,
		Conversion:  conversion,
	}
}

func (e *Engine) convertRange(ctx context.Context, cost CostEvidence, target money.Currency, on time.Time) (feasibility.CostRange, *money.Conversion, error) {
	if cost.Expected.Currency() == target {
		return feasibility.CostRange{Minimum: cost.Minimum, Expected: cost.Expected, Maximum: cost.Maximum}, nil, nil
	}
	if e.rates == nil {
		return feasibility.CostRange{}, nil, errRateUnavailable
	}
	quote, err := e.rates.Rate(ctx, cost.Expected.Currency(), target, on)
	if err != nil {
		return feasibility.CostRange{}, nil, errRateUnavailable
	}
	converted := feasibility.CostRange{}
	for _, pair := range []struct {
		source money.Amount
		into   *money.Amount
	}{{cost.Minimum, &converted.Minimum}, {cost.Expected, &converted.Expected}, {cost.Maximum, &converted.Maximum}} {
		value, rateErr := applyRate(pair.source, quote.Rate, target)
		if rateErr != nil {
			return feasibility.CostRange{}, nil, errRateUnavailable
		}
		*pair.into = value
	}
	conversion, err := money.NewConversion(cost.Expected, converted.Expected, quote.Rate, quote.ObservedAt, quote.Source)
	if err != nil {
		return feasibility.CostRange{}, nil, errRateUnavailable
	}
	return converted, &conversion, nil
}

func applyRate(source money.Amount, rate money.Rate, target money.Currency) (money.Amount, error) {
	minor, err := rate.Apply(source)
	if err != nil {
		return money.Amount{}, err
	}
	return money.New(minor, target)
}

func ratioOf(cost money.Amount, budget money.Amount) int {
	if budget.Minor() <= 0 {
		return -1
	}
	return int((cost.Minor() * percent) / budget.Minor())
}

func affordabilityStatus(ratio int, maximum money.Amount, budget money.Amount) feasibility.AffordabilityStatus {
	switch {
	case ratio < 0:
		return feasibility.InsufficientEvidence
	case ratio > percent:
		return feasibility.OverBudget
	case maximum.Minor() > budget.Minor():
		return feasibility.Tight
	case ratio <= strongFitRatio:
		return feasibility.StrongFit
	default:
		return feasibility.GoodFit
	}
}
