package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

// Retrieve the "id" URL parameter from the current request context, then convert it
// to an integer and return it. If it is not successful, return 0 and an error
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Takes the destination http.ResponseWriter, the HTTP status code to send, the data to encode to JSON
// and a headers map containing any additional HTTP headers we want to include in the response
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Use the json.MarshalIndent() function so that whitespace is added to the encoded JSON.
	// Here, we use no line prefix ("") and tab indents ("\t") for each element
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	// Loop through the headers map (which has the type map[string][]string)
	// and add all the header key and values to the http.ResponseWriter's header map
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
