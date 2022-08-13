package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
	"golang-developer-test-task/infrastructure/redis_db"
	"golang-developer-test-task/infrastructure/structs"
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
	JsonObjectsProcessorFunc func(context.Context, io.Reader) error

	DBProcessor struct {
		client *redis.Client
	}
)

func (f *DBProcessor) ProcessJSONs(ctx context.Context, reader io.Reader) (err error) {
	dec := json.NewDecoder(reader)
	for dec.More() {
		var info structs.Info
		err = dec.Decode(&info)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// TODO: should we accumulate json objects to insert?
		//  or restrict number of goroutines?
		go func(info structs.Info) {
			err := redis_db.AddValue(ctx, f.client, info)
			if err != nil {
				return
			}
		}(info)
	}
	return nil
}

func ProcessFileFromURL(ctx context.Context, url string, processor JsonObjectsProcessorFunc) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = processor(ctx, resp.Body)
	return err
}

func ProcessFileFromRequest(r *http.Request, fileName string, processor JsonObjectsProcessorFunc) (err error) {
	file, _, err := r.FormFile(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	err = processor(r.Context(), file)
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

	dbLogic := DBProcessor{client: client}
	mux := http.NewServeMux()
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

	mux.HandleFunc("/api/load_file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//err := r.ParseMultipartForm(32 << 20)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		err := ProcessFileFromRequest(r, "uploadFile", dbLogic.ProcessJSONs)
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
		err = ProcessFileFromURL(r.Context(), urlObj.URL, dbLogic.ProcessJSONs)
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

	//mux := DBProcessor{}
	http.ListenAndServe(":"+port, mux)
}
