package personalization

import (
	"slices"
	"time"
)

type Origin string

const (
	ExplicitProfile  Origin = "explicit_profile"
	ConfirmedLearned Origin = "confirmed_learned"
	InferredLearned  Origin = "inferred_learned"
)

func (o Origin) Rank() int {
	switch o {
	case ExplicitProfile:
		return 3
	case ConfirmedLearned:
		return 2
	default:
		return 1
	}
}

type Preference struct {
	Key        Key
	Origin     Origin
	Confidence Confidence
	Weight     int
}

type Context struct {
	UserID                 string
	HistoryLearningEnabled bool
	Explicit               []Preference
	Learned                []Preference
	Suppressed             []Key
	GeneratedAt            time.Time
	RulesVersion           string
}

func EmptyContext(userID string, enabled bool, generatedAt time.Time) Context {
	return Context{
		UserID: userID, HistoryLearningEnabled: enabled,
		Explicit: []Preference{}, Learned: []Preference{}, Suppressed: []Key{},
		GeneratedAt: generatedAt.UTC(), RulesVersion: RulesVersion,
	}
}

func (c Context) Influence(key Key) int {
	for _, preference := range c.Explicit {
		if preference.Key == key {
			return preference.Weight
		}
	}
	if !c.HistoryLearningEnabled {
		return 0
	}
	if slices.Contains(c.Suppressed, key) {
		return 0
	}
	for _, preference := range c.Learned {
		if preference.Key == key {
			return preference.Weight
		}
	}
	return 0
}

func (c Context) Holds(key Key) bool { return c.Influence(key) > 0 }

func (c Context) OriginOf(key Key) (Origin, bool) {
	for _, preference := range c.Explicit {
		if preference.Key == key {
			return ExplicitProfile, true
		}
	}
	if !c.HistoryLearningEnabled || slices.Contains(c.Suppressed, key) {
		return "", false
	}
	for _, preference := range c.Learned {
		if preference.Key == key {
			return preference.Origin, true
		}
	}
	return "", false
}

const ExplicitWeight = 100
const ConfirmedWeight = 60
const InferredHighWeight = 35
const InferredMediumWeight = 20
const InferredLowWeight = 8

func WeightFor(origin Origin, confidence Confidence) int {
	if origin == ExplicitProfile {
		return ExplicitWeight
	}
	if origin == ConfirmedLearned {
		return ConfirmedWeight
	}
	switch confidence {
	case High:
		return InferredHighWeight
	case MediumConfidence:
		return InferredMediumWeight
	default:
		return InferredLowWeight
	}
}
