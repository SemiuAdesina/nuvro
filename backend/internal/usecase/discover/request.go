package discover

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
)

type Request struct {
	Overrides understanding.Supplied
	Party     discovery.Party
	Window    discovery.Window
	Query     string
	Limit     int
}

type Response struct {
	Intent          discovery.Intent
	Result          Ranked
	MissingContext  []string
	GeneratedAt     time.Time
	EvidenceVersion string
}

func profileIntent(profile user.Profile, preferences user.Preferences) discovery.Intent {
	intent := discovery.Intent{
		Preferences: discovery.Preferences{
			Interests: preferences.Interests,
			Weather:   preferences.Weather,
			Crowds:    preferences.Crowds,
			Pace:      preferences.Pace,
			Paperwork: paperworkFrom(preferences.EntrySimplicity),
		},
	}
	if preferences.MaxFlightMinutes != nil {
		intent.Preferences.MaxFlightMinutes = int(*preferences.MaxFlightMinutes)
	}
	if profile.HomeCountry != nil {
		intent.Traveler.ResidenceCountry = *profile.HomeCountry
	}
	if profile.DepartureAirport != nil {
		intent.Origin.DepartureAirport = *profile.DepartureAirport
	}
	return intent
}

func paperworkFrom(entrySimplicity string) string {
	if entrySimplicity == "simple" {
		return "keep_it_simple"
	}
	return "any"
}
