package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang-developer-test-task/infrastructure/redclient"
	"golang-developer-test-task/structs"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestProcessJSONs(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)
	err := processor.processJSONs(errReader(0), processor.saveInfo)
	if err.Error() != "test error" {
		t.Fatal(err)
	}
}

func TestHandleMainPage(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleMainPage, "GET")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}

func TestHandleMainPageBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("POST", "/", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleMainPage, "GET")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleSearchBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("GET", "/api/search", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleSearchMode(t *testing.T) {
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

	var paginationSize int64 = 5
	mock.ExpectLLen(mode).SetVal(1)
	mock.ExpectLRange(mode, 0, paginationSize).SetVal([]string{info.SystemObjectID})
	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{Mode: &info.Mode}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchModeEn(t *testing.T) {
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

	var paginationSize int64 = 5
	mock.ExpectLLen(modeEn).SetVal(1)
	mock.ExpectLRange(modeEn, 0, paginationSize).SetVal([]string{info.SystemObjectID})
	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{ModeEn: &info.ModeEn}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchID(t *testing.T) {
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

	mock.ExpectGet(id).SetVal(info.SystemObjectID)
	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{ID: &info.ID}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchIDEn(t *testing.T) {
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

	mock.ExpectGet(idEn).SetVal(info.SystemObjectID)
	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{IDEn: &info.IDEn}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchSystemObjectID(t *testing.T) {
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

	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{SystemObjectID: &info.SystemObjectID}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchGlobalID(t *testing.T) {
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

	mock.ExpectGet(globalID).SetVal(info.SystemObjectID)
	mock.ExpectGet(info.SystemObjectID).SetVal(string(bs))

	client := &redclient.RedisClient{*db, 10}
	err := client.AddValue(context.Background(), info)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{GlobalID: &info.GlobalID}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	var result structs.PaginationObject
	err = easyjson.Unmarshal(res.Body.Bytes(), &result)
	fmt.Println(err)
	fmt.Println(result)

	if result.Size != 1 {
		t.Errorf("result.Size != 1; result = %v", result)
	}
	if len(result.Data) != 1 {
		t.Errorf("len(result.Data) != 1; result = %v", result)
	}
	if result.HasNext {
		t.Errorf("result has next page")
	}
	if result.HasPrevious {
		t.Errorf("result has previous page")
	}
	if result.Data[0] != info {
		t.Errorf("result data is not equal to expected; result data: %v ; expected: %v",
			result.Data[0], info)
	}
}

func TestHandleSearchErrDuringSearch(t *testing.T) {
	info := structs.Info{
		GlobalID:       42,
		SystemObjectID: "777",
		ID:             1,
		IDEn:           9,
		Mode:           "abc",
		ModeEn:         "cba",
	}

	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{GlobalID: &info.GlobalID}
	bs1, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs1))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("GET", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLResourceWithoutFile(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	defer server.Close()

	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	urlObject := structs.URLObject{URL: server.URL}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLResourceWithoutFileButRightContentType(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(32<<20+1))
		}),
	)
	defer server.Close()

	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	urlObject := structs.URLObject{URL: server.URL}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLResourceBadFile(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{"))
		}),
	)
	defer server.Close()

	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	urlObject := structs.URLObject{URL: server.URL}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLWrongResource(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	urlObject := structs.URLObject{URL: "https://a.a"}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLWrongURLResource(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	urlObject := structs.URLObject{URL: "://192.1./1"}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLNilBody(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("POST", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFileBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("GET", "/api/load_file", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLWrongMethod(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)
	req := httptest.NewRequest("POST", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFileWrongMethod(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)
	req := httptest.NewRequest("POST", "/api/load_file", nil)
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFile(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	filePath := "test_data/data.json"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("uploadFile", filepath.Base(file.Name()))
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_file", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}

func TestHandleLoadFileWithParenthesisProblem(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	filePath := "test_data/parenthesis_problem.json"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("uploadFile", filepath.Base(file.Name()))
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_file", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	h := processor.MethodMiddleware(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleSearchWithoutNilSearchObject(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	req := httptest.NewRequest("POST", "/api/search", nil)
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleSearchWithoutNecessaryParamsInsideSearchObject(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	searchObject := structs.SearchObject{}
	bs, _ := easyjson.Marshal(searchObject)

	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(bs))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFileWrongFileName(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db, 10}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := NewDBProcessor(client, logger)

	filePath := "test_data/data.json"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("wrongFileName", filepath.Base(file.Name()))
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_file", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	// processor.HandleLoadFile(res, req)
	h := processor.MethodMiddleware(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}
