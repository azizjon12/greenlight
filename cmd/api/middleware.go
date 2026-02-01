package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/azizjon12/greenlight/internal/data"
	"github.com/azizjon12/greenlight/internal/validator"
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

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, the value will be in format "Bearer <token>". Split into constituent parts.
		// If header is not in expected format, return a 401 Unauthorized response
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidCredentialsResponse(w, r)
			return
		}

		// Extract the actual authentication token from the header parts
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format
		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidCredentialsResponse(w, r)
			return
		}

		// Retrieve the details of the user associated with the authentication token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)

			default:
				app.serverErrorResponse(w, r, err)
			}

			return
		}

		// Call the contextSetUser() helper to add the user information to the request context
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// Create a new middleware to check that a user is not anonymous
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Check that a user is both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Rather than returning this http.HandlerFunc we assign it to the variable fn
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requireAuthenticatedUser() middleware before returning it
	return app.requireAuthenticatedUser(fn)
}
