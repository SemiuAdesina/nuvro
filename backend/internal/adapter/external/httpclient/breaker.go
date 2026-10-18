package httpclient

import (
	"sync"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type BreakerPolicy struct {
	FailureThreshold int
	Cooldown         time.Duration
}

func DefaultBreaker() BreakerPolicy {
	return BreakerPolicy{FailureThreshold: 5, Cooldown: 30 * time.Second}
}

func (p BreakerPolicy) enabled() bool { return p.FailureThreshold > 0 && p.Cooldown > 0 }

type breaker struct {
	policy   BreakerPolicy
	timer    clock.Clock
	mutex    sync.Mutex
	failures int
	openedAt time.Time
	probing  bool
}

func newBreaker(policy BreakerPolicy, timer clock.Clock) *breaker {
	return &breaker{policy: policy, timer: timer}
}

func (b *breaker) allow() bool {
	if !b.policy.enabled() {
		return true
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.openedAt.IsZero() {
		return true
	}
	if b.timer.Now().Sub(b.openedAt) < b.policy.Cooldown {
		return false
	}
	if b.probing {
		return false
	}
	b.probing = true
	return true
}

func (b *breaker) succeeded() {
	if !b.policy.enabled() {
		return
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.failures = 0
	b.openedAt = time.Time{}
	b.probing = false
}

func (b *breaker) failed() {
	if !b.policy.enabled() {
		return
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.failures++
	b.probing = false
	if b.failures >= b.policy.FailureThreshold {
		b.openedAt = b.timer.Now()
	}
}
