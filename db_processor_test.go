package main

import (
	"golang-developer-test-task/redclient"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redismock/v8"
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
	// processor.HandleMainPage(res, req)
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
	// processor.HandleMainPage(res, req)
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
	// processor.HandleSearch(res, req)
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
	// processor.HandleLoadFromURL(res, req)
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFromURL, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}

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
	// processor.HandleLoadFile(res, req)
	h := processor.CheckHandlerRequestMethod(processor.HandleLoadFile, "POST")
	h(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusBadRequest)
	}
}
