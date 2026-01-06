package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/azizjon12/greenlight/internal/data"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		// Using the new notFoundResponse() helper
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Create an envelope{"movie": movie} instance and pass it to writeJSON(),
	// instead of passing the plain movie struct
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// Using the new serverErrorResponse() helper
		app.serverErrorResponse(w, r, err)
	}
}
