package personalization

import (
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

type Thresholds struct {
	MinimumEvidence   int
	MediumScore       int
	HighScore         int
	SuggestionScore   int
	SuppressionWindow time.Duration
}

func DefaultThresholds() Thresholds {
	return Thresholds{
		MinimumEvidence: 2, MediumScore: 8, HighScore: 18,
		SuggestionScore: 12, SuppressionWindow: 180 * 24 * time.Hour,
	}
}

type DecayClass string

const (
	SlowDecay DecayClass = "slow"
	FastDecay DecayClass = "fast"
)

type Decay struct {
	SlowHalfLife time.Duration
	FastHalfLife time.Duration
}

func DefaultDecay() Decay {
	return Decay{SlowHalfLife: 365 * 24 * time.Hour, FastHalfLife: 60 * 24 * time.Hour}
}

func ClassOf(key domain.Key) DecayClass {
	switch key {
	case domain.PaceRelaxed, domain.PacePacked, domain.TransferTolerance:
		return FastDecay
	default:
		return SlowDecay
	}
}

func (d Decay) HalfLife(class DecayClass) time.Duration {
	if class == FastDecay {
		return d.FastHalfLife
	}
	return d.SlowHalfLife
}

func (d Decay) Factor(class DecayClass, age time.Duration) int {
	if age <= 0 {
		return 100
	}
	halfLife := d.HalfLife(class)
	if halfLife <= 0 {
		return 100
	}
	factor := 100
	for elapsed := halfLife; elapsed <= age; elapsed += halfLife {
		factor /= 2
		if factor == 0 {
			return 0
		}
	}
	return factor
}

type Config struct {
	Thresholds Thresholds
	Decay      Decay
	Version    string
}

func DefaultConfig() Config {
	return Config{Thresholds: DefaultThresholds(), Decay: DefaultDecay(), Version: domain.RulesVersion}
}

func (c Config) Validate() error {
	if c.Version == "" || c.Thresholds.MinimumEvidence < 1 {
		return domain.ErrInvalid
	}
	if c.Thresholds.MediumScore < 1 || c.Thresholds.HighScore <= c.Thresholds.MediumScore {
		return domain.ErrInvalid
	}
	if c.Decay.SlowHalfLife <= 0 || c.Decay.FastHalfLife <= 0 {
		return domain.ErrInvalid
	}
	return nil
}
