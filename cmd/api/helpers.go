package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// 包装返回的json内容，使其有一个明确的key，更直观；客户端且必须根据明确的key取值，更安全
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	param := httprouter.ParamsFromContext(r.Context()).ByName("id")

	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil || id < 1 {
		return 0, ErrInvalidId
	}

	return id, nil
}

func (app *application) writeJson(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
