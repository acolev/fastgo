package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"fastgo/internal/config"
)

func Init(cfg *config.Config) error {
	if strings.TrimSpace(cfg.REDIS_URL) == "" {
		return fmt.Errorf("REDIS_URL is empty: set it in the environment or in a .env file")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.REDIS_URL,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return fmt.Errorf("ping redis: %w", err)
	}

	SetClient(rdb)

	return nil
}
