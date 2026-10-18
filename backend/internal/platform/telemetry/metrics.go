package telemetry

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	Requests         metric.Int64Counter
	Latency          metric.Float64Histogram
	Errors           metric.Int64Counter
	ProviderRequests metric.Int64Counter
	ProviderLatency  metric.Float64Histogram
	ProviderFailures metric.Int64Counter
	DatabaseLatency  metric.Float64Histogram
	RedisLatency     metric.Float64Histogram
}

func NewMetrics(meter metric.Meter) (*Metrics, error) {
	m := &Metrics{}
	var failures []error
	var err error
	m.Requests, err = meter.Int64Counter("http.server.requests")
	failures = append(failures, err)
	m.Errors, err = meter.Int64Counter("http.server.errors")
	failures = append(failures, err)
	m.Latency, err = meter.Float64Histogram("http.server.duration", metric.WithUnit("s"))
	failures = append(failures, err)
	m.ProviderRequests, err = meter.Int64Counter("provider.requests")
	failures = append(failures, err)
	m.ProviderFailures, err = meter.Int64Counter("provider.failures")
	failures = append(failures, err)
	m.ProviderLatency, err = meter.Float64Histogram("provider.duration", metric.WithUnit("s"))
	failures = append(failures, err)
	m.DatabaseLatency, err = meter.Float64Histogram("db.query.duration", metric.WithUnit("s"))
	failures = append(failures, err)
	m.RedisLatency, err = meter.Float64Histogram("redis.operation.duration", metric.WithUnit("s"))
	failures = append(failures, err)
	return m, errors.Join(failures...)
}

func (m *Metrics) HTTP(ctx context.Context, method, route string, status int, elapsed time.Duration) {
	options := metric.WithAttributes(attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", status))
	m.Requests.Add(ctx, 1, options)
	m.Latency.Record(ctx, elapsed.Seconds(), options)
	if status >= 400 {
		m.Errors.Add(ctx, 1, options)
	}
}

func (m *Metrics) Provider(ctx context.Context, provider string, failed bool, elapsed time.Duration) {
	if m == nil {
		return
	}
	options := metric.WithAttributes(attribute.String("provider.name", provider))
	m.ProviderRequests.Add(ctx, 1, options)
	if elapsed > 0 {
		m.ProviderLatency.Record(ctx, elapsed.Seconds(), options)
	}
	if failed {
		m.ProviderFailures.Add(ctx, 1, options)
	}
}
