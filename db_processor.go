package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"golang-developer-test-task/redclient"
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
	"golang.org/x/text/encoding/charmap"
)

type (
	jsonObjectsProcessorFunc func(io.Reader) error

	// DBProcessor needs for dependency injection
	DBProcessor struct {
		client *redclient.RedisClient
		logger *zap.Logger
	}

	// Handler is type for handler function
	Handler func(http.ResponseWriter, *http.Request)
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
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during change encoding to cp1251",
			zap.Error(err))
		return err
	}
	var infoList structs.InfoList
	err = easyjson.Unmarshal(out, &infoList)
	if err != nil {
		f.logger.Error("error inside ProcessJSONs during Unmarshal",
			zap.Error(err))
		return err
	}

	for _, info := range infoList {
		// TODO: should we accumulate json objects to insert?
		//  or restrict number of goroutines?
		go func(info structs.Info) {
			err := f.client.AddValue(context.Background(), info)
			if err != nil {
				f.logger.Error("error inside ProcessJSONs in goroutine",
					zap.Error(err))
				return
			}
		}(info)
	}
	return nil
}

// CheckHandlerRequestMethod is a function to return wrapped handler
func (f *DBProcessor) CheckHandlerRequestMethod(handler Handler, validMethod string) Handler {
	// TODO: make this method private, make unwrapped handlers private
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != validMethod {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handler(w, r)
	}
}

// ProcessFileFromURL handle json file from URL
func (f *DBProcessor) ProcessFileFromURL(url string, processor jsonObjectsProcessorFunc) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		f.logger.Error("error inside ProcessFileFromURL",
			zap.Error(err))
		return err
	}
	if resp.ContentLength > 32<<20 {
		s := fmt.Sprintf("too big resp body: %d", resp.ContentLength)
		f.logger.Error(s)
		return errors.New(s)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		s := fmt.Sprintf("unsupported Content-Type: %s", contentType)
		f.logger.Error(s)
		return errors.New(s)
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

// HandleLoadFile is handler for /api/load_file
func (f *DBProcessor) HandleLoadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = f.ProcessFileFromRequest(r, "uploadFile", f.ProcessJSONs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HandleLoadFromURL is handler for /api/load_from_url
func (f *DBProcessor) HandleLoadFromURL(w http.ResponseWriter, r *http.Request) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var urlObj structs.URLObject
	err = easyjson.Unmarshal(bs, &urlObj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := url.Parse(urlObj.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = f.ProcessFileFromURL(urlObj.URL, f.ProcessJSONs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HandleSearch is handler for /api/search
func (f *DBProcessor) HandleSearch(w http.ResponseWriter, r *http.Request) {
	var searchObj structs.SearchObject

	ctx := r.Context()
	searchStr := ""
	multiple := false
	// TODO
	a := "1704691"
	searchObj.SystemObjectID = &a

	switch {
	case searchObj.SystemObjectID != nil:
		searchStr = *searchObj.SystemObjectID
	case searchObj.GlobalID != nil:
		searchStr = strconv.Itoa(*searchObj.GlobalID)
	case searchObj.ID != nil:
		searchStr = strconv.Itoa(*searchObj.ID)
	case searchObj.IDEn != nil:
		searchStr = strconv.Itoa(*searchObj.IDEn)
	case searchObj.Mode != nil:
		searchStr = *searchObj.Mode
		multiple = true
	case searchObj.ModeEn != nil:
		searchStr = *searchObj.ModeEn
		multiple = true
	}

	var v string
	var err error
	paginationObj := structs.PaginationObject{}
	paginationObj.Data = make(structs.InfoList, 0)
	if !multiple {
		v, err = f.client.Get(ctx, searchStr).Result()
		if err != redis.Nil {
			paginationObj.Size = 1
			var info structs.Info
			err = easyjson.Unmarshal([]byte(v), &info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			paginationObj.Data = append(paginationObj.Data, info)
		}
	} else {
		paginationSize := 5
		var start, end int64
		start = int64(searchObj.Offset)
		end = int64(searchObj.Offset + paginationSize)
		var vs []string
		vs, err = f.client.LRange(ctx, searchStr, start, end).Result()
		if err != redis.Nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		size, _ := f.client.LLen(ctx, searchStr).Result()
		paginationObj.Size = int(size)
		data := make(structs.InfoList, len(vs))
		for _, v := range vs {
			var info structs.Info
			err = easyjson.Unmarshal([]byte(v), &info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		paginationObj.Data = data
	}

	bs, err := easyjson.Marshal(paginationObj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=windows-1251")
	_, err = w.Write(bs)
	if err != nil {
		panic(err)
	}
}

// HandleMainPage is handler for main page
func (f *DBProcessor) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	tmp := time.Now().Unix()
	h := md5.New()
	_, err := io.WriteString(h, strconv.FormatInt(tmp, 10))
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
}
