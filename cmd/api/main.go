package main

import (
	"context"
	"errors"
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
	cfg := config.Load()
	app := bootstrap.New(cfg)
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
			log.Fatal(err)
		}
		return
	case <-ctx.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil && !errors.Is(err, fiber.ErrNotRunning) {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	if err := <-serverErr; err != nil {
		log.Fatal(err)
	}
}
