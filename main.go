package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
	"golang-developer-test-task/infrastructure/redis_db"
	"golang-developer-test-task/structs"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
type (
	JsonObjectsProcessorFunc func(io.Reader) error

	DBProcessor struct {
		client *redis.Client
		logger *zap.Logger
	}
)

func (f *DBProcessor) ProcessJSONs(reader io.Reader) (err error) {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during ReadAll",
			zap.Error(err))
		return err
	}
	var infoList structs.InfoList
	err = easyjson.Unmarshal(bs, &infoList)
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during Unmarshal",
			zap.Error(err))
		return err
	}
	for _, info := range infoList {
		// TODO: should we accumulate json objects to insert?
		//  or restrict number of goroutines?
		go func(info structs.Info) {
			bs, err := easyjson.Marshal(info)
			if err != nil {
				f.logger.Error("error inside ProcessJSONs' goroutine during Marshal",
					zap.Error(err))
				return
			}
			err = redis_db.AddValue(context.TODO(), f.client, info, bs)
			if err != nil {
				f.logger.Error("error inside ProcessJSONs' goroutine during adding value",
					zap.Error(err))
				return
			}
		}(info)
	}
	return nil
}

func (f *DBProcessor) ProcessFileFromURL(url string, processor JsonObjectsProcessorFunc) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		f.logger.Error("error inside ProcessFileFromURL",
			zap.Error(err))
		return err
	}
	defer resp.Body.Close()
	err = processor(resp.Body)
	return err
}

func (f *DBProcessor) ProcessFileFromRequest(r *http.Request, fileName string, processor JsonObjectsProcessorFunc) (err error) {
	file, _, err := r.FormFile(fileName)
	if err != nil {
		f.logger.Error("error inside ProcessFileFromRequest",
			zap.Error(err))
		return err
	}
	defer file.Close()
	err = processor(file)
	return err
}

type URLObject struct {
	URL string `json:"url"`
}

func main() {
	port := "3000"

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// write your code
	ctx := context.TODO()
	conf := redis_db.RedisConfig{}
	conf.Load()

	client := redis_db.RedisConnect(ctx, conf)
	defer client.Close()

	dbLogic := DBProcessor{client: client, logger: logger}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/load_file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = dbLogic.ProcessFileFromRequest(r, "uploadFile", dbLogic.ProcessJSONs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/api/load_from_url", func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		var urlObj URLObject
		err = json.Unmarshal(bs, &urlObj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := url.ParseRequestURI(urlObj.URL); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = dbLogic.ProcessFileFromURL(urlObj.URL, dbLogic.ProcessJSONs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	//mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
	//
	//})
	//
	//mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
	//
	//})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		t, _ := template.ParseFiles("static/index.tmpl")
		t.Execute(w, token)
	})

	//mux := DBProcessor{}
	http.ListenAndServe(":"+port, mux)
}
