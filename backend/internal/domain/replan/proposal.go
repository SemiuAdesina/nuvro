package replan

import (
	"errors"
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

var ErrInvalid = errors.New("replan: invalid proposal")

type Trigger string

const (
	Weather  Trigger = "weather"
	Late     Trigger = "late"
	Tired    Trigger = "tired"
	Missed   Trigger = "missed"
	Closed   Trigger = "closed"
	FreeTime Trigger = "free_time"
	Manual   Trigger = "manual"
)

func Triggers() []Trigger {
	return []Trigger{Weather, Late, Tired, Missed, Closed, FreeTime, Manual}
}

func (t Trigger) Valid() bool { return slices.Contains(Triggers(), t) }

type Reason string

const (
	WeatherConflict            Reason = "weather_conflict"
	PlaceClosed                Reason = "place_closed"
	InsufficientTravelTime     Reason = "insufficient_travel_time"
	TravelerRequestedLowerPace Reason = "traveler_requested_lower_pace"
	MissedStartTime            Reason = "missed_start_time"
	BetterTimeWindow           Reason = "better_time_window"
	NoAlternativeFound         Reason = "no_alternative_found"
)

type ChangeKind string

const (
	Replace ChangeKind = "replace"
	Move    ChangeKind = "move"
	Remove  ChangeKind = "remove"
	Add     ChangeKind = "add"
)

type Change struct {
	Kind     ChangeKind
	Original *itinerary.Item
	Proposed *itinerary.Item
	Reasons  []Reason
}

type Constraint string

const (
	LockedItems   Constraint = "locked_items"
	Reservations  Constraint = "reservations"
	TripDates     Constraint = "trip_dates"
	BudgetCeiling Constraint = "budget_ceiling"
	Accessibility Constraint = "accessibility"
	TravelPace    Constraint = "travel_pace"
)

type Proposal struct {
	ID               string
	TripID           string
	Trigger          Trigger
	ItineraryVersion int
	Changes          []Change
	Reasons          []Reason
	Preserved        []Constraint
	BudgetDelta      money.Amount
	HasBudgetDelta   bool
	Violations       []itinerary.Violation
	Summary          string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	AppliedAt        *time.Time
}

func (p Proposal) Validate() error {
	if p.TripID == "" || !p.Trigger.Valid() || p.ItineraryVersion < 1 {
		return ErrInvalid
	}
	if p.CreatedAt.IsZero() || p.ExpiresAt.IsZero() || !p.ExpiresAt.After(p.CreatedAt) {
		return ErrInvalid
	}
	for _, change := range p.Changes {
		if err := change.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Change) Validate() error {
	if !c.shapeValid() {
		return ErrInvalid
	}
	return nil
}

func (c Change) shapeValid() bool {
	switch c.Kind {
	case Replace:
		return c.Original != nil && c.Proposed != nil
	case Move:
		return c.Original != nil && c.Proposed != nil && c.Original.ID == c.Proposed.ID
	case Remove:
		return c.Original != nil && c.Proposed == nil
	case Add:
		return c.Original == nil && c.Proposed != nil
	default:
		return false
	}
}

func (p Proposal) Stale(now time.Time) bool { return !p.ExpiresAt.After(now.UTC()) }

func (p Proposal) Applied() bool { return p.AppliedAt != nil }

func (p Proposal) Size() int { return len(p.Changes) }

func (p Proposal) Valid() bool { return len(p.Violations) == 0 }
