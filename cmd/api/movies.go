package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/azizjon12/greenlight/internal/data"
	"github.com/azizjon12/greenlight/internal/validator"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information we expect to be in the HTTP
	// request body. It will be our *target decode destination*
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Use the readJSON() helper to decode the request body into the input struct.
	// If this returns an error, send the error message along with a 400 Bad Request code
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Initialze a new Validator instance
	v := validator.New()

	// Use the Check() method to execute our validation checks
	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// Using Unique() helper to check all values in the input.Genres slice are unique
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")

	// Checking for a failed check
	if !v.Valid() {
		app.failedValidatorResponse(w, r, v.Errors)
		return
	}

	// Dump the contents of the input struct in an HTTP response
	fmt.Fprintf(w, "%+v\n", input)
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
