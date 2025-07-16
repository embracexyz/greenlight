package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (app *application) readJson(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var umarshalTypeError *json.UnmarshalTypeError
		var invalidUmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed json(at character %d)", syntaxError.Offset)
		case errors.As(err, &umarshalTypeError):
			if umarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect json type for field %q", umarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect json type(at character %d)", umarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUmarshalError):
			panic(err) // ?
		default:
			return err
		}
	}
	return nil
}
