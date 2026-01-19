package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	// If rate limiting is not enabled, return the next handler in the chain with no further action
	if !app.config.limiter.enabled {
		return next
	}

	// Define a client struct to hold the rate limiter and last seen time for each client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limits
	var (
		mu sync.Mutex
		// Update the map so the values are pointers to a client struct
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while cleanup is taking place
			mu.Lock()

			// Loop through all clients. Delete if not see within last 3 minutes
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when cleanup is complete
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the realip.FromRequest() function to get the clients' IP address
		ip := realip.FromRequest(r)

		// Lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if not already exist
			clients[ip] = &client{
				// Use the requests-per-second and burst values from the config struct
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		// Update the last seen time for the client
		clients[ip].lastSeen = time.Now()

		// Call the Allow() method on the rate limiter for the current IP address. If the request
		// is not allowed, unlock mutex and send 429 Too Many Requests code
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Unlock mutex before calling the next handler in the chain
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
