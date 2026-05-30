package seeds

import (
	"context"
	"fmt"

	"fastgo/internal/shared/logger"

	"gorm.io/gorm"
)

type Seeder interface {
	Name() string
	Run(ctx context.Context, tx *gorm.DB) error
}

type ResettableSeeder interface {
	Seeder
	Reset(ctx context.Context, tx *gorm.DB) error
}

type Runner struct {
	db *gorm.DB
}

func NewRunner(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

func (r *Runner) Run(ctx context.Context, seeders []Seeder, fresh bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if fresh {
			for i := len(seeders) - 1; i >= 0; i-- {
				seeder, ok := seeders[i].(ResettableSeeder)
				if !ok {
					return fmt.Errorf("reset seed %q: reset is not supported", seeders[i].Name())
				}

				logger.Info("resetting seed", "seed", seeder.Name())
				if err := seeder.Reset(ctx, tx); err != nil {
					return fmt.Errorf("reset seed %q: %w", seeder.Name(), err)
				}
			}
		}

		for _, seeder := range seeders {
			logger.Info("running seed", "seed", seeder.Name())
			if err := seeder.Run(ctx, tx); err != nil {
				return fmt.Errorf("run seed %q: %w", seeder.Name(), err)
			}
		}

		return nil
	})
}
