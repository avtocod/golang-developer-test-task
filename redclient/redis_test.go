package redclient

import (
	"context"
	"fmt"
	"golang-developer-test-task/structs"

	"github.com/go-redis/redis/v9"
	"github.com/go-redis/redismock/v8"
	"github.com/mailru/easyjson"
	"testing"
)

func TestAddValue(t *testing.T) {
	info := structs.Info{
		GlobalID:       42,
		SystemObjectID: "777",
		ID:             1,
		IDEn:           9,
		Mode:           "abc",
		ModeEn:         "cba",
	}
	bs, _ := easyjson.Marshal(info)

	db, mock := redismock.NewClientMock()
	mock.ExpectSet(info.SystemObjectID, bs, 0).SetErr(redis.Nil)
	mock.ExpectSet(fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).SetErr(redis.Nil)
	mock.ExpectSet(fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).SetErr(redis.Nil)
	mock.ExpectSet(fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).SetErr(redis.Nil)
	mock.ExpectRPush(fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).SetErr(redis.Nil)
	mock.ExpectRPush(fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID, 0).SetErr(redis.Nil)

	client := &RedisClient{*db}
	err := client.AddValue(context.Background(), info)

	if err != redis.Nil {
		t.Fatal(err)
	}
}
