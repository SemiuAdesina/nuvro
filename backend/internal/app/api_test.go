package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/app"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

func developmentConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load(func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	return cfg
}

func TestNewAPIBindsTheConfiguredPort(t *testing.T) {
	cfg := developmentConfig(t)
	cfg.HTTP.Port = 9317

	api, err := app.NewAPI(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("NewAPI returned error: %v", err)
	}
	if api.Address() != ":9317" {
		t.Fatalf("Address() = %q, want :9317", api.Address())
	}
}

func TestNewAPIRejectsUnusableLogLevel(t *testing.T) {
	cfg := developmentConfig(t)
	cfg.LogLevel = "loud"

	if _, err := app.NewAPI(cfg, &bytes.Buffer{}); err == nil {
		t.Fatal("NewAPI returned no error for an unusable log level")
	}
}

func TestAPIServesHealthAndShutsDownGracefully(t *testing.T) {
	api, err := app.NewAPI(developmentConfig(t), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("NewAPI returned error: %v", err)
	}

	listener, err := listen(t)
	if err != nil {
		t.Fatalf("listen returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- api.ServeListener(ctx, listener) }()

	baseURL := "http://" + listener.Addr().String()
	response := getWithRetry(t, baseURL+"/health/live")
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if response.Header.Get("X-Request-Id") == "" {
		t.Error("the live probe carried no request identifier")
	}

	var decoded struct {
		Status string `json:"status"`
	}
	if decodeErr := json.NewDecoder(response.Body).Decode(&decoded); decodeErr != nil {
		t.Fatalf("body is not JSON: %v", decodeErr)
	}
	if decoded.Status != "live" {
		t.Fatalf("status = %q, want live", decoded.Status)
	}

	cancel()
	select {
	case servedErr := <-served:
		if servedErr != nil {
			t.Fatalf("ServeListener returned error: %v", servedErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not shut down after cancellation")
	}

	closed, err := get(t, baseURL+"/health/live")
	if err == nil {
		_ = closed.Body.Close()
		t.Fatal("the server still accepts connections after shutdown")
	}
}

func TestAPIServeReportsAnUnusableAddress(t *testing.T) {
	cfg := developmentConfig(t)
	cfg.HTTP.Port = 70000

	api, err := app.NewAPI(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("NewAPI returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	err = api.Serve(ctx)
	if err == nil {
		t.Fatal("Serve returned no error for an unusable address")
	}
	if !strings.Contains(err.Error(), "listen") {
		t.Fatalf("error = %v, want a listen failure", err)
	}
}

func TestAPILogsListeningAddress(t *testing.T) {
	var logs bytes.Buffer
	api, err := app.NewAPI(developmentConfig(t), &logs)
	if err != nil {
		t.Fatalf("NewAPI returned error: %v", err)
	}

	listener, err := listen(t)
	if err != nil {
		t.Fatalf("listen returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- api.ServeListener(ctx, listener) }()

	warmup := getWithRetry(t, "http://"+listener.Addr().String()+"/health/live")
	_ = warmup.Body.Close()
	cancel()

	select {
	case <-served:
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not shut down after cancellation")
	}

	recorded := logs.String()
	for _, fragment := range []string{"api listening", "api shutting down", "api stopped"} {
		if !strings.Contains(recorded, fragment) {
			t.Errorf("log %q does not contain %q", recorded, fragment)
		}
	}
}

func listen(t *testing.T) (net.Listener, error) {
	t.Helper()
	var configuration net.ListenConfig
	return configuration.Listen(t.Context(), "tcp", "127.0.0.1:0")
}

func get(t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("building a request for %s failed: %v", url, err)
	}
	return http.DefaultClient.Do(request)
}

func getWithRetry(t *testing.T, url string) *http.Response {
	t.Helper()
	var lastErr error
	for range 50 {
		response, err := get(t, url)
		if err == nil {
			return response
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("GET %s never succeeded: %v", url, lastErr)
	return nil
}
