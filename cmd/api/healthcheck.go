package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an envelope map containing the data for the response.
	// Here, environment and version data will now be nested under a system_info key in the JSON response
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		// Using the new serverErrorResponse() helper
		app.serverErrorResponse(w, r, err)
	}
}
