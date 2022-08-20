package main

import (
	"golang-developer-test-task/redclient"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redismock/v8"
	"go.uber.org/zap"
)


func TestHandleMainPage (t *testing.T) {
	t.Parallel()

	db, _ := redismock.NewClientMock()
	client := &redclient.RedisClient{*db}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = logger.Sync()
	}()

	processor := DBProcessor{client: client, logger: logger}

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	processor.HandleMainPage(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}
}
