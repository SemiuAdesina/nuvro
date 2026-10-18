package redis

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
)

const idempotencyCapability = "idempotency"

type IdempotencyStore struct{ cache *Cache }

func NewIdempotencyStore(cache *Cache) *IdempotencyStore {
	if cache == nil {
		return nil
	}
	return &IdempotencyStore{cache: cache}
}

func (s *IdempotencyStore) Recorded(ctx context.Context, subject, key string) (replay.Response, bool) {
	var recorded replay.Response
	if !s.cache.read(ctx, idempotencyCapability, digest(subject+"|"+key), &recorded) || recorded.Status == 0 {
		return replay.Response{}, false
	}
	return recorded, true
}

func (s *IdempotencyStore) Record(ctx context.Context, subject, key string, recorded replay.Response, ttl time.Duration) {
	s.cache.write(ctx, idempotencyCapability, digest(subject+"|"+key), recorded, ttl)
}
