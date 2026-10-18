package app

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	personalizationusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/personalization"
)

func newPersonalizationService(store *postgres.Store, cache *redisadapter.Cache,
	access tripAccess, timer clock.Clock) (*personalizationusecase.Service, error) {
	learning, err := engine.New(engine.DefaultConfig(), timer)
	if err != nil {
		return nil, err
	}
	return personalizationusecase.NewService(personalizationusecase.Dependencies{
		Trips: access, Preferences: store, Store: store, Feedback: store,
		Cache: personalizationCache(cache), Engine: learning, Identifiers: id.NewUUID(),
	}, timer), nil
}

func personalizationCache(cache *redisadapter.Cache) personalizationusecase.ContextCache {
	if cache == nil {
		return nil
	}
	return redisadapter.NewPersonalizationCache(cache)
}
