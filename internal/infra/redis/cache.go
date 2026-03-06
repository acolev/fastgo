package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fastgo/internal/shared/logger"

	goredis "github.com/redis/go-redis/v9"
)

const scanBatchSize = 100

type cacheCmdable interface {
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd
}

type JSONCache[T any] struct {
	prefix string
	ttl    time.Duration
	client cacheCmdable
}

func NewJSONCache[T any](prefix string, ttl time.Duration) JSONCache[T] {
	return JSONCache[T]{
		prefix: normalizeCachePart(prefix),
		ttl:    ttl,
	}
}

func (c JSONCache[T]) Key(parts ...any) string {
	keyParts := make([]string, 0, len(parts)+1)
	if c.prefix != "" {
		keyParts = append(keyParts, c.prefix)
	}

	for _, part := range parts {
		normalized := normalizeCachePart(fmt.Sprint(part))
		if normalized == "" {
			continue
		}

		keyParts = append(keyParts, normalized)
	}

	return strings.Join(keyParts, ":")
}

func (c JSONCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	var zero T

	cacheKey, err := c.cacheKey(key)
	if err != nil {
		return zero, false, err
	}

	payload, err := c.cmdable().Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return zero, false, nil
		}

		return zero, false, fmt.Errorf("redis get %q: %w", cacheKey, err)
	}

	var value T
	if err := json.Unmarshal([]byte(payload), &value); err != nil {
		return zero, false, fmt.Errorf("decode redis json %q: %w", cacheKey, err)
	}

	return value, true, nil
}

func (c JSONCache[T]) Set(ctx context.Context, key string, value T) error {
	cacheKey, err := c.cacheKey(key)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode redis json %q: %w", cacheKey, err)
	}

	if err := c.cmdable().Set(ctx, cacheKey, payload, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", cacheKey, err)
	}

	return nil
}

func (c JSONCache[T]) Remember(ctx context.Context, key string, loader func(context.Context) (T, error)) (T, error) {
	var zero T

	cacheKey, err := c.cacheKey(key)
	if err != nil {
		return zero, err
	}

	value, found, err := c.Get(ctx, key)
	if err != nil {
		logger.WarnContext(ctx, "redis cache get failed", "key", cacheKey, "error", err)
	} else if found {
		logger.DebugContext(ctx, "[cache:hit] "+cacheKey)
		return value, nil
	}

	logger.DebugContext(ctx, "[cache:miss] "+cacheKey)

	value, err = loader(ctx)
	if err != nil {
		return zero, err
	}

	if err := c.Set(ctx, key, value); err != nil {
		logger.WarnContext(ctx, "redis cache set failed", "key", cacheKey, "error", err)
	} else {
		logger.DebugContext(ctx, "[cache:set] "+cacheKey+" ttl="+c.ttl.String())
	}

	return value, nil
}

func (c JSONCache[T]) Invalidate(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	cacheKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		cacheKey, err := c.cacheKey(key)
		if err != nil {
			return err
		}

		cacheKeys = append(cacheKeys, cacheKey)
	}

	if err := c.cmdable().Del(ctx, cacheKeys...).Err(); err != nil {
		return fmt.Errorf("redis delete %q: %w", strings.Join(cacheKeys, ", "), err)
	}

	return nil
}

func (c JSONCache[T]) InvalidateAll(ctx context.Context) error {
	pattern := c.keyPattern()
	cursor := uint64(0)

	for {
		keys, nextCursor, err := c.cmdable().Scan(ctx, cursor, pattern, scanBatchSize).Result()
		if err != nil {
			return fmt.Errorf("redis scan %q: %w", pattern, err)
		}

		if len(keys) > 0 {
			if err := c.cmdable().Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis delete %q: %w", pattern, err)
			}
		}

		if nextCursor == 0 {
			return nil
		}

		cursor = nextCursor
	}
}

func (c JSONCache[T]) InvalidateBestEffort(ctx context.Context, keys ...string) {
	if err := c.Invalidate(ctx, keys...); err != nil {
		logger.WarnContext(ctx, "redis cache invalidate failed", "prefix", c.prefix, "error", err)
		return
	}

	logger.DebugContext(ctx, fmt.Sprintf("[cache:invalidate] %s keys=%d", c.prefix, len(keys)))
}

func (c JSONCache[T]) InvalidateAllBestEffort(ctx context.Context) {
	if err := c.InvalidateAll(ctx); err != nil {
		logger.WarnContext(ctx, "redis cache invalidate all failed", "prefix", c.prefix, "error", err)
		return
	}

	logger.DebugContext(ctx, "[cache:invalidate-all] "+c.prefix)
}

func (c JSONCache[T]) cmdable() cacheCmdable {
	if c.client != nil {
		return c.client
	}

	return Client()
}

func (c JSONCache[T]) cacheKey(key string) (string, error) {
	normalizedKey := normalizeCachePart(key)
	if normalizedKey == "" {
		return "", fmt.Errorf("redis cache key is empty")
	}

	if c.prefix == "" {
		return normalizedKey, nil
	}

	if normalizedKey == c.prefix || strings.HasPrefix(normalizedKey, c.prefix+":") {
		return normalizedKey, nil
	}

	return c.prefix + ":" + normalizedKey, nil
}

func (c JSONCache[T]) keyPattern() string {
	if c.prefix == "" {
		return "*"
	}

	return c.prefix + ":*"
}

func normalizeCachePart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, ":")

	return value
}
