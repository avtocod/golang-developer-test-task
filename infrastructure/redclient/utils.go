package redclient

import (
	"context"
	"fmt"
	"golang-developer-test-task/structs"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/mailru/easyjson"
)

// AddValue add info to Redis storage
func (r *RedisClient) AddValue(ctx context.Context, info structs.Info) (err error) {
	bs, err := easyjson.Marshal(info)
	if err != nil {
		return err
	}

	globalID := fmt.Sprintf("global_id:%d", info.GlobalID)
	id := fmt.Sprintf("id:%d", info.ID)
	idEn := fmt.Sprintf("id_en:%d", info.IDEn)
	mode := fmt.Sprintf("mode:%s", info.Mode)
	modeEn := fmt.Sprintf("mode_en:%s", info.ModeEn)

	txf := func(tx *redis.Tx) error {
		err := tx.Get(ctx, info.SystemObjectID).Err()
		if err != nil {
			return err
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, info.SystemObjectID, bs, 0)
			pipe.Set(ctx, globalID, info.SystemObjectID, 0)
			pipe.Set(ctx, id, info.SystemObjectID, 0)
			pipe.Set(ctx, idEn, info.SystemObjectID, 0)
			pipe.RPush(ctx, mode, info.SystemObjectID)
			pipe.RPush(ctx, modeEn, info.SystemObjectID)
			return nil
		})
		return err
	}

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err = r.Watch(ctx, txf, info.SystemObjectID, globalID, id, idEn, mode, modeEn)
		if err != redis.TxFailedErr {
			return err
		}
	}
	return err
}

// FindValues is a method for searching values by searchStr
func (r *RedisClient) FindValues(ctx context.Context, searchStr string, multiple bool, paginationSize, offset int64) (infoList structs.InfoList, totalSize int64, err error) {
	if !multiple {
		v, err := r.Get(ctx, searchStr).Result()
		if err != nil {
			return infoList, 0, err
		}
		if strings.Contains(searchStr, ":") {
			v, err = r.Get(ctx, v).Result()
			if err != nil {
				return infoList, 0, err
			}
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

	if paginationSize <= 0 {
		return infoList, size, nil
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
		vv, err := r.Get(ctx, v).Result()
		if err != nil {
			return infoList, size, err
		}
		err = easyjson.Unmarshal([]byte(vv), &info)
		if err != nil {
			return infoList, size, err
		}
		infoList = append(infoList, info)
	}
	return infoList, size, nil
}
