package main

import (
	"fmt"
	"os"

	"fastgo/internal/config"
	"fastgo/internal/infra/database"
	"fastgo/internal/infra/database/migrations"
	"fastgo/internal/shared/logger"
)

func main() {
	if err := run(); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed")
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger.Init(cfg.APP_NAME, cfg.APP_ENV)

	if err := database.Init(cfg); err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("database shutdown failed", "error", err)
		}
	}()

	return migrations.Run()
}
