package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Wraps the MovieModel. Other models like UserModel, PermissionModel will be added
type Models struct {
	Movies MovieModel
	Tokens TokenModel // Add a new Tokens field
	Users  UserModel
}

// Returns a Models struct containing the initialized MovieModel and others
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Tokens: TokenModel{DB: db}, // Initialize a new TokenModel instance
		Users:  UserModel{DB: db},
	}
}
