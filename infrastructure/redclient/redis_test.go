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

func TestRedisConfigLoad(t *testing.T) {
	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	t.Setenv("DB", "0")
	config := RedisConfig{}
	config.Load()
}

func TestRedisConfigLoadPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	// t.Setenv("DB", "0")
	config := RedisConfig{}
	config.Load()
}

func TestRedisConfigLoadPanicConvertDBToInt(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	t.Setenv("DB", "abracadabra")
	config := RedisConfig{}
	config.Load()
}

func TestNewRedisClientPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	config := RedisConfig{Addr: ":", Password: "", DB: 0}
	_ = NewRedisClient(context.Background(), config)
}

func TestNewRedisClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	config := RedisConfig{Addr: ":", Password: "", DB: 0}
	_ = NewRedisClient(context.Background(), config)
}

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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	client := &RedisClient{*db}
	err := client.AddValue(context.Background(), info)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAddValueGetErrInsideWatch(t *testing.T) {
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

	db, mock := redismock.NewClientMock()
	mock.ExpectWatch(info.SystemObjectID, globalID, id, idEn, mode, modeEn)

	client := &RedisClient{*db}
	err := client.AddValue(context.Background(), info)

	if err == nil {
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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	key := info.SystemObjectID
	mock.ExpectGet(key).SetVal(string(bs))
	client := &RedisClient{*db}

	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}
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

func TestFindValuesSingleIdEn(t *testing.T) {
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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	key := info.SystemObjectID
	mock.ExpectGet(idEn).SetVal(key)
	mock.ExpectGet(key).SetVal(string(bs))
	client := &RedisClient{*db}

	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}
	infoList, totalSize, err := client.FindValues(context.Background(), idEn, false, 0, 0)

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

func TestFindValuesSingleIdEnSecondGetErr(t *testing.T) {
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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	key := info.SystemObjectID
	mock.ExpectGet(idEn).SetVal(key)
	client := &RedisClient{*db}

	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.FindValues(context.Background(), idEn, false, 0, 0)

	if err == nil {
		t.Fatal(err)
	}
}

func TestFindValuesSingleIdEnFirstGetErr(t *testing.T) {
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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	client := &RedisClient{*db}

	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.FindValues(context.Background(), idEn, false, 0, 0)

	if err == nil {
		t.Fatal(err)
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
		t.Errorf("infoList length is not equal to 0; infoList: %v", infoList)
	}
	if size := len(infoList); int64(size) != totalSize {
		t.Errorf("infoList length is not equal to totalSize; len(infoList) = %d ; totalSize = %d", size, totalSize)
	}
}

func TestFindValuesMultipleZeroPaginationSize(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "777"
	mock.ExpectLLen(key).SetVal(0)
	client := &RedisClient{*db}

	infoList, _, err := client.FindValues(context.Background(), key, true, 0, 0)

	if err != nil {
		t.Fatal(err)
	}
	if len(infoList) != 0 {
		t.Errorf("infoList length is not equal to 0; infoList: %v", infoList)
	}
}

func TestFindValuesMultipleStartIsMoreThanSize(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "777"
	mock.ExpectLLen(key).SetVal(0)
	client := &RedisClient{*db}

	infoList, _, err := client.FindValues(context.Background(), key, true, 1, 1)

	if err != nil {
		t.Fatal(err)
	}
	if len(infoList) != 0 {
		t.Errorf("infoList length is not equal to 0; infoList: %v", infoList)
	}
}

func TestFindValuesMultiple(t *testing.T) {
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
	mock.ExpectSet(globalID, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(id, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectSet(idEn, info.SystemObjectID, 0).SetVal("OK")
	mock.ExpectRPush(mode, info.SystemObjectID).SetVal(0)
	mock.ExpectRPush(modeEn, info.SystemObjectID).SetVal(0)
	mock.ExpectTxPipelineExec()

	key := info.SystemObjectID
	var paginationSize int64 = 5
	mock.ExpectLLen(mode).SetVal(1)
	mock.ExpectLRange(mode, 0, paginationSize).SetVal([]string{info.SystemObjectID})
	mock.ExpectGet(key).SetVal(string(bs))
	client := &RedisClient{*db}

	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}
	infoList, totalSize, err := client.FindValues(context.Background(), mode, true, paginationSize, 0)

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
