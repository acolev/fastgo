package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"fastgo/internal/config"
)

const initPingTimeout = 3 * time.Second

func Init(cfg *config.Config) error {
	if strings.TrimSpace(cfg.REDIS_URL) == "" {
		return fmt.Errorf("REDIS_URL is empty: set it in the environment or in a .env file")
	}

	options, err := redisOptions(cfg.REDIS_URL)
	if err != nil {
		return err
	}

	rdb := redis.NewClient(options)

	pingCtx, cancel := context.WithTimeout(context.Background(), initPingTimeout)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return fmt.Errorf("ping redis: %w", err)
	}

	SetClient(rdb)

	return nil
}

func redisOptions(raw string) (*redis.Options, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("REDIS_URL is empty: set it in the environment or in a .env file")
	}

	if strings.Contains(raw, "://") {
		options, err := redis.ParseURL(raw)
		if err != nil {
			return nil, fmt.Errorf("parse REDIS_URL: %w", err)
		}

		return options, nil
	}

	return &redis.Options{Addr: raw}, nil
}
