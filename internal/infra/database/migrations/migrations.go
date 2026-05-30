package migrations

import (
	"fmt"

	"fastgo/internal/infra/database"
	"fastgo/internal/models"
)

func Run() error {
	if err := database.DB().AutoMigrate(&models.Number{}); err != nil {
		return fmt.Errorf("auto migrate numbers: %w", err)
	}

	return nil
}
