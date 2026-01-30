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
	Users  UserModel
}

// Returns a Models struct containing the initialized MovieModel and others
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
