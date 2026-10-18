package telemetry_test

import (
	"testing"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
)

func TestHTTPMetricsRecordOutcome(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })
	metrics, err := telemetry.NewMetrics(provider.Meter("test"))
	if err != nil {
		t.Fatal(err)
	}
	metrics.HTTP(t.Context(), "GET", "/health/ready", 503, 25*time.Millisecond)
	var collected metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &collected); err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, scope := range collected.ScopeMetrics {
		for _, metric := range scope.Metrics {
			found[metric.Name] = true
		}
	}
	for _, name := range []string{"http.server.requests", "http.server.errors", "http.server.duration"} {
		if !found[name] {
			t.Errorf("missing %s", name)
		}
	}
}
