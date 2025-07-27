package main

import (
	"context"
	"net/http"

	"github.com/embracexyz/greenlight/internal/data"
)

type contextKey string

const userKey contextKey = "user"

func (app *application) setContextUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

func (app *application) getContextUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
