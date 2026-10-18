package feasibility

import (
	"errors"
	"time"
)

var ErrInvalid = errors.New("feasibility: invalid result")

type Overall string

const (
	Strong         Overall = "strong"
	Viable         Overall = "viable"
	Challenging    Overall = "challenging"
	Infeasible     Overall = "not_feasible"
	UnknownOverall Overall = "unknown"
)

type FlightStatus string

const (
	FlightAvailable   FlightStatus = "available"
	FlightUnavailable FlightStatus = "unavailable"
	FlightUnknown     FlightStatus = "unknown"
)

type Flight struct {
	Status          FlightStatus
	Stops           int
	DurationMinutes int
	HasPrice        bool
	PriceMinor      int64
	PriceCurrency   string
	Confidence      Confidence
	ObservedAt      time.Time
}

func UnknownFlight() Flight {
	return Flight{Status: FlightUnknown, Stops: -1, DurationMinutes: -1, Confidence: UnknownConfidence}
}

type Result struct {
	Access        Access
	Affordability Affordability
	Flight        Flight
	Overall       Overall
	Cautions      []string
	EvaluatedAt   time.Time
}

func (r Result) Validate() error {
	if !r.Access.Confidence.Valid() || !r.Affordability.Confidence.Valid() || !r.Flight.Confidence.Valid() {
		return ErrInvalid
	}
	switch r.Overall {
	case Strong, Viable, Challenging, Infeasible, UnknownOverall:
		return nil
	default:
		return ErrInvalid
	}
}

func (r Result) Rank() int {
	switch r.Overall {
	case Strong:
		return 4
	case Viable:
		return 3
	case Challenging:
		return 2
	case UnknownOverall:
		return 1
	default:
		return 0
	}
}
