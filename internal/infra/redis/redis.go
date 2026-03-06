package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func SetClient(c *redis.Client) {
	client = c
}

func Client() *redis.Client {
	if client == nil {
		panic("redis not initialized")
	}
	return client
}

func Ping(ctx context.Context) error {
	if client == nil {
		return errors.New("redis not initialized")
	}

	return client.Ping(ctx).Err()
}

func Close() error {
	if client == nil {
		return nil
	}

	current := client
	client = nil

	return current.Close()
}
