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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Decode the request body into the target destination
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		// If there is an error during the decoding, start the triage ...
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
		)

		switch {
		// Use error.As() function to check whether the error has the type
		// *json.SyntaxError. If so, return user friendly error message
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// Sometimes, Decode() may also return an io.ErrUnexpectedEOF error for syntax errorrs in the JSON
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// If the request body is empty, an io.EOF error is returned
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If something with non-nil pointer is passed, below error is returned
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For any other error, return it as-is
		default:
			return err
		}
	}
	return nil
}
