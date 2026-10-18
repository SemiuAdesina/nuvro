package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	logger          *slog.Logger
	server          *http.Server
	shutdownTimeout time.Duration
	dependencies    *dependencies
}

func (a *API) Address() string {
	return a.server.Addr
}

func (a *API) Serve(ctx context.Context) error {
	var configuration net.ListenConfig

	listener, err := configuration.Listen(ctx, "tcp", a.server.Addr)
	if err != nil {
		cleanup, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		_ = a.dependencies.close(cleanup)
		return fmt.Errorf("app: listen on %s: %w", a.server.Addr, err)
	}
	return a.ServeListener(ctx, listener)
}

func (a *API) ServeListener(ctx context.Context, listener net.Listener) error {
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		if err := a.dependencies.close(cleanup); err != nil {
			a.logger.Error("dependency shutdown failed")
		}
	}()
	serverFailed := make(chan error, 1)
	go func() {
		serverFailed <- a.server.Serve(listener)
	}()

	a.logger.InfoContext(ctx, "api listening", slog.String("address", listener.Addr().String()))

	select {
	case err := <-serverFailed:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("app: serve: %w", err)
	case <-ctx.Done():
		return a.shutdown()
	}
}

func (a *API) shutdown() error {
	a.logger.Info("api shutting down", slog.Duration("grace", a.shutdownTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		_ = a.server.Close()
		return fmt.Errorf("app: shutdown: %w", err)
	}
	a.logger.Info("api stopped")
	return nil
}

func addressFromPort(port int) string {
	return net.JoinHostPort("", strconv.Itoa(port))
}
