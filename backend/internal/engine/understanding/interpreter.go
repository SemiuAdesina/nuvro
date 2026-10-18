package understanding

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
)

type Interpreter interface {
	Interpret(context.Context, string) (Inferred, error)
}

const maximumQueryLength = 500

func (i Inferred) Validate() error {
	preferences := discovery.Preferences{
		Interests: i.Interests, Weather: i.Weather, Crowds: i.Crowds,
		Pace: i.Pace, FlightTolerance: i.FlightTolerance, Paperwork: i.Paperwork,
	}
	return preferences.Validate()
}

func (i Inferred) normalized() Inferred {
	i.Interests = discovery.NormalizeTags(i.Interests)
	return i
}
