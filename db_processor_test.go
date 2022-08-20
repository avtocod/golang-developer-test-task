package main

import (
	"github.com/go-redis/redismock/v8"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/ozontech/cute"
	"go.uber.org/zap"
	"golang-developer-test-task/redclient"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type SuiteStruct struct {
	suite.Suite
	host *url.URL

	testMaker *cute.HTTPTestMaker
}

func (s *SuiteStruct) BeforeAll(t provider.T) {
	s.testMaker = cute.NewHTTPTestMaker()
}


func (s *SuiteStruct) TestAddValue(t provider.T) {
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
	//http.HandleFunc("/", processor.HandleMainPage)
	processor.HandleMainPage(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("got status %d but wanted %d", res.Code, http.StatusOK)
	}

	//cute.NewTestBuilder().
	//	Create().
	//	BeforeExecute()

	//info := structs.Info{
	//	GlobalID:       42,
	//	SystemObjectID: "777",
	//	ID:             1,
	//	IDEn:           9,
	//	Mode:           "abc",
	//	ModeEn:         "cba",
	//}
	//bs, _ := easyjson.Marshal(info)
	//
	//db, mock := redismock.NewClientMock()
	//mock.ExpectSet(info.SystemObjectID, bs, 0).SetErr(redis.Nil)
	//mock.ExpectSet(fmt.Sprintf("global_id:%d", info.GlobalID), info.SystemObjectID, 0).SetErr(redis.Nil)
	//mock.ExpectSet(fmt.Sprintf("id:%d", info.ID), info.SystemObjectID, 0).SetErr(redis.Nil)
	//mock.ExpectSet(fmt.Sprintf("id_en:%d", info.IDEn), info.SystemObjectID, 0).SetErr(redis.Nil)
	//mock.ExpectRPush(fmt.Sprintf("mode:%s", info.Mode), info.SystemObjectID).SetErr(redis.Nil)
	//mock.ExpectRPush(fmt.Sprintf("mode_en:%s", info.ModeEn), info.SystemObjectID, 0).SetErr(redis.Nil)
	//
	//client := &RedisClient{*db}
	//err := client.AddValue(context.Background(), info)
	//
	//if err != redis.Nil {
	//	t.Fail()
	//	return
	//}
}

func TestRun(t *testing.T) {
	suite.RunSuite(t, new(SuiteStruct))
}
