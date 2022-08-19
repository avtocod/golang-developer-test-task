package redis_db

import (
	"context"

	"github.com/go-redis/redis/v9"
)

type RedisClient struct {
	redis.Client
}

func RedisConnect(ctx context.Context, config RedisConfig) *RedisClient {
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
