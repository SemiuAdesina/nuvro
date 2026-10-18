package itinerary

import (
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

var ErrInvalid = errors.New("itinerary: invalid item")

type Kind string

const (
	Activity Kind = "activity"
	Meal     Kind = "meal"
	Rest     Kind = "rest"
	Transfer Kind = "transfer"
	Note     Kind = "note"
)

func (k Kind) Valid() bool {
	switch k {
	case Activity, Meal, Rest, Transfer, Note:
		return true
	default:
		return false
	}
}

type Status string

const (
	Planned   Status = "planned"
	Reserved  Status = "reserved"
	Completed Status = "completed"
	Skipped   Status = "skipped"
	Cancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case Planned, Reserved, Completed, Skipped, Cancelled:
		return true
	default:
		return false
	}
}

func (s Status) CanTransitionTo(next Status) bool {
	if !next.Valid() {
		return false
	}
	switch s {
	case Planned:
		return next == Reserved || next == Completed || next == Skipped || next == Cancelled
	case Reserved:
		return next == Completed || next == Skipped || next == Cancelled
	default:
		return next == s
	}
}

type Source string

const (
	Generated Source = "generated"
	Traveller Source = "traveller"
)

type Item struct {
	ID                 string
	TripID             string
	Day                int
	StartAt            time.Time
	EndAt              time.Time
	Kind               Kind
	Title              string
	PlaceID            string
	Category           place.Category
	Location           geo.Point
	RouteFromPrevious  *route.Evidence
	WeatherImpact      weather.Impact
	EstimatedCostMinor int64
	CostCurrency       string
	Status             Status
	Source             Source
	UserLocked         bool
	ReservationLocked  bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (i Item) Validate() error {
	if i.ID == "" || i.TripID == "" || i.Title == "" || i.Day < 1 {
		return ErrInvalid
	}
	if !i.Kind.Valid() || !i.Status.Valid() {
		return ErrInvalid
	}
	if i.StartAt.IsZero() || i.EndAt.IsZero() || !i.EndAt.After(i.StartAt) {
		return ErrInvalid
	}
	if i.EstimatedCostMinor < 0 {
		return ErrInvalid
	}
	return nil
}

func (i Item) Duration() time.Duration { return i.EndAt.Sub(i.StartAt) }

func (i Item) Locked() bool { return i.UserLocked || i.ReservationLocked }

func (i Item) Overlaps(other Item) bool {
	return i.StartAt.Before(other.EndAt) && other.StartAt.Before(i.EndAt)
}
