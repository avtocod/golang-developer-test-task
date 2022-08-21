package redclient

import (
	"context"
	"fmt"
	"golang-developer-test-task/structs"
	"testing"

	"github.com/mailru/easyjson"

	"github.com/go-redis/redis/v9"
	"github.com/go-redis/redismock/v8"
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

	globalID := fmt.Sprintf("global_id:%d", info.GlobalID)
	id := fmt.Sprintf("id:%d", info.ID)
	idEn := fmt.Sprintf("id_en:%d", info.IDEn)
	mode := fmt.Sprintf("mode:%s", info.Mode)
	modeEn := fmt.Sprintf("mode_en:%s", info.ModeEn)

	bs, _ := easyjson.Marshal(info)

	db, mock := redismock.NewClientMock()
	mock.ExpectWatch(info.SystemObjectID, globalID, id, idEn, mode, modeEn)
	mock.ExpectGet(info.SystemObjectID).SetVal("")
	mock.ExpectTxPipeline()
	mock.ExpectSet(info.SystemObjectID, bs, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	client := &RedisClient{*db}
	err := client.AddValue(context.Background(), info)

	if err != nil {
		t.Fatal(err)
	}
}

func TestFindValuesNotFoundSingle(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "42"
	mock.ExpectGet(key).SetErr(redis.Nil)
	client := &RedisClient{*db}
	_, _, err := client.FindValues(context.Background(), key, false, 5, 0)

	if err != redis.Nil {
		t.Fatal(err)
	}
}

func TestFindValuesNotFoundMultiple(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "42"
	mock.ExpectLLen(key).SetErr(redis.Nil)
	client := &RedisClient{*db}
	_, _, err := client.FindValues(context.Background(), key, true, 5, 0)

	if err != redis.Nil {
		t.Fatal(err)
	}
}

func TestFindValuesSingle(t *testing.T) {
	info := structs.Info{
		GlobalID:       42,
		SystemObjectID: "777",
		ID:             1,
		IDEn:           9,
		Mode:           "abc",
		ModeEn:         "cba",
	}

	globalID := fmt.Sprintf("global_id:%d", info.GlobalID)
	id := fmt.Sprintf("id:%d", info.ID)
	idEn := fmt.Sprintf("id_en:%d", info.IDEn)
	mode := fmt.Sprintf("mode:%s", info.Mode)
	modeEn := fmt.Sprintf("mode_en:%s", info.ModeEn)

	bs, _ := easyjson.Marshal(info)

	db, mock := redismock.NewClientMock()
	mock.ExpectWatch(info.SystemObjectID, globalID, id, idEn, mode, modeEn)
	mock.ExpectGet(info.SystemObjectID).SetVal("")
	mock.ExpectTxPipeline()
	mock.ExpectSet(info.SystemObjectID, bs, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	key := info.SystemObjectID
	mock.ExpectGet(key).SetVal(string(bs))
	client := &RedisClient{*db}

	client.AddValue(context.Background(), info)
	infoList, totalSize, err := client.FindValues(context.Background(), key, false, 0, 0)

	if err != nil {
		t.Fatal(err)
	}
	if len(infoList) != 1 {
		t.Errorf("infoList length is not equal to 1; infoList: %v", infoList)
	}
	if size := len(infoList); int64(size) != totalSize {
		t.Errorf("infoList length is not equal to totalSize; len(infoList) = %d ; totalSize = %d", size, totalSize)
	}
	if info != infoList[0] {
		t.Errorf("data inside infoList is not info; info: %v ; infoList: %v", info, infoList)
	}
}

func TestFindValuesSingleNothing(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "777"
	mock.ExpectGet(key).SetErr(redis.Nil)
	client := &RedisClient{*db}

	infoList, totalSize, err := client.FindValues(context.Background(), key, false, 0, 0)

	if err != redis.Nil {
		t.Fatal(err)
	}
	if len(infoList) != 0 {
		t.Errorf("infoList length is not equal to 1; infoList: %v", infoList)
	}
	if size := len(infoList); int64(size) != totalSize {
		t.Errorf("infoList length is not equal to totalSize; len(infoList) = %d ; totalSize = %d", size, totalSize)
	}
}
