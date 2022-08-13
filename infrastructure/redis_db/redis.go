package redis_db

import (
	"context"
	"github.com/go-redis/redis/v9"
)

func RedisConnect(ctx context.Context, config RedisConfig) *redis.Client {
	options := redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	}
	client := redis.NewClient(&options)

	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}
	return client
}
