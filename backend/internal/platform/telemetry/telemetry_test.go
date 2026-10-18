package telemetry_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
)

func TestRuntimeProducesTraceAndCloses(t *testing.T) {
	runtime := telemetry.New()
	_, span := runtime.Tracer().Start(t.Context(), "test")
	if !span.SpanContext().IsValid() {
		t.Fatal("invalid trace context")
	}
	span.End()
	if err := runtime.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
}
