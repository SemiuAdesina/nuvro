package understanding

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Supplied struct {
	Interests         []string
	Weather           string
	Crowds            string
	Pace              string
	FlightTolerance   string
	Paperwork         string
	MaxFlightMinutes  *int
	Budget            *money.Amount
	DepartureAirport  string
	ResidenceCountry  string
	PassportCountries []string
	ResidencePermits  []string
}

type Input struct {
	Profile  discovery.Intent
	Supplied Supplied
	Party    discovery.Party
	Window   discovery.Window
	Query    string
	Now      time.Time
}

type Inferred struct {
	Interests       []string
	Weather         string
	Crowds          string
	Pace            string
	FlightTolerance string
	Paperwork       string
}
