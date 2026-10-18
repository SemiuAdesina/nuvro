package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

func TestTracePropagationAndBoundedRoute(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })
	router := chi.NewRouter()
	router.Use(middleware.Trace(provider.Tracer("test"), nil))
	router.Get("/trips/{tripId}", func(w http.ResponseWriter, r *http.Request) {
		if got := trace.SpanContextFromContext(r.Context()).TraceID().String(); got != "00112233445566778899aabbccddeeff" {
			t.Error(got)
		}
		w.WriteHeader(503)
	})
	request := httptest.NewRequestWithContext(t.Context(), "GET", "/trips/private-user-id", nil)
	request.Header.Set("traceparent", "00-00112233445566778899aabbccddeeff-0011223344556677-01")
	router.ServeHTTP(httptest.NewRecorder(), request)
	spans := recorder.Ended()
	if len(spans) != 1 || spans[0].Name() != "GET /trips/{tripId}" || !spans[0].Parent().IsRemote() {
		t.Fatal(spans)
	}
}
