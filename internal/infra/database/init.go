package database

import (
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fastgo/internal/config"
)

func Init(cfg *config.Config) {
	if strings.TrimSpace(cfg.DB_DSN) == "" {
		log.Fatal("DB_DSN is empty: set it in the environment or in a .env file")
	}

	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	SetDB(db)

}
