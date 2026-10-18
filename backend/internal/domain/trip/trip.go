package trip

import (
	"errors"
	"regexp"
	"time"
)

type Status string

const (
	Draft     Status = "draft"
	Planning  Status = "planning"
	Upcoming  Status = "upcoming"
	Active    Status = "active"
	Completed Status = "completed"
	Cancelled Status = "cancelled"
	Archived  Status = "archived"
)

var ErrInvalid = errors.New("trip: invalid trip")
var ErrTransition = errors.New("trip: invalid transition")
var airportPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Trip struct {
	ID            string
	OwnerID       string
	DestinationID *string
	OriginAirport *string
	StartDate     *time.Time
	EndDate       *time.Time
	Adults        int32
	Children      int32
	Currency      string
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (t Trip) Validate() error {
	if t.ID == "" || t.OwnerID == "" || t.Adults < 1 || t.Adults > 100 || t.Children < 0 || t.Children > 100 {
		return ErrInvalid
	}
	if err := t.validateContext(); err != nil {
		return err
	}
	switch t.Status {
	case Draft, Planning, Upcoming, Active, Completed, Cancelled, Archived:
		return nil
	default:
		return ErrInvalid
	}
}

func (s Status) Transition(next Status) (Status, error) {
	allowed := map[Status][]Status{
		Draft:     {Planning, Cancelled, Archived},
		Planning:  {Upcoming, Cancelled, Archived},
		Upcoming:  {Planning, Active, Cancelled},
		Active:    {Completed, Cancelled},
		Completed: {Archived}, Cancelled: {Archived},
	}
	for _, candidate := range allowed[s] {
		if next == candidate {
			return next, nil
		}
	}
	return s, ErrTransition
}

func (t Trip) validateContext() error {
	if !airportPattern.MatchString(t.Currency) {
		return ErrInvalid
	}
	if t.OriginAirport != nil && !airportPattern.MatchString(*t.OriginAirport) {
		return ErrInvalid
	}
	if t.DestinationID != nil && (len(*t.DestinationID) == 0 || len(*t.DestinationID) > 200) {
		return ErrInvalid
	}
	if (t.StartDate == nil) != (t.EndDate == nil) {
		return ErrInvalid
	}
	if t.StartDate != nil && t.EndDate.Before(*t.StartDate) {
		return ErrInvalid
	}
	return nil
}
