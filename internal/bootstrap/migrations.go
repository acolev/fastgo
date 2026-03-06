package bootstrap

import (
	"log"

	"fastgo/internal/infra/database"
	"fastgo/internal/models"
)

func RunMigrations() {
	if err := database.DB().AutoMigrate(&models.Number{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
}
