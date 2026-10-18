package matching

import "errors"

const Version = "match-v1"

var ErrInvalidConfig = errors.New("matching: weights must total one hundred points")

type ComponentKey string

const (
	Interests ComponentKey = "interests"
	Weather   ComponentKey = "weather"
	Crowds    ComponentKey = "crowds"
	Pace      ComponentKey = "pace"
	FlightFit ComponentKey = "flight"
	BudgetFit ComponentKey = "budget"
	AccessFit ComponentKey = "access"
)

type Weight struct {
	Key            ComponentKey
	Label          string
	PossiblePoints int
}

type Config struct {
	Version string
	Weights []Weight
}

func DefaultConfig() Config {
	return Config{
		Version: Version,
		Weights: []Weight{
			{Key: Interests, Label: "Interests", PossiblePoints: 25},
			{Key: BudgetFit, Label: "Budget", PossiblePoints: 20},
			{Key: Weather, Label: "Weather", PossiblePoints: 15},
			{Key: AccessFit, Label: "Access", PossiblePoints: 15},
			{Key: Crowds, Label: "Crowds", PossiblePoints: 10},
			{Key: FlightFit, Label: "Flight", PossiblePoints: 10},
			{Key: Pace, Label: "Pace", PossiblePoints: 5},
		},
	}
}

func (c Config) Validate() error {
	total := 0
	seen := make(map[ComponentKey]struct{}, len(c.Weights))
	for _, weight := range c.Weights {
		if weight.PossiblePoints < 0 || weight.Key == "" || weight.Label == "" {
			return ErrInvalidConfig
		}
		if _, duplicate := seen[weight.Key]; duplicate {
			return ErrInvalidConfig
		}
		seen[weight.Key] = struct{}{}
		total += weight.PossiblePoints
	}
	if total != 100 || c.Version == "" {
		return ErrInvalidConfig
	}
	return nil
}
