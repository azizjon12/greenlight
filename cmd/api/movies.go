package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing request, any inserted URL parameters will be 
	// stored in the request context. We can use the ParamsFromContext() function to 
	// retrieve a slice containing these parameter names and values
	params := httprouter.ParamsFromContext(r.Context())

	// In our project all movies will have a unique positive integer ID, but the value
	// returned by ByName() is always a string. So, we convert it to an integer. 
	// If the parameter could not be converted, or is less than 1, the ID will be invalid
	// http.NotFound will be used to return a 404 Not Found response
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// Insert the movie ID in a placeholder response
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}
