package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var ErrRateLimited = errors.New("provider: rate limited")
var ErrUnavailable = errors.New("provider: unavailable")
var ErrMalformed = errors.New("provider: malformed response")
var ErrCircuitOpen = errors.New("provider: circuit open")

const MaximumResponseBytes = 4 << 20

type Recorder interface {
	Provider(ctx context.Context, provider string, failed bool, elapsed time.Duration)
}

type Options struct {
	Provider string
	BaseURL  string
	Timeout  time.Duration
	Retry    RetryPolicy
	Breaker  BreakerPolicy
	Metrics  Recorder
	Clock    clock.Clock
}

type Client struct {
	http     *http.Client
	provider string
	baseURL  string
	retry    RetryPolicy
	breaker  *breaker
	metrics  Recorder
	fraction func() float64
	timer    func(time.Duration) <-chan time.Time
}

func New(options Options) *Client {
	if options.Timeout <= 0 {
		options.Timeout = 5 * time.Second
	}
	if options.Clock == nil {
		options.Clock = clock.NewSystem()
	}
	return &Client{
		http:     &http.Client{Timeout: options.Timeout, Transport: transport()},
		provider: options.Provider,
		baseURL:  options.BaseURL,
		retry:    options.Retry,
		breaker:  newBreaker(options.Breaker, options.Clock),
		metrics:  options.Metrics,
		fraction: rand.Float64,
		timer:    time.After,
	}
}

func transport() http.RoundTripper {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultTransport
	}
	cloned := base.Clone()
	cloned.MaxIdleConnsPerHost = 8
	cloned.ResponseHeaderTimeout = 5 * time.Second
	return cloned
}

type Request struct {
	Path      string
	Body      any
	Headers   map[string]string
	Retryable bool
}

func (c *Client) Post(ctx context.Context, request Request, into any) error {
	payload, err := json.Marshal(request.Body)
	if err != nil {
		return ErrMalformed
	}
	return c.call(ctx, request, into, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+request.Path, bytes.NewReader(payload))
	})
}

func (c *Client) Get(ctx context.Context, request Request, into any) error {
	request.Retryable = true
	return c.call(ctx, request, into, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+request.Path, nil)
	})
}

func (c *Client) call(ctx context.Context, request Request, into any, build func() (*http.Request, error)) error {
	if !c.breaker.allow() {
		c.record(ctx, true, 0)
		return ErrCircuitOpen
	}
	attempts := 1
	if request.Retryable {
		attempts = c.retry.attempts()
	}
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		var status int
		status, err = c.attempt(ctx, request, into, build)
		if err == nil {
			c.breaker.succeeded()
			return nil
		}
		if attempt == attempts || !retryable(err, status) {
			break
		}
		if waitErr := wait(ctx, c.timer, c.retry.gapBefore(attempt, c.fraction())); waitErr != nil {
			break
		}
	}
	c.breaker.failed()
	return err
}

func (c *Client) attempt(ctx context.Context, request Request, into any, build func() (*http.Request, error)) (int, error) {
	call, err := build()
	if err != nil {
		return 0, ErrUnavailable
	}
	if request.Body != nil {
		call.Header.Set("Content-Type", "application/json")
	}
	for name, value := range request.Headers {
		call.Header.Set(name, value)
	}
	started := time.Now()
	status, err := c.send(call, into)
	c.record(ctx, err != nil, time.Since(started))
	return status, err
}

func (c *Client) send(call *http.Request, into any) (int, error) {
	response, err := c.http.Do(call)
	if err != nil {
		return 0, ErrUnavailable
	}
	defer func() { _ = response.Body.Close() }()
	if statusErr := statusError(response.StatusCode); statusErr != nil {
		return response.StatusCode, statusErr
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, MaximumResponseBytes))
	if readErr != nil {
		return response.StatusCode, ErrUnavailable
	}
	if into == nil {
		return response.StatusCode, nil
	}
	if err := json.Unmarshal(body, into); err != nil {
		return response.StatusCode, ErrMalformed
	}
	return response.StatusCode, nil
}

func (c *Client) record(ctx context.Context, failed bool, elapsed time.Duration) {
	if c.metrics == nil || c.provider == "" {
		return
	}
	c.metrics.Provider(ctx, c.provider, failed, elapsed)
}

func statusError(status int) error {
	switch {
	case status == http.StatusTooManyRequests:
		return ErrRateLimited
	case status >= 400:
		return ErrUnavailable
	default:
		return nil
	}
}
