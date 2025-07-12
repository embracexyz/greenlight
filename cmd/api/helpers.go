package main

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	param := httprouter.ParamsFromContext(r.Context()).ByName("id")

	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil || id < 1 {
		return 0, ErrInvalidId
	}

	return id, nil
}
