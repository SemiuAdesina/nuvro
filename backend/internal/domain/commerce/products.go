package commerce

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

type FlightOffer struct {
	Offer
	Stops           int
	DurationMinutes int
	LayoverMinutes  int
	DepartAt        time.Time
	ArriveAt        time.Time
	Carrier         string
	BaggageIncluded bool
	Changeable      bool
}

func (f FlightOffer) Validate() error {
	if err := f.Offer.Validate(); err != nil {
		return err
	}
	if f.Product != Flight || f.Stops < 0 || f.DurationMinutes <= 0 || f.LayoverMinutes < 0 {
		return ErrInvalid
	}
	return nil
}

func (f FlightOffer) Signature() string {
	return f.Carrier + "|" + f.DepartAt.UTC().Format(time.RFC3339) + "|" +
		f.ArriveAt.UTC().Format(time.RFC3339) + "|" + itoa(f.Stops)
}

type StayOffer struct {
	Offer
	PropertyID       string
	Location         geo.Point
	Rating           float64
	ReviewCount      int
	Neighbourhood    string
	FreeCancellation bool
	Nights           int
}

func (s StayOffer) Validate() error {
	if err := s.Offer.Validate(); err != nil {
		return err
	}
	if s.Product != Stay || s.Nights <= 0 || s.Rating < 0 || s.Rating > 5 || s.ReviewCount < 0 {
		return ErrInvalid
	}
	return s.Location.Validate()
}

func (s StayOffer) Signature() string { return s.PropertyID }

type ActivityOffer struct {
	Offer
	PlaceID         string
	Location        geo.Point
	DurationMinutes int
	StartsAt        time.Time
	Outdoor         bool
	Rating          float64
}

func (a ActivityOffer) Validate() error {
	if err := a.Offer.Validate(); err != nil {
		return err
	}
	if a.Product != Activity || a.DurationMinutes <= 0 || a.Rating < 0 || a.Rating > 5 {
		return ErrInvalid
	}
	return a.Location.Validate()
}

func (a ActivityOffer) Signature() string { return a.PlaceID }

type TransferOffer struct {
	Offer
	Mode            string
	DurationMinutes int
	Passengers      int
	PickupAt        time.Time
}

func (t TransferOffer) Validate() error {
	if err := t.Offer.Validate(); err != nil {
		return err
	}
	if t.Product != Transfer || t.DurationMinutes <= 0 || t.Passengers < 1 {
		return ErrInvalid
	}
	return nil
}

func (t TransferOffer) Signature() string { return t.Mode + "|" + itoa(t.DurationMinutes) }

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	if negative {
		return "-" + digits
	}
	return digits
}
