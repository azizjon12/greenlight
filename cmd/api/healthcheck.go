package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	// Pass the map to json.Marshal() function. This returns a []byte slice containing
	// the encoded JSON. If there was an error, we log it and send the message to the client
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}

	// Append a newline to the JSON for easier view
	js = append(js, '\n')

	// Set necessary HTTP headers for a successful response
	w.Header().Set("Content-Type", "application/json")

	// Send the []byte slice containing the JSON as the response body
	w.Write(js)
}
