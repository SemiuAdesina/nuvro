package understanding

import (
	"context"
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
)

type Engine struct{ interpreter Interpreter }

func New(interpreter Interpreter) *Engine { return &Engine{interpreter: interpreter} }

func (e *Engine) Understand(ctx context.Context, input Input) (discovery.Intent, error) {
	intent := input.Profile
	intent.Party = input.Party
	intent.Window = input.Window
	intent.Sources = map[string]string{}
	applyProfileSources(&intent)
	applySupplied(&intent, input.Supplied)
	if inferred, ok := e.infer(ctx, input); ok {
		applyInferred(&intent, inferred)
	}
	intent.Preferences.Interests = discovery.NormalizeTags(intent.Preferences.Interests)
	if err := intent.Validate(input.Now); err != nil {
		return discovery.Intent{}, err
	}
	return intent, nil
}

func (e *Engine) infer(ctx context.Context, input Input) (Inferred, bool) {
	query := strings.TrimSpace(input.Query)
	if e.interpreter == nil || query == "" || len(query) > maximumQueryLength {
		return Inferred{}, false
	}
	inferred, err := e.interpreter.Interpret(ctx, query)
	if err != nil {
		return Inferred{}, false
	}
	inferred = inferred.normalized()
	if err := inferred.Validate(); err != nil {
		return Inferred{}, false
	}
	return inferred, true
}

func applyProfileSources(intent *discovery.Intent) {
	for _, field := range []string{"interests", "weather", "crowds", "pace", "flight_tolerance", "paperwork", "budget", "departure_airport", "residence", "passports"} {
		intent.Sources[field] = discovery.SourceProfile
	}
}

func applySupplied(intent *discovery.Intent, supplied Supplied) {
	assign(intent, "interests", len(supplied.Interests) > 0, func() { intent.Preferences.Interests = supplied.Interests })
	assign(intent, "weather", supplied.Weather != "", func() { intent.Preferences.Weather = supplied.Weather })
	assign(intent, "crowds", supplied.Crowds != "", func() { intent.Preferences.Crowds = supplied.Crowds })
	assign(intent, "pace", supplied.Pace != "", func() { intent.Preferences.Pace = supplied.Pace })
	assign(intent, "flight_tolerance", supplied.FlightTolerance != "", func() { intent.Preferences.FlightTolerance = supplied.FlightTolerance })
	assign(intent, "paperwork", supplied.Paperwork != "", func() { intent.Preferences.Paperwork = supplied.Paperwork })
	assign(intent, "flight_tolerance", supplied.MaxFlightMinutes != nil, func() { intent.Preferences.MaxFlightMinutes = *supplied.MaxFlightMinutes })
	assign(intent, "budget", supplied.Budget != nil, func() { intent.Budget = *supplied.Budget })
	assign(intent, "departure_airport", supplied.DepartureAirport != "", func() { intent.Origin.DepartureAirport = supplied.DepartureAirport })
	assign(intent, "residence", supplied.ResidenceCountry != "", func() { intent.Traveler.ResidenceCountry = supplied.ResidenceCountry })
	assign(intent, "passports", len(supplied.PassportCountries) > 0, func() { intent.Traveler.PassportCountries = supplied.PassportCountries })
	if len(supplied.ResidencePermits) > 0 {
		intent.Traveler.ResidencePermits = supplied.ResidencePermits
	}
}

func applyInferred(intent *discovery.Intent, inferred Inferred) {
	fill(intent, "interests", len(intent.Preferences.Interests) == 0 && len(inferred.Interests) > 0, func() { intent.Preferences.Interests = inferred.Interests })
	fill(intent, "weather", isUnset(intent.Preferences.Weather) && inferred.Weather != "", func() { intent.Preferences.Weather = inferred.Weather })
	fill(intent, "crowds", isUnset(intent.Preferences.Crowds) && inferred.Crowds != "", func() { intent.Preferences.Crowds = inferred.Crowds })
	fill(intent, "pace", isUnset(intent.Preferences.Pace) && inferred.Pace != "", func() { intent.Preferences.Pace = inferred.Pace })
	fill(intent, "flight_tolerance", isUnset(intent.Preferences.FlightTolerance) && inferred.FlightTolerance != "", func() { intent.Preferences.FlightTolerance = inferred.FlightTolerance })
	fill(intent, "paperwork", isUnset(intent.Preferences.Paperwork) && inferred.Paperwork != "", func() { intent.Preferences.Paperwork = inferred.Paperwork })
}

func assign(intent *discovery.Intent, field string, present bool, apply func()) {
	if !present {
		return
	}
	apply()
	intent.Sources[field] = discovery.SourceRequest
}

func fill(intent *discovery.Intent, field string, permitted bool, apply func()) {
	if !permitted {
		return
	}
	apply()
	intent.Sources[field] = discovery.SourceInferred
}

func isUnset(value string) bool { return value == "" || value == "any" }
