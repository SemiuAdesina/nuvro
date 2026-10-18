package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Runtime struct {
	Traces  *sdktrace.TracerProvider
	Metrics *sdkmetric.MeterProvider
}

func New() *Runtime {
	return &Runtime{Traces: sdktrace.NewTracerProvider(), Metrics: sdkmetric.NewMeterProvider()}
}

func (r *Runtime) Tracer() trace.Tracer { return r.Traces.Tracer("nuvro") }
func (r *Runtime) Meter() metric.Meter  { return r.Metrics.Meter("nuvro") }
func (r *Runtime) Close(ctx context.Context) error {
	return errors.Join(r.Traces.Shutdown(ctx), r.Metrics.Shutdown(ctx))
}
