package redis_db

import (
	"context"
	"github.com/go-redis/redis/v9"
	"golang-developer-test-task/structs"
)

func AddValueToSortedSet(ctx context.Context, client *redis.Client, value, collection string, score float64) (err error) {
	err = client.ZAdd(ctx, collection, redis.Z{Score: score, Member: value}).Err()
	return err
}

func AddValue(ctx context.Context, client *redis.Client, info structs.Info, bs []byte) (err error) {
	//err = client.ZAdd(ctx, collection, redis.Z{Score: score, Member: value}).Err()
	err = client.Set(ctx, info.SystemObjectID, bs, 0).Err()
	return err
}

func GetValueWithTheLeastScore(ctx context.Context, client *redis.Client, collection string) (result []redis.Z, err error) {
	result, err = client.ZPopMin(ctx, collection).Result()
	return
}

func GetStringValueWithTheLeastScore(ctx context.Context, client *redis.Client, collection string) (result []string, err error) {
	res, err := GetValueWithTheLeastScore(ctx, client, collection)
	for _, r := range res {
		result = append(result, r.Member.(string))
	}
	return
}
