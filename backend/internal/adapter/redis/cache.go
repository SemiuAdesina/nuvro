package redis

import (
	"context"
	"encoding/json"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type Cache struct{ client *redisclient.Client }

func NewCache(client *Client) *Cache {
	if client == nil {
		return nil
	}
	return &Cache{client: client.client}
}

func (c *Cache) read(ctx context.Context, capability, identifier string, into any) bool {
	if c == nil {
		return false
	}
	key, err := Key(capability, identifier)
	if err != nil {
		return false
	}
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, into) == nil
}

func (c *Cache) write(ctx context.Context, capability, identifier string, value any, ttl time.Duration) {
	if c == nil || ttl <= 0 {
		return
	}
	key, err := Key(capability, identifier)
	if err != nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = c.client.Set(ctx, key, raw, ttl).Err()
}

func (c *Cache) delete(ctx context.Context, capability, identifier string) {
	if c == nil {
		return
	}
	key, err := Key(capability, identifier)
	if err != nil {
		return
	}
	_ = c.client.Del(ctx, key).Err()
}
