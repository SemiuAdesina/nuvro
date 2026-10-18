package redis

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

const personalizationCapability = "personalization"

type PersonalizationCache struct{ cache *Cache }

func NewPersonalizationCache(cache *Cache) *PersonalizationCache {
	return &PersonalizationCache{cache: cache}
}

func (p *PersonalizationCache) Context(ctx context.Context, userID string) (personalization.Context, bool) {
	var value personalization.Context
	if !p.cache.read(ctx, personalizationCapability, digest(userID), &value) || value.UserID == "" {
		return personalization.Context{}, false
	}
	return value, true
}

func (p *PersonalizationCache) StoreContext(ctx context.Context, userID string, value personalization.Context, ttl time.Duration) {
	p.cache.write(ctx, personalizationCapability, digest(userID), value, ttl)
}

func (p *PersonalizationCache) Invalidate(ctx context.Context, userID string) {
	p.cache.delete(ctx, personalizationCapability, digest(userID))
}
