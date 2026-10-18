package httpclient

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"
)

type RetryPolicy struct {
	Attempts   int
	BaseDelay  time.Duration
	MaximumGap time.Duration
}

func DefaultRetry() RetryPolicy {
	return RetryPolicy{Attempts: 3, BaseDelay: 100 * time.Millisecond, MaximumGap: 2 * time.Second}
}

func (p RetryPolicy) attempts() int {
	if p.Attempts < 1 {
		return 1
	}
	return p.Attempts
}

func (p RetryPolicy) gapBefore(attempt int, fraction float64) time.Duration {
	if p.BaseDelay <= 0 {
		return 0
	}
	scaled := float64(p.BaseDelay) * math.Pow(2, float64(attempt-1))
	if p.MaximumGap > 0 && scaled > float64(p.MaximumGap) {
		scaled = float64(p.MaximumGap)
	}
	return time.Duration(scaled * (0.5 + fraction/2))
}

func retryable(err error, status int) bool {
	if errors.Is(err, ErrMalformed) {
		return false
	}
	switch status {
	case 0, http.StatusTooManyRequests, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return err != nil
	default:
		return false
	}
}

func wait(ctx context.Context, timer func(time.Duration) <-chan time.Time, gap time.Duration) error {
	if gap <= 0 {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer(gap):
		return nil
	}
}
