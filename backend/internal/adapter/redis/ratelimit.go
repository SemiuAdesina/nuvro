package redis

import (
	"context"
	"errors"
	"time"

	redisclient "github.com/redis/go-redis/v9"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const rateLimitCapability = "rate_limit"
const milliTokens = 1000

var ErrNoAdmissionBackend = errors.New("redis: rate limiting requires a connection")

var bucket = redisclient.NewScript(`
local tokens = tonumber(redis.call('HGET', KEYS[1], 'tokens'))
local updated = tonumber(redis.call('HGET', KEYS[1], 'updated'))
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
if tokens == nil or updated == nil or updated > now then
  tokens = capacity
  updated = now
end
tokens = math.min(capacity, tokens + (now - updated) * rate)
local allowed = 0
if tokens >= 1000 then
  tokens = tokens - 1000
  allowed = 1
end
redis.call('HSET', KEYS[1], 'tokens', tokens, 'updated', now)
redis.call('PEXPIRE', KEYS[1], ARGV[4])
return allowed
`)

type RateLimiter struct {
	client     *redisclient.Client
	allowances map[admission.Class]admission.Allowance
	timer      clock.Clock
}

func NewRateLimiter(client *Client, allowances map[admission.Class]admission.Allowance, timer clock.Clock) (*RateLimiter, error) {
	if client == nil {
		return nil, ErrNoAdmissionBackend
	}
	for class, allowance := range allowances {
		if !class.Valid() || !allowance.Valid() {
			return nil, errors.New("redis: rate allowance for " + string(class) + " is not usable")
		}
	}
	return &RateLimiter{client: client.client, allowances: allowances, timer: timer}, nil
}

func (r *RateLimiter) Allow(ctx context.Context, class admission.Class, subject string) (bool, error) {
	allowance, known := r.allowances[class]
	if !known || subject == "" {
		return true, nil
	}
	key, err := Key(rateLimitCapability, string(class)+"."+digest(subject))
	if err != nil {
		return false, err
	}
	capacity := float64(allowance.Burst * milliTokens)
	allowed, err := bucket.Run(ctx, r.client, []string{key},
		capacity, allowance.PerSecond, r.timer.Now().UTC().UnixMilli(), lifetimeOf(allowance)).Int64()
	if err != nil {
		return false, err
	}
	return allowed == 1, nil
}

func lifetimeOf(allowance admission.Allowance) int64 {
	refill := time.Duration(float64(allowance.Burst) / allowance.PerSecond * float64(time.Second))
	return (refill + time.Minute).Milliseconds()
}
