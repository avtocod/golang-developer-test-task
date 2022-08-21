package main

import (
	"context"
	redclient2 "golang-developer-test-task/infrastructure/redclient"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	port := "3000"

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = logger.Sync()
	}()

	ctx := context.Background()
	conf := redclient2.RedisConfig{}
	conf.Load()

	client := redclient2.NewRedisClient(ctx, conf)
	defer func() {
		err = client.Close()
		if err != nil {
			panic(err)
		}
	}()

	dbLogic := NewDBProcessor(client, logger)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/load_file", dbLogic.HandleLoadFile)

	mux.HandleFunc("/api/load_from_url", dbLogic.HandleLoadFromURL)

	//https://nimblehq.co/blog/getting-started-with-redisearch
	mux.HandleFunc("/api/search", dbLogic.HandleSearch)

	mux.HandleFunc("/", dbLogic.HandleMainPage)

	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
