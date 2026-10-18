package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	domainlanguage "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	languageengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	alertsusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/alerts"
	languageusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/language"
	travelusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

func (s *store) SaveAlert(_ context.Context, value alert.Alert) (alert.Alert, error) {
	if s.alerts == nil {
		s.alerts = map[string]alert.Alert{}
	}
	s.alerts[value.ID] = value
	return value, nil
}

func (s *store) Alerts(_ context.Context, tripID string) ([]alert.Alert, error) {
	values := make([]alert.Alert, 0, len(s.alerts))
	for _, value := range s.alerts {
		if value.TripID == tripID {
			values = append(values, value)
		}
	}
	return values, nil
}

func (s *store) Alert(_ context.Context, alertID string) (alert.Alert, error) {
	value, found := s.alerts[alertID]
	if !found {
		return alert.Alert{}, errors.New("alert not found")
	}
	return value, nil
}

func (s *store) UpdateAlertStatus(_ context.Context, alertID string, status alert.Status) error {
	value, found := s.alerts[alertID]
	if !found {
		return errors.New("alert not found")
	}
	value.Status = status
	s.alerts[alertID] = value
	return nil
}

func (s *store) ExpireAlerts(context.Context, string, time.Time) error { return nil }

func (s *store) SaveProposal(_ context.Context, proposal replan.Proposal) error {
	if s.proposals == nil {
		s.proposals = map[string]replan.Proposal{}
	}
	s.proposals[proposal.ID] = proposal
	return nil
}

func (s *store) Proposal(_ context.Context, proposalID string) (replan.Proposal, error) {
	proposal, found := s.proposals[proposalID]
	if !found {
		return replan.Proposal{}, errors.New("proposal not found")
	}
	return proposal, nil
}

func (s *store) MarkProposalApplied(_ context.Context, proposalID, _ string, at time.Time) error {
	proposal, found := s.proposals[proposalID]
	if !found {
		return errors.New("proposal not found")
	}
	applied := at
	proposal.AppliedAt = &applied
	s.proposals[proposalID] = proposal
	return nil
}

func (s *store) ReplaceItineraryItems(_ context.Context, _ string, expected int, items []itinerary.Item) (itinerary.Itinerary, error) {
	s.itinerary.Items = items
	s.itinerary.Version = expected + 1
	return s.itinerary, nil
}

type runtimeFinder struct{}

func (runtimeFinder) Nearby(context.Context, geo.Point, []place.Category, string, int) ([]place.Place, error) {
	return []place.Place{{
		ID: "museum", Name: "Madeira Story Centre", Category: place.Museum,
		Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048}, Status: place.Open,
		Rating: 4.4, ReviewCount: 900, TypicalMinutes: 90,
	}}, nil
}

type runtimeRouter struct{}

func (runtimeRouter) Compute(_ context.Context, origin, destination geo.Point, mode route.Mode, _ time.Time) route.Evidence {
	return route.Evidence{
		Status: route.Found, Origin: origin, Destination: destination, Mode: mode,
		DurationSeconds: 720, DistanceMeters: 900, Confidence: route.High,
	}
}

type runtimeWeather struct{}

func (runtimeWeather) For(context.Context, geo.Point, time.Time, time.Time, string) weather.Window {
	return weather.Window{Availability: weather.Available, Current: &weather.Observation{
		At: contractNow, Condition: weather.Clear, TemperatureCelsius: 23,
	}}
}

func (runtimeWeather) Impact(_ weather.Window, _ time.Time, sensitivity weather.Sensitivity) weather.Impact {
	if sensitivity == weather.Outdoor {
		return weather.HighImpact
	}
	return weather.NoImpact
}

type runtimeTranslator struct{}

func (runtimeTranslator) Translate(_ context.Context, query languageengine.Query) (languageengine.Result, error) {
	return languageengine.Result{
		Source: domainlanguage.Code("en"), Translated: "[" + query.Target.String() + "] " + query.Text,
		Confidence: "high", Provider: "fixture",
	}, nil
}

func (runtimeTranslator) Supports(domainlanguage.Code) bool { return true }

func travelService(t *testing.T, repository *store, timer clock.Clock) *travelusecase.Service {
	t.Helper()
	return travelusecase.NewService(travelusecase.Dependencies{
		Trips: tripAccess{store: repository}, Catalog: repository, Itineraries: repository,
		Alerts: repository, Proposals: repository, Forecaster: runtimeWeather{}, Routes: runtimeRouter{},
		Planner:     replanning.New(runtimeFinder{}, runtimeRouter{}, runtimeWeather{}, nil, timer),
		Preferences: repository, Identifiers: id.NewUUID(),
	}, timer)
}

func alertService(t *testing.T, repository *store, timer clock.Clock) *alertsusecase.Service {
	t.Helper()
	return alertsusecase.NewService(tripAccess{store: repository}, repository, id.NewUUID(), timer)
}

func languageService(t *testing.T, repository *store) *languageusecase.Service {
	t.Helper()
	return languageusecase.NewService(tripAccess{store: repository}, repository, repository,
		languageengine.New(runtimeTranslator{}))
}

type phraseSource struct{}

func (p phraseSource) Phrases(context.Context, string) ([]domainlanguage.Phrase, error) {
	return domainlanguage.EssentialPhrases(domainlanguage.Emergency), nil
}
