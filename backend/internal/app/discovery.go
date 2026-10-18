package app

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/matching"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

const recommendationLimit = 10

func newDiscoveryEngines(cfg config.Config, timer clock.Clock) (discover.Engines, error) {
	scoring, err := matching.New(matching.DefaultConfig())
	if err != nil {
		return discover.Engines{}, err
	}
	return discover.Engines{
		Understanding: understanding.New(nil),
		Feasibility:   feasibility.New(discoveryProviders(cfg), timer),
		Matching:      scoring,
		Ranking:       recommendation.New(nil, recommendationLimit),
	}, nil
}

func discoveryProviders(config.Config) feasibility.Providers {
	return feasibility.Providers{}
}
