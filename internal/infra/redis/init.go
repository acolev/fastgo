package redis

import (
	"github.com/redis/go-redis/v9"

	"fastgo/internal/config"
)

func Init(cfg *config.Config) {

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.REDIS_URL,
	})

	SetClient(rdb)

}
