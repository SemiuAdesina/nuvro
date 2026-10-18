package app

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	commerceusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

func newCommerceService(cfg config.Config, store *postgres.Store, cache *redisadapter.Cache,
	access tripAccess, timer clock.Clock) *commerceusecase.Service {
	registry := newPartnerRegistry(cfg)
	commerceEngine := engine.NewEngine(registry, offerCache(cache), nil, id.NewUUID(), timer)
	return commerceusecase.NewService(commerceusecase.Dependencies{
		Trips: access, Catalog: store, Itineraries: store, Handoffs: store, Conversions: store,
		Watches: store, Entitlements: pendingEntitlements{}, Engine: commerceEngine,
		Providers: commerceusecase.Providers{}, Identifiers: id.NewUUID(),
	}, timer)
}

func offerCache(cache *redisadapter.Cache) engine.OfferStore {
	if cache == nil {
		return nil
	}
	return redisadapter.NewOfferCache(cache)
}

type partnerRegistry struct{ partners map[string]domain.Partner }

func newPartnerRegistry(config.Config) *partnerRegistry {
	return &partnerRegistry{partners: map[string]domain.Partner{}}
}

func (p *partnerRegistry) Partner(code string) (domain.Partner, bool) {
	partner, found := p.partners[code]
	return partner, found
}

type pendingEntitlements struct{}

func (pendingEntitlements) AllowsPriceMonitoring(context.Context, user.Principal) (bool, error) {
	return false, nil
}

type watchScheduler struct{}

func (watchScheduler) NextCheck(now, departure time.Time) time.Time {
	return commerceusecase.NextCheck(now, departure)
}
