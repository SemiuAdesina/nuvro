package redis

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

const offerCapability = "offer"

type OfferCache struct{ cache *Cache }

func NewOfferCache(cache *Cache) *OfferCache { return &OfferCache{cache: cache} }

func (o *OfferCache) SaveOffers(ctx context.Context, offers []commerce.StoredOffer, ttl time.Duration) error {
	for _, offer := range offers {
		o.cache.write(ctx, offerCapability, digest(offer.Offer.ID), offer, ttl)
	}
	return nil
}

func (o *OfferCache) Offer(ctx context.Context, offerID string) (commerce.StoredOffer, bool) {
	var stored commerce.StoredOffer
	if !o.cache.read(ctx, offerCapability, digest(offerID), &stored) || stored.Offer.ID == "" {
		return commerce.StoredOffer{}, false
	}
	return stored, true
}
