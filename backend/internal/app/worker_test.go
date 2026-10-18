package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/app"
)

func TestNewWorkerRejectsUnusableLogLevel(t *testing.T) {
	cfg := developmentConfig(t)
	cfg.LogLevel = "loud"

	if _, err := app.NewWorker(cfg, &bytes.Buffer{}); err == nil {
		t.Fatal("NewWorker returned no error for an unusable log level")
	}
}

func TestWorkerRunsUntilCancelled(t *testing.T) {
	var logs bytes.Buffer
	worker, err := app.NewWorker(developmentConfig(t), &logs)
	if err != nil {
		t.Fatalf("NewWorker returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	select {
	case <-done:
		t.Fatal("Run returned before cancellation")
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the worker did not stop after cancellation")
	}

	recorded := logs.String()
	for _, fragment := range []string{"worker starting", "worker stopped"} {
		if !strings.Contains(recorded, fragment) {
			t.Errorf("log %q does not contain %q", recorded, fragment)
		}
	}
}
