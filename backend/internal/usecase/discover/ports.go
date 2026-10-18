package discover

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
)

type Personalization interface {
	Context(context.Context, user.Principal) (personalization.Context, error)
}

type ProfileReader interface {
	Profile(context.Context, string) (user.Profile, error)
	Preferences(context.Context, string) (user.Preferences, error)
}

type DestinationCatalog interface {
	Destinations(context.Context, int32) ([]destination.Destination, error)
}

type Understanding interface {
	Understand(context.Context, understanding.Input) (discovery.Intent, error)
}

type Feasibility interface {
	Evaluate(context.Context, discovery.Intent, destination.Destination) feasibility.Result
}

type Matching interface {
	Score(discovery.Intent, destination.Destination, feasibility.Result, string) (destination.MatchScore, error)
	Version() string
}

type Ranking interface {
	Rank(context.Context, []recommendation.Candidate) recommendation.Result
}
