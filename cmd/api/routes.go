package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize new httprouter instance
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the http.HandlerFunc() adapter,
	// and then set it as the custom error handler for 404 Not Found Response
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Doing the same for methodNotAllowed() helper
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register methods, URL patterns and handler functions for our endpoints
	// using HandlerFunc()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// Wrap the router with the panic recovery middleware
	return app.recoverPanic(router)
}
