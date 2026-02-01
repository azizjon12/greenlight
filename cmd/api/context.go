package main

import (
	"context"
	"net/http"

	"github.com/azizjon12/greenlight/internal/data"
)

type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey constant
const userContextKey = contextKey("user")

// This method returns a new copy of the request with the provided User struct added to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Retrieves the User struct from the request context. This helper will be used when we logically
// expect there to be User struct value in the context, and if it doesn't exist it will firmly be a programmer error
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
