package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function that will always run in the event of panic
		defer func() {
			// Using bult-in recover() function to check if a panic occured.
			// Will return panic value if not will return nil
			pv := recover()
			if pv != nil {
				// If there was a panic, we close the current connection after sending the response "Connection: close" header
				w.Header().Set("Connection", "close")

				// recover() returns "any" type, we call fmt.Errorf() with "%v" to coerce it into an error
				// and call our serverErrorResponse helper
				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
