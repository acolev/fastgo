// @title FastGo API
// @version 1.0
// @description FastGo starter API documentation.
// @BasePath /api
// @schemes http https
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

	"fastgo/internal/bootstrap"
	"fastgo/internal/config"
	"fastgo/internal/shared/logger"

	"github.com/gofiber/fiber/v3"
)

func main() {
	if err := run(); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger.Init(cfg.APP_NAME, cfg.APP_ENV)

	app, err := bootstrap.New(cfg)
	if err != nil {
		return err
	}

	defer func() {
		if err := bootstrap.ShutdownProviders(); err != nil {
			logger.Error("providers shutdown failed", "error", err)
		}
	}()

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting server", "addr", cfg.AppAddr())

		if err := app.Listen(cfg.AppAddr()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if err := <-serverErr; err != nil {
		return err
	}

	return nil
}
