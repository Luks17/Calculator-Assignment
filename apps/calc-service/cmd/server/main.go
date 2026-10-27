package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/app"
	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/infrastructure/config"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())
	if err := run(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("calculator server failed")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	return serve(ctx, srv)
}

func serve(ctx context.Context, srv *http.Server) error {
	zerolog.Ctx(ctx).Info().Str("address", srv.Addr).Msg("starting calculator server")
	stopped := make(chan error, 1)
	go func() {
		stopped <- srv.ListenAndServe()
	}()

	select {
	case err := <-stopped:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return gracefulShutdown(shutdownCtx, srv)
}

func gracefulShutdown(ctx context.Context, srv *http.Server) error {
	zerolog.Ctx(ctx).Info().Msg("shutting down calculator server")
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	zerolog.Ctx(ctx).Info().Msg("calculator server stopped")
	return nil
}
