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
	// TODO: add rollout when Set/RPush fails
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

// FindValues is a method for searching values by searchStr
func (r *RedisClient) FindValues(ctx context.Context, searchStr string, multiple bool, paginationSize, offset int64) (infoList structs.InfoList, totalSize int64, err error) {
	if !multiple {
		v, err := r.Get(ctx, searchStr).Result()
		if err != nil {
			return infoList, 0, err
		}
		var info structs.Info
		err = easyjson.Unmarshal([]byte(v), &info)
		if err != nil {
			return infoList, 1, err
		}
		infoList = append(infoList, info)
		return infoList, 1, nil
	}

	size, err := r.LLen(ctx, searchStr).Result()
	if err != nil {
		return infoList, 0, err
	}

	start := offset
	end := offset + paginationSize
	if start > size {
		return infoList, size, nil
	}

	var vs []string
	vs, err = r.LRange(ctx, searchStr, start, end).Result()
	if err != nil {
		return infoList, size, err
	}

	for _, v := range vs {
		var info structs.Info
		err = easyjson.Unmarshal([]byte(v), &info)
		if err != nil {
			return
		}
		infoList = append(infoList, info)
	}
	return infoList, size, nil
}
