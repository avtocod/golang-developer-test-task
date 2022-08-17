package redis_db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/mailru/easyjson"
	"golang-developer-test-task/structs"
)

func AddValue(ctx context.Context, client *redis.Client, info structs.Info) (err error) {
	bs, err := easyjson.Marshal(info)
	err = client.Set(ctx, info.SystemObjectID, bs, 0).Err()
	if err != nil {
		return err
	}
	err = client.Set(ctx, fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = client.Set(ctx, fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = client.Set(ctx, fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = client.RPush(ctx, fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).Err()
	if err != nil {
		return err
	}
	err = client.RPush(ctx, fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID).Err()
	return err
}
