package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	// custom default handler for httprouter
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// 所有登录账户即可访问
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.authenticatedRequired(app.listMoviesHandler))
	// 登录且激活账户可访问
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.authenticatedActivated(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.authenticatedActivated(app.showMovieHandler))
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.authenticatedActivated(app.updateMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.authenticatedActivated(app.partialUpdateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.authenticatedActivated(app.deleteMovieHandler))

	// users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authentication(router)))
}
