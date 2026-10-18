package pricewatch

import (
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

var ErrInvalid = errors.New("pricewatch: invalid watch")

const MaximumPerUser = 20

type Status string

const (
	Active    Status = "active"
	Paused    Status = "paused"
	Completed Status = "completed"
	Ended     Status = "ended"
)

func (s Status) Valid() bool {
	switch s {
	case Active, Paused, Completed, Ended:
		return true
	default:
		return false
	}
}

type Target struct {
	Origin      string
	Destination string
	PropertyID  string
	PlaceID     string
	StartDate   time.Time
	EndDate     time.Time
	Adults      int
	Children    int
	Cabin       string
}

func (t Target) Validate(product commerce.ProductType) error {
	if t.Adults < 1 || t.Children < 0 {
		return ErrInvalid
	}
	if t.StartDate.IsZero() || t.EndDate.IsZero() || t.EndDate.Before(t.StartDate) {
		return ErrInvalid
	}
	if !t.identifiesProduct(product) {
		return ErrInvalid
	}
	return nil
}

func (t Target) identifiesProduct(product commerce.ProductType) bool {
	switch product {
	case commerce.Flight:
		return t.Origin != "" && t.Destination != ""
	case commerce.Stay:
		return t.PropertyID != ""
	case commerce.Activity:
		return t.PlaceID != ""
	default:
		return false
	}
}

type Threshold struct {
	PercentageDrop int
	AbsoluteMinor  int64
}

func DefaultThreshold() Threshold {
	return Threshold{PercentageDrop: 5, AbsoluteMinor: 1000}
}

func (t Threshold) Validate() error {
	if t.PercentageDrop < 1 || t.PercentageDrop > 90 || t.AbsoluteMinor < 0 {
		return ErrInvalid
	}
	return nil
}

type Watch struct {
	ID            string
	UserID        string
	TripID        string
	Product       commerce.ProductType
	Target        Target
	Baseline      money.Amount
	Threshold     Threshold
	Status        Status
	CreatedAt     time.Time
	LastCheckedAt time.Time
	NextCheckAt   time.Time
	ExpiresAt     time.Time
}

func (w Watch) Validate() error {
	if w.UserID == "" || !w.Product.Valid() || !w.Status.Valid() {
		return ErrInvalid
	}
	if w.Baseline.Minor() <= 0 || w.CreatedAt.IsZero() || w.ExpiresAt.IsZero() {
		return ErrInvalid
	}
	if !w.ExpiresAt.After(w.CreatedAt) {
		return ErrInvalid
	}
	if err := w.Threshold.Validate(); err != nil {
		return err
	}
	return w.Target.Validate(w.Product)
}

func (w Watch) Watching(now time.Time) bool {
	return w.Status == Active && w.ExpiresAt.After(now.UTC())
}

func (w Watch) Due(now time.Time) bool {
	return w.Watching(now) && (w.NextCheckAt.IsZero() || !w.NextCheckAt.After(now.UTC()))
}
