package bootstrap

import (
	"errors"
	"fmt"

	"fastgo/internal/config"
	"fastgo/internal/infra/database"
	"fastgo/internal/infra/redis"
)

func InitProviders(cfg *config.Config) error {
	if err := database.Init(cfg); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	if err := redis.Init(cfg); err != nil {
		_ = database.Close()
		return fmt.Errorf("init redis: %w", err)
	}

	return nil
}

func ShutdownProviders() error {
	var errs []error

	if err := database.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close database: %w", err))
	}

	if err := redis.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close redis: %w", err))
	}

	return errors.Join(errs...)
}
