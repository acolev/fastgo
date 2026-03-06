package database

import (
	"context"
	"fmt"
	"strings"

	"fastgo/internal/shared/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"fastgo/internal/config"
)

func Init(cfg *config.Config) error {
	if strings.TrimSpace(cfg.DB_DSN) == "" {
		return fmt.Errorf("DB_DSN is empty: set it in the environment or in a .env file")
	}

	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{
		Logger: logger.NewGORMLogger(cfg.APP_ENV, cfg.DB_LOG_LEVEL),
	})
	if err != nil {
		return fmt.Errorf("open primary database: %w", err)
	}

	if err := configurePrimaryPool(db, cfg); err != nil {
		return err
	}

	if err := configureResolver(db, cfg); err != nil {
		return err
	}

	SetDB(db)

	if err := Ping(context.Background()); err != nil {
		_ = Close()
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func configurePrimaryPool(db *gorm.DB, cfg *config.Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db handle: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.DB_MAX_IDLE_CONNS)
	sqlDB.SetMaxOpenConns(cfg.DB_MAX_OPEN_CONNS)
	sqlDB.SetConnMaxLifetime(cfg.DB_CONN_MAX_LIFETIME)
	sqlDB.SetConnMaxIdleTime(cfg.DB_CONN_MAX_IDLE_TIME)

	return nil
}

func configureResolver(db *gorm.DB, cfg *config.Config) error {
	if len(cfg.DB_READ_DSNS) == 0 {
		return nil
	}

	replicas := make([]gorm.Dialector, 0, len(cfg.DB_READ_DSNS))
	for _, dsn := range cfg.DB_READ_DSNS {
		replicas = append(replicas, postgres.Open(dsn))
	}

	resolver := dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}).
		SetMaxIdleConns(cfg.DB_MAX_IDLE_CONNS).
		SetMaxOpenConns(cfg.DB_MAX_OPEN_CONNS).
		SetConnMaxLifetime(cfg.DB_CONN_MAX_LIFETIME).
		SetConnMaxIdleTime(cfg.DB_CONN_MAX_IDLE_TIME)

	if err := db.Use(resolver); err != nil {
		return fmt.Errorf("register dbresolver: %w", err)
	}

	return nil
}
