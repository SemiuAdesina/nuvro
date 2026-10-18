package weather

import "github.com/SemiuAdesina/nuvro/backend/internal/domain/place"

type Impact string

const (
	NoImpact       Impact = "none"
	LowImpact      Impact = "low"
	ModerateImpact Impact = "moderate"
	HighImpact     Impact = "high"
	UnknownImpact  Impact = "unknown"
)

type Sensitivity string

const (
	Outdoor   Sensitivity = "outdoor"
	Sheltered Sensitivity = "sheltered"
	Indoor    Sensitivity = "indoor"
)

const highHeatCelsius = 35
const severeHeatCelsius = 40
const freezingCelsius = -5
const strongWindKilometresPerHour = 55

func SensitivityOf(category place.Category) Sensitivity {
	switch {
	case category.Outdoor():
		return Outdoor
	case category.Indoor():
		return Indoor
	default:
		return Sheltered
	}
}

func Assess(observation Observation, sensitivity Sensitivity) Impact {
	if !observation.Condition.Valid() || observation.Condition == UnknownCondition {
		return UnknownImpact
	}
	severity := conditionSeverity(observation)
	switch sensitivity {
	case Outdoor:
		return outdoorImpact(severity)
	case Sheltered:
		return shelteredImpact(severity)
	default:
		return indoorImpact(severity)
	}
}

func conditionSeverity(observation Observation) int {
	severity := 0
	switch observation.Condition {
	case Storm:
		severity = 3
	case HeavyRain, Snow:
		severity = 2
	case Rain, Fog, Wind:
		severity = 1
	}
	if observation.TemperatureCelsius >= severeHeatCelsius || observation.TemperatureCelsius <= freezingCelsius {
		severity = max(severity, 2)
	} else if observation.TemperatureCelsius >= highHeatCelsius {
		severity = max(severity, 1)
	}
	if observation.WindKilometresPerHour >= strongWindKilometresPerHour {
		severity = max(severity, 2)
	}
	return severity
}

func outdoorImpact(severity int) Impact {
	switch severity {
	case 0:
		return NoImpact
	case 1:
		return ModerateImpact
	case 2:
		return HighImpact
	default:
		return HighImpact
	}
}

func shelteredImpact(severity int) Impact {
	switch severity {
	case 0:
		return NoImpact
	case 1:
		return LowImpact
	case 2:
		return ModerateImpact
	default:
		return HighImpact
	}
}

func indoorImpact(severity int) Impact {
	if severity >= 3 {
		return ModerateImpact
	}
	if severity >= 2 {
		return LowImpact
	}
	return NoImpact
}

func (i Impact) Blocks() bool { return i == HighImpact }
