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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fastgo/internal/bootstrap"
	"fastgo/internal/config"

	"github.com/gofiber/fiber/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	app, err := bootstrap.New(cfg)
	if err != nil {
		return err
	}

	defer func() {
		if err := bootstrap.ShutdownProviders(); err != nil {
			log.Printf("providers shutdown failed: %v", err)
		}
	}()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("starting server on %s", cfg.AppAddr())

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
		log.Println("shutdown signal received")
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
