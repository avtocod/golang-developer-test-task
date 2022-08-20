package redclient

import (
	"context"
	"fmt"
	"golang-developer-test-task/structs"

	"github.com/mailru/easyjson"
)

// AddValue add info to Redis storage
func (r *RedisClient) AddValue(ctx context.Context, info structs.Info) (err error) {
	bs, err := easyjson.Marshal(info)
	if err != nil {
		return err
	}
	err = r.Set(ctx, info.SystemObjectID, bs, 0).Err()
	if err != nil {
		return err
	}
	err = r.Set(ctx, fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = r.Set(ctx, fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = r.Set(ctx, fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).Err()
	if err != nil {
		return err
	}
	err = r.RPush(ctx, fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).Err()
	if err != nil {
		return err
	}
	err = r.RPush(ctx, fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID).Err()
	return err
}
