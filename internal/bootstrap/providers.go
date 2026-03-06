package bootstrap

import (
	"fastgo/internal/config"
	"fastgo/internal/infra/database"
	"fastgo/internal/infra/redis"
)

func InitProviders(cfg *config.Config) {
	database.Init(cfg)
	redis.Init(cfg)
}
