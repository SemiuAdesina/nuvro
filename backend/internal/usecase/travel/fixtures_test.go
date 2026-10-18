package travel_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"

var tripStart = time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

func at(dayOffset, hour, minute int) time.Time {
	return tripStart.AddDate(0, 0, dayOffset).Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

type trips struct {
	record trip.Trip
	role   trip.Role
	err    error
}

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if t.err != nil {
		return trip.Trip{}, t.err
	}
	if principal.UserID == "" {
		return trip.Trip{}, errors.New("unauthenticated")
	}
	return t.record, nil
}

func (t trips) Role(context.Context, string, string) (trip.Role, error) { return t.role, t.err }

type catalog struct{ timezone string }

func (c catalog) Destination(context.Context, string) (destination.Destination, error) {
	timezone := c.timezone
	if timezone == "" {
		timezone = "Atlantic/Madeira"
	}
	return destination.Destination{
		ID: "madeira-pt", Name: "Madeira", CountryCode: "PT", Timezone: timezone,
		Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
	}, nil
}

type itineraries struct {
	plan     itinerary.Itinerary
	present  bool
	replaced []itinerary.Item
	version  int
	err      error
}

func (i *itineraries) Itinerary(context.Context, string) (itinerary.Itinerary, error) {
	if !i.present {
		return itinerary.Itinerary{}, errors.New("no itinerary")
	}
	return i.plan, nil
}

func (i *itineraries) ReplaceItineraryItems(_ context.Context, _ string, expected int, items []itinerary.Item) (itinerary.Itinerary, error) {
	if i.err != nil {
		return itinerary.Itinerary{}, i.err
	}
	i.version = expected + 1
	i.replaced = items
	i.plan.Items = items
	i.plan.Version = i.version
	return i.plan, nil
}

type alerts struct {
	stored []alert.Alert
	err    error
}

func (a alerts) Alerts(context.Context, string) ([]alert.Alert, error) { return a.stored, a.err }

type proposals struct {
	saved    map[string]replan.Proposal
	applied  map[string]time.Time
	saveErr  error
	applyErr error
}

func newProposals() *proposals {
	return &proposals{saved: map[string]replan.Proposal{}, applied: map[string]time.Time{}}
}

func (p *proposals) SaveProposal(_ context.Context, proposal replan.Proposal) error {
	if p.saveErr != nil {
		return p.saveErr
	}
	p.saved[proposal.ID] = proposal
	return nil
}

func (p *proposals) Proposal(_ context.Context, proposalID string) (replan.Proposal, error) {
	proposal, found := p.saved[proposalID]
	if !found {
		return replan.Proposal{}, errors.New("not found")
	}
	if applied, ok := p.applied[proposalID]; ok {
		proposal.AppliedAt = &applied
	}
	return proposal, nil
}

func (p *proposals) MarkProposalApplied(_ context.Context, proposalID, _ string, at time.Time) error {
	if p.applyErr != nil {
		return p.applyErr
	}
	p.applied[proposalID] = at
	return nil
}

type forecaster struct{ window weather.Window }

func (f forecaster) For(context.Context, geo.Point, time.Time, time.Time, string) weather.Window {
	return f.window
}

type router struct{ seconds int }

func (r router) Compute(_ context.Context, origin, destination geo.Point, mode route.Mode, _ time.Time) route.Evidence {
	if r.seconds < 0 {
		return route.UnavailableEvidence(origin, destination, mode)
	}
	return route.Evidence{
		Status: route.Found, Origin: origin, Destination: destination, Mode: mode,
		DurationSeconds: r.seconds, DistanceMeters: r.seconds, Confidence: route.High,
	}
}

type finder struct{ places []place.Place }

func (f finder) Nearby(context.Context, geo.Point, []place.Category, string, int) ([]place.Place, error) {
	return f.places, nil
}

type preferences struct {
	profile user.Profile
	values  user.Preferences
}

func (p preferences) Profile(context.Context, string) (user.Profile, error) { return p.profile, nil }
func (p preferences) Preferences(context.Context, string) (user.Preferences, error) {
	return p.values, nil
}

func item(id string, hour int, category place.Category) itinerary.Item {
	return itinerary.Item{
		ID: id, TripID: tripID, Day: 3,
		StartAt: at(2, hour, 0), EndAt: at(2, hour+2, 0),
		Kind: itinerary.Activity, Title: id, PlaceID: "place-" + id, Category: category,
		Location: funchal, Status: itinerary.Planned, Source: itinerary.Generated,
	}
}

func madeiraTrip() trip.Trip {
	start := tripStart
	end := tripStart.AddDate(0, 0, 6)
	destinationID := "madeira-pt"
	return trip.Trip{
		ID: tripID, OwnerID: owner, DestinationID: &destinationID, StartDate: &start, EndDate: &end,
		Adults: 2, Currency: "EUR", Status: trip.Active,
	}
}

func candidate(id string, category place.Category, minutes int) place.Place {
	return place.Place{
		ID: id, Name: id, Category: category, Location: funchal, Status: place.Open,
		Rating: 4.5, ReviewCount: 200, TypicalMinutes: minutes,
	}
}

func newService(t *testing.T, now time.Time, dependencies usecase.Dependencies) *usecase.Service {
	t.Helper()
	if dependencies.Identifiers == nil {
		dependencies.Identifiers = id.NewUUID()
	}
	return usecase.NewService(dependencies, clock.NewFixed(now))
}

func baseDependencies(store *itineraries, proposalStore *proposals, now time.Time) usecase.Dependencies {
	return usecase.Dependencies{
		Trips: trips{record: madeiraTrip(), role: trip.Owner}, Catalog: catalog{},
		Itineraries: store, Alerts: alerts{}, Proposals: proposalStore,
		Forecaster: forecaster{window: weather.Window{Availability: weather.Available}},
		Routes:     router{seconds: 720},
		Planner: replanning.New(finder{places: []place.Place{candidate("museum", place.Museum, 90)}},
			router{seconds: 600}, rainOutdoors{}, nil, clock.NewFixed(now)),
		Preferences: preferences{profile: user.DefaultProfile(owner), values: user.DefaultPreferences()},
	}
}

type rainOutdoors struct{}

func (rainOutdoors) Impact(_ weather.Window, _ time.Time, sensitivity weather.Sensitivity) weather.Impact {
	if sensitivity == weather.Outdoor {
		return weather.HighImpact
	}
	return weather.NoImpact
}

func principal() user.Principal { return user.Principal{UserID: owner} }
