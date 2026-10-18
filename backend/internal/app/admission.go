package app

import (
	"errors"

	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
)

var errAdmissionUnavailable = errors.New("app: rate limiting is enabled but Redis is not configured")
var errReplayUnavailable = errors.New("app: idempotent replay requires Redis outside development")

func newLimiter(cfg config.Config, client *redisadapter.Client, timer clock.Clock) (admission.Limiter, error) {
	if !cfg.Admission.Enabled {
		return nil, nil
	}
	if client == nil {
		if cfg.RequiresBackingServices() {
			return nil, errAdmissionUnavailable
		}
		return nil, nil
	}
	return redisadapter.NewRateLimiter(client, cfg.Admission.Allowances, timer)
}

func newReplayStore(cfg config.Config, cache *redisadapter.Cache) (replay.Store, error) {
	store := redisadapter.NewIdempotencyStore(cache)
	if store == nil {
		if cfg.RequiresBackingServices() {
			return nil, errReplayUnavailable
		}
		return nil, nil
	}
	return store, nil
}

type gates struct {
	limiter admission.Limiter
	replays replay.Store
}

func newGates(cfg config.Config, client *redisadapter.Client, timer clock.Clock) (gates, error) {
	limiter, err := newLimiter(cfg, client, timer)
	if err != nil {
		return gates{}, err
	}
	replays, err := newReplayStore(cfg, redisadapter.NewCache(client))
	if err != nil {
		return gates{}, err
	}
	return gates{limiter: limiter, replays: replays}, nil
}
