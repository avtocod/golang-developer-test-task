package redclient

import (
	"context"

	"github.com/go-redis/redis/v9"
)

// RedisClient is for wrapping original redis.Client
type RedisClient struct {
	redis.Client
}

// NewRedisClient is constructor for RedisClient
func NewRedisClient(ctx context.Context, config RedisConfig) *RedisClient {
	options := redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	}
	client := redis.NewClient(&options)

	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}
	return &RedisClient{*client}
}
