package discovery

import (
	"errors"
	"regexp"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

var ErrInvalid = errors.New("discovery: invalid request")

const maximumTripNights = 90
const maximumTravellers = 20

var (
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
	airportPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type Traveler struct {
	ResidenceCountry  string
	PassportCountries []string
	ResidencePermits  []string
}

type Origin struct {
	DepartureAirport string
}

type Party struct {
	Adults   int
	Children int
}

type Window struct {
	Start time.Time
	End   time.Time
}

func (t Traveler) HasPassport() bool { return len(t.PassportCountries) > 0 }

func (t Traveler) Validate() error {
	if t.ResidenceCountry != "" && !countryPattern.MatchString(t.ResidenceCountry) {
		return ErrInvalid
	}
	for _, code := range append(append([]string{}, t.PassportCountries...), t.ResidencePermits...) {
		if !countryPattern.MatchString(code) {
			return ErrInvalid
		}
	}
	return nil
}

func (o Origin) Validate() error {
	if o.DepartureAirport != "" && !airportPattern.MatchString(o.DepartureAirport) {
		return ErrInvalid
	}
	return nil
}

func (p Party) Validate() error {
	if p.Adults < 1 || p.Children < 0 || p.Adults+p.Children > maximumTravellers {
		return ErrInvalid
	}
	return nil
}

func (p Party) Size() int { return p.Adults + p.Children }

func (w Window) Validate(now time.Time) error {
	if w.Start.IsZero() || w.End.IsZero() || w.End.Before(w.Start) {
		return ErrInvalid
	}
	if w.Start.Before(now.UTC().Truncate(24 * time.Hour)) {
		return ErrInvalid
	}
	if w.Nights() > maximumTripNights {
		return ErrInvalid
	}
	return nil
}

func (w Window) Nights() int {
	return int(w.End.UTC().Sub(w.Start.UTC()).Hours() / 24)
}

func validateBudget(budget money.Amount) error {
	if budget.Minor() <= 0 {
		return ErrInvalid
	}
	if _, err := money.ParseCurrency(budget.Currency().String()); err != nil {
		return ErrInvalid
	}
	return nil
}
