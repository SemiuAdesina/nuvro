package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
)

func Trace(tracer trace.Tracer, metrics *telemetry.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, "http.request", trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()
			recorder := &recordingWriter{ResponseWriter: w}
			next.ServeHTTP(recorder, r.WithContext(ctx))
			route := Route(r)
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			span.SetName(r.Method + " " + route)
			span.SetAttributes(attribute.String("http.route", route), attribute.Int("http.response.status_code", status))
			if status >= 500 {
				span.SetStatus(codes.Error, "request failed")
			}
			if metrics != nil {
				metrics.HTTP(ctx, r.Method, route, status, time.Since(started))
			}
		})
	}
}

func Route(r *http.Request) string {
	if route := chi.RouteContext(r.Context()); route != nil && route.RoutePattern() != "" {
		return route.RoutePattern()
	}
	return "unmatched"
}
