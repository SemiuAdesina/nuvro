package budget

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Estimate struct {
	Category   budget.Category
	Minimum    money.Amount
	Expected   money.Amount
	Maximum    money.Amount
	Confidence budget.Confidence
	Source     string
	ObservedAt time.Time
}

type Query struct {
	TripID    string
	Nights    int
	Travelers int
	Currency  money.Currency
	Country   string
}

type EstimateProvider interface {
	Estimates(context.Context, Query) ([]Estimate, error)
}

type BookedReader interface {
	Booked(context.Context, string) ([]Booked, error)
}

type Booked struct {
	Category budget.Category
	Amount   money.Amount
	Source   string
}

type Rates interface {
	Convert(ctx context.Context, amount money.Amount, target money.Currency, on time.Time) (money.Amount, error)
}
