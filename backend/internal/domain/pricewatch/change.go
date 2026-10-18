package pricewatch

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Direction string

const (
	Dropped     Direction = "price_dropped"
	Increased   Direction = "price_increased"
	Unchanged   Direction = "unchanged"
	Unavailable Direction = "availability_changed"
)

type Change struct {
	Direction        Direction
	Baseline         money.Amount
	Current          money.Amount
	AbsoluteMinor    int64
	PercentagePoints int
	Meaningful       bool
	ObservedAt       time.Time
}

func Compare(baseline, current money.Amount, threshold Threshold, observedAt time.Time) (Change, error) {
	if baseline.Currency() != current.Currency() {
		return Change{}, ErrInvalid
	}
	delta := current.Minor() - baseline.Minor()
	change := Change{
		Baseline: baseline, Current: current, AbsoluteMinor: absolute(delta),
		ObservedAt: observedAt.UTC(), Direction: Unchanged,
	}
	if baseline.Minor() > 0 {
		change.PercentagePoints = int((absolute(delta) * 100) / baseline.Minor())
	}
	switch {
	case delta < 0:
		change.Direction = Dropped
	case delta > 0:
		change.Direction = Increased
	}
	change.Meaningful = change.Direction != Unchanged &&
		change.PercentagePoints >= threshold.PercentageDrop &&
		change.AbsoluteMinor >= threshold.AbsoluteMinor
	return change, nil
}

func UnavailableChange(baseline money.Amount, observedAt time.Time) Change {
	return Change{
		Direction: Unavailable, Baseline: baseline, Current: baseline,
		Meaningful: true, ObservedAt: observedAt.UTC(),
	}
}

func absolute(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

type Observation struct {
	WatchID    string
	Price      money.Amount
	Available  bool
	ObservedAt time.Time
}
