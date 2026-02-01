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

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// Use the requireActivatedUser() middleware on our five /v1/movies** endpoints.
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser(app.listMovieHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser(app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// Add the route for the POST /v1/tokens/authentication endpoint
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Wrap the router with the rateLimit() middleware and use the authenticate() middleware on all requests
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
