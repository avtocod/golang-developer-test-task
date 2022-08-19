package main

import (
	"context"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
	"golang-developer-test-task/redclient"
	"golang-developer-test-task/structs"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
)

type (
	jsonObjectsProcessorFunc func(io.Reader) error

	// DBProcessor needs for dependency injection
	DBProcessor struct {
		client *redclient.RedisClient
		logger *zap.Logger
	}
)

// ProcessJSONs read jsons from reader and write it to Redis client
func (f *DBProcessor) ProcessJSONs(reader io.Reader) (err error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during ReadAll",
			zap.Error(err))
		return err
	}
	dec := charmap.Windows1251.NewDecoder()
	out, err := dec.Bytes(bs)
	var infoList structs.InfoList
	err = easyjson.Unmarshal(out, &infoList)
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during Unmarshal",
			zap.Error(err))
		return err
	}
	//fmt.Println("INFOLIST", infoList)
	//fmt.Println()
	for _, info := range infoList {
		// TODO: should we accumulate json objects to insert?
		//  or restrict number of goroutines?
		go func(info structs.Info) {
			err = f.client.AddValue(context.Background(), info)
			if err != nil {
				f.logger.Error("error inside ProcessJSONs' goroutine during adding value",
					zap.Error(err))
				return
			}
		}(info)
	}
	return nil
}

// ProcessFileFromURL handle json file from URL
func (f *DBProcessor) ProcessFileFromURL(url string, processor jsonObjectsProcessorFunc) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		f.logger.Error("error inside ProcessFileFromURL",
			zap.Error(err))
		return err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	err = processor(resp.Body)
	return err
}

// ProcessFileFromRequest handle json file from request
func (f *DBProcessor) ProcessFileFromRequest(r *http.Request, fileName string, processor jsonObjectsProcessorFunc) (err error) {
	file, _, err := r.FormFile(fileName)
	if err != nil {
		f.logger.Error("error inside ProcessFileFromRequest",
			zap.Error(err))
		return err
	}
	defer func() {
		err = file.Close()
	}()
	err = processor(file)
	return err
}
