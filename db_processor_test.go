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
	processor.HandleMainPage(res, req)

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
	processor.HandleMainPage(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}
