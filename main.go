package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	redclient2 "golang-developer-test-task/redclient"
	"golang-developer-test-task/structs"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type urlObject struct {
	URL string `json:"url"`
}

type searchObject struct {
	GlobalID       *int    `json:"global_id,omitempty"`
	SystemObjectID *string `json:"system_object_id,omitempty"`
	ID             *int    `json:"id,omitempty"`
	Mode           *string `json:"mode,omitempty"`
	IDEn           *int    `json:"id_en,omitempty"`
	ModeEn         *string `json:"mode_en,omitempty"`
	Offset         int     `json:"offset"`
}

type paginationObject struct {
	Size        int              `json:"size"`
	Offset      int              `json:"offset"`
	HasNext     bool             `json:"hasNext"`
	HasPrevious bool             `json:"hasPrevious"`
	Data        structs.InfoList `json:"data"`
	//Data []string `json:"data"`
}

func main() {
	port := "3000"

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = logger.Sync()
	}()

	ctx := context.TODO()
	conf := redclient2.RedisConfig{}
	conf.Load()

	client := redclient2.NewRedisClient(ctx, conf)
	defer func() {
		err = client.Close()
		if err != nil {
			panic(err)
		}
	}()

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
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var urlObj urlObject
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

	//https://nimblehq.co/blog/getting-started-with-redisearch
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		//var bs1 []byte
		//bs1, err := io.ReadAll(r.Body)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//defer r.Body.Close()
		var searchObj searchObject
		//err = json.Unmarshal(bs1, &searchObj)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}

		//ctx := context.TODO()
		ctx := r.Context()
		searchStr := ""
		multiple := false
		// TODO
		a := "1704691"
		//a := ""
		searchObj.SystemObjectID = &a

		if searchObj.SystemObjectID != nil {
			searchStr = *searchObj.SystemObjectID
		} else if searchObj.GlobalID != nil {
			// TODO: add global_id: and additional queries
			searchStr = strconv.Itoa(*searchObj.GlobalID)
		} else if searchObj.ID != nil {
			searchStr = strconv.Itoa(*searchObj.ID)
		} else if searchObj.IDEn != nil {
			searchStr = strconv.Itoa(*searchObj.IDEn)
		} else if searchObj.Mode != nil {
			searchStr = *searchObj.Mode
			multiple = true
		} else if searchObj.ModeEn != nil {
			searchStr = *searchObj.ModeEn
			multiple = true
		}

		var v string
		paginationObj := paginationObject{}
		paginationObj.Data = make(structs.InfoList, 0)
		if !multiple {
			v, err = client.Get(ctx, searchStr).Result()
			if err != redis.Nil {
				paginationObj.Size = 1
				var info structs.Info
				err = easyjson.Unmarshal([]byte(v), &info)
				paginationObj.Data = append(paginationObj.Data, info)
			}
		} else {
			paginationSize := 5
			var start, end int64
			start = int64(searchObj.Offset)
			end = int64(searchObj.Offset + paginationSize)
			var vs []string
			vs, err = client.LRange(ctx, searchStr, start, end).Result()
			size, _ := client.LLen(ctx, searchStr).Result()
			paginationObj.Size = int(size)
			data := make(structs.InfoList, len(vs))
			for _, v := range vs {
				var info structs.Info
				err = easyjson.Unmarshal([]byte(v), &info)
			}
			paginationObj.Data = data
		}
		//v, err := client.Get(ctx, "1704691").Result()
		//v, err := client.Get(ctx, "id:161").Result()
		//vs, err := client.LRange(ctx, "mode:круглосуточно", 0, -1).Result()
		//logger.Info("inside `api/search` during getting data from Redis",
		//	zap.String("val", fmt.Sprintf("%v", vs)))
		//v, err := client.Get(ctx, vs[0]).Result()
		//logger.Info("inside `api/search` during getting data from Redis",
		//	zap.String("val", fmt.Sprintf("%v", v)))
		//if err != redis.Nil {
		//	logger.Error("error inside `api/search` during getting data from Redis",
		//		zap.Error(err))
		//	w.WriteHeader(http.StatusInternalServerError)
		//}
		//var bs []byte
		//bs = v.([]byte)
		//bs1 = []byte(v)
		bs, err := json.Marshal(paginationObj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=windows-1251")
		_, err = w.Write(bs)
		if err != nil {
			panic(err)
		}
	})

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
		_, err := io.WriteString(h, strconv.FormatInt(crutime, 10))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		token := fmt.Sprintf("%x", h.Sum(nil))
		t, _ := template.ParseFiles("static/index.tmpl")
		err = t.Execute(w, token)
		if err != nil {
			panic(err)
		}
	})

	//mux := DBProcessor{}
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
