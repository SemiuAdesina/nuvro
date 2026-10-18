package discovery

import (
	"slices"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Preferences struct {
	Interests        []string
	Weather          string
	Crowds           string
	Pace             string
	FlightTolerance  string
	Paperwork        string
	MaxFlightMinutes int
}

type Intent struct {
	Traveler    Traveler
	Origin      Origin
	Party       Party
	Window      Window
	Budget      money.Amount
	Preferences Preferences
	Sources     map[string]string
}

const (
	SourceRequest  = "request"
	SourceProfile  = "profile"
	SourceInferred = "inferred"
)

func (i Intent) Validate(now time.Time) error {
	if err := i.Traveler.Validate(); err != nil {
		return err
	}
	if err := i.Origin.Validate(); err != nil {
		return err
	}
	if err := i.Party.Validate(); err != nil {
		return err
	}
	if err := i.Window.Validate(now); err != nil {
		return err
	}
	if err := validateBudget(i.Budget); err != nil {
		return err
	}
	return i.Preferences.Validate()
}

func (p Preferences) Validate() error {
	if !isOneOf(p.Weather, "", "any", "warm", "mild", "cold") {
		return ErrInvalid
	}
	if !isOneOf(p.Crowds, "", "any", "quiet", "lively") {
		return ErrInvalid
	}
	if !isOneOf(p.Pace, "", "relaxed", "balanced", "packed") {
		return ErrInvalid
	}
	if !isOneOf(p.FlightTolerance, "", "any", "fewer_stops", "shortest") {
		return ErrInvalid
	}
	if !isOneOf(p.Paperwork, "", "any", "keep_it_simple") {
		return ErrInvalid
	}
	if p.MaxFlightMinutes < 0 || p.MaxFlightMinutes > 2880 || len(p.Interests) > 50 {
		return ErrInvalid
	}
	return nil
}

func (i Intent) MissingContext() []string {
	var missing []string
	if !i.Traveler.HasPassport() {
		missing = append(missing, "passport")
	}
	if i.Traveler.ResidenceCountry == "" {
		missing = append(missing, "residence")
	}
	if i.Origin.DepartureAirport == "" {
		missing = append(missing, "departure_airport")
	}
	return missing
}

func NormalizeTags(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		tag := strings.ToLower(strings.TrimSpace(value))
		if tag == "" || len(tag) > 40 {
			continue
		}
		if _, duplicate := seen[tag]; duplicate {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	slices.Sort(normalized)
	return normalized
}

func isOneOf(value string, allowed ...string) bool {
	return slices.Contains(allowed, value)
}
