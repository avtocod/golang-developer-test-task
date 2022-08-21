package main

import (
	"bytes"
	"golang-developer-test-task/redclient"
	"golang-developer-test-task/structs"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

func TestHandleMainPage(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleMainPage, "GET")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}

func TestHandleMainPageBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("POST", "/", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleMainPage, "GET")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleSearchBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("GET", "/api/search", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleSearch, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("GET", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
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
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	urlObject := structs.URLObject{URL: server.URL}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLWrongResource(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	urlObject := structs.URLObject{URL: "https://a.a"}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFromURLWrongURLResource(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	urlObject := structs.URLObject{URL: "://192.1./1"}
	bs, err := easyjson.Marshal(urlObject)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLNilBody(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("POST", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

// func TestHandleLoadFromURLGoodURLWithErrorDueAdding(t *testing.T) {
//	db, _ := redismock.NewClientMock()
//	// TODO: add data to mock before it
//	client := &redclient.RedisClient{*db}
//
//	logger, _ := zap.NewProduction()
//	defer func() {
//		_ = logger.Sync()
//	}()
//
//	processor := DBProcessor{client: client, logger: logger}
//
//	urlObject := structs.URLObject{URL: "http://op.mos.ru/opendata/files/7704786030-TaxiParking/data-20200706T0000-structure-20200706T0000.json"}
//	bs, err := easyjson.Marshal(urlObject)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	req := httptest.NewRequest("POST", "/api/load_from_url", bytes.NewBuffer(bs))
//	res := httptest.NewRecorder()
//	// processor.HandleLoadFromURL(res, req)
//	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
//	h(res, req)
//
//	if res.Code != http.StatusOK {
//		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
//	}
//}

func TestHandleLoadFileBadRequest(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("GET", "/api/load_file", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandleLoadFromURLWrongMethod(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}
	req := httptest.NewRequest("POST", "/api/load_from_url", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFileWrongMethod(t *testing.T) {
	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}
	req := httptest.NewRequest("POST", "/api/load_file", nil)
	res := httptest.NewRecorder()
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFile(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

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
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}

func TestHandleLoadFileWithParenthesisProblem(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

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
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}

func TestHandleLoadFileWrongFileName(t *testing.T) {
	db, _ := redismock.NewClientMock()
	// TODO: add data to mock before it
	client := &redclient.RedisClient{*db}

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

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
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusInternalServerError)
	}
}
