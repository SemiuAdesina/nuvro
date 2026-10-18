package budget

import (
	"context"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Engine struct {
	estimates EstimateProvider
	booked    BookedReader
	rates     Rates
	timer     clock.Clock
}

func New(estimates EstimateProvider, booked BookedReader, rates Rates, timer clock.Clock) *Engine {
	return &Engine{estimates: estimates, booked: booked, rates: rates, timer: timer}
}

type Input struct {
	Query      Query
	Budget     money.Amount
	HasBudget  bool
	Activities []money.Amount
	On         time.Time
}

func (e *Engine) Calculate(ctx context.Context, input Input) (domain.Trip, error) {
	lines := make([]domain.Line, 0, len(domain.Categories()))
	lines = append(lines, e.estimateLines(ctx, input)...)
	if activities, ok := e.activityLine(input); ok {
		lines = append(lines, activities)
	}
	lines = e.applyBooked(ctx, input, lines)
	return domain.Aggregate(input.Query.Currency, input.Budget, input.HasBudget, lines, e.timer.Now())
}

func (e *Engine) estimateLines(ctx context.Context, input Input) []domain.Line {
	if e.estimates == nil {
		return nil
	}
	found, err := e.estimates.Estimates(ctx, input.Query)
	if err != nil {
		return nil
	}
	lines := make([]domain.Line, 0, len(found))
	for _, estimate := range found {
		line, ok := e.line(ctx, estimate, input)
		if !ok {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func (e *Engine) line(ctx context.Context, estimate Estimate, input Input) (domain.Line, bool) {
	if !estimate.Category.Valid() {
		return domain.Line{}, false
	}
	minimum, first := e.convert(ctx, estimate.Minimum, input)
	expected, second := e.convert(ctx, estimate.Expected, input)
	maximum, third := e.convert(ctx, estimate.Maximum, input)
	if !first || !second || !third {
		return domain.Line{
			Category: estimate.Category, Confidence: domain.Unknown,
			Source: estimate.Source, ObservedAt: estimate.ObservedAt,
		}, true
	}
	amounts, err := domain.NewRange(minimum, expected, maximum)
	if err != nil {
		return domain.Line{}, false
	}
	return domain.Line{
		Category: estimate.Category, Estimated: amounts, HasEstimate: true,
		Confidence: confidenceOrUnknown(estimate.Confidence), Source: estimate.Source, ObservedAt: estimate.ObservedAt,
	}, true
}

func (e *Engine) activityLine(input Input) (domain.Line, bool) {
	if len(input.Activities) == 0 {
		return domain.Line{}, false
	}
	total, err := money.New(0, input.Query.Currency)
	if err != nil {
		return domain.Line{}, false
	}
	for _, amount := range input.Activities {
		if amount.Currency() != input.Query.Currency {
			return domain.Line{}, false
		}
		total, err = total.Add(amount)
		if err != nil {
			return domain.Line{}, false
		}
	}
	amounts, err := domain.NewRange(total, total, total)
	if err != nil {
		return domain.Line{}, false
	}
	return domain.Line{
		Category: domain.Activities, Estimated: amounts, HasEstimate: true,
		Confidence: domain.Medium, Source: "itinerary", ObservedAt: e.timer.Now().UTC(),
	}, true
}

func (e *Engine) applyBooked(ctx context.Context, input Input, lines []domain.Line) []domain.Line {
	if e.booked == nil {
		return lines
	}
	found, err := e.booked.Booked(ctx, input.Query.TripID)
	if err != nil {
		return lines
	}
	for _, record := range found {
		amount, ok := e.convert(ctx, record.Amount, input)
		if !ok {
			continue
		}
		lines = merge(lines, record.Category, amount)
	}
	return lines
}

func merge(lines []domain.Line, category domain.Category, amount money.Amount) []domain.Line {
	for index := range lines {
		if lines[index].Category != category {
			continue
		}
		lines[index].Booked = amount
		lines[index].HasBooked = true
		return lines
	}
	return append(lines, domain.Line{
		Category: category, Booked: amount, HasBooked: true, Confidence: domain.High, Source: "booking",
	})
}

func (e *Engine) convert(ctx context.Context, amount money.Amount, input Input) (money.Amount, bool) {
	if amount.Currency() == input.Query.Currency {
		return amount, true
	}
	if e.rates == nil {
		return money.Amount{}, false
	}
	converted, err := e.rates.Convert(ctx, amount, input.Query.Currency, input.On)
	if err != nil {
		return money.Amount{}, false
	}
	return converted, true
}

func confidenceOrUnknown(confidence domain.Confidence) domain.Confidence {
	if confidence.Valid() {
		return confidence
	}
	return domain.Unknown
}
