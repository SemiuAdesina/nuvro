package travel

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
	Role(context.Context, string, string) (trip.Role, error)
}

type Catalog interface {
	Destination(context.Context, string) (destination.Destination, error)
}

type Itineraries interface {
	Itinerary(context.Context, string) (itinerary.Itinerary, error)
	ReplaceItineraryItems(context.Context, string, int, []itinerary.Item) (itinerary.Itinerary, error)
}

type Alerts interface {
	Alerts(context.Context, string) ([]alert.Alert, error)
}

type Proposals interface {
	SaveProposal(context.Context, replan.Proposal) error
	Proposal(context.Context, string) (replan.Proposal, error)
	MarkProposalApplied(ctx context.Context, proposalID, userID string, at time.Time) error
}

type Forecaster interface {
	For(ctx context.Context, location geo.Point, start, end time.Time, language string) weather.Window
}

type RouteFinder interface {
	Compute(ctx context.Context, origin, destination geo.Point, mode route.Mode, departAt time.Time) route.Evidence
}

type Planner interface {
	Propose(context.Context, replanning.Request) replan.Proposal
}

type Preferences interface {
	Preferences(context.Context, string) (user.Preferences, error)
	Profile(context.Context, string) (user.Profile, error)
}
