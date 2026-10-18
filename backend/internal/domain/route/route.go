package route

import (
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

var ErrInvalid = errors.New("route: invalid evidence")

type Mode string

const (
	Walk    Mode = "walk"
	Drive   Mode = "drive"
	Transit Mode = "transit"
	Bike    Mode = "bike"
)

func (m Mode) Valid() bool {
	switch m {
	case Walk, Drive, Transit, Bike:
		return true
	default:
		return false
	}
}

type Status string

const (
	Found       Status = "found"
	NotFound    Status = "not_found"
	Unsupported Status = "unsupported"
	Unavailable Status = "unavailable"
)

type Confidence string

const (
	High    Confidence = "high"
	Medium  Confidence = "medium"
	Low     Confidence = "low"
	Unknown Confidence = "unknown"
)

type Evidence struct {
	Status          Status
	Origin          geo.Point
	Destination     geo.Point
	Mode            Mode
	DistanceMeters  int
	DurationSeconds int
	ComputedAt      time.Time
	Confidence      Confidence
	Source          evidence.Source
}

func UnavailableEvidence(origin, destination geo.Point, mode Mode) Evidence {
	return Evidence{
		Status: Unavailable, Origin: origin, Destination: destination, Mode: mode,
		DistanceMeters: -1, DurationSeconds: -1, Confidence: Unknown,
	}
}

func (e Evidence) Known() bool { return e.Status == Found && e.DurationSeconds >= 0 }

func (e Evidence) Duration() time.Duration {
	if !e.Known() {
		return 0
	}
	return time.Duration(e.DurationSeconds) * time.Second
}

func (e Evidence) Validate() error {
	if !e.Mode.Valid() {
		return ErrInvalid
	}
	if e.Status == Found && (e.DurationSeconds < 0 || e.DistanceMeters < 0) {
		return ErrInvalid
	}
	return nil
}

type Leg struct {
	FromIndex int
	ToIndex   int
	Evidence  Evidence
}

type Matrix struct {
	Points []geo.Point
	Mode   Mode
	Legs   []Leg
}

func (m Matrix) Between(from, to int) (Evidence, bool) {
	for _, leg := range m.Legs {
		if leg.FromIndex == from && leg.ToIndex == to {
			return leg.Evidence, true
		}
	}
	return Evidence{}, false
}
