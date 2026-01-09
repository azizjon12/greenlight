package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// Wraps the MovieModel. Other models like UserModel, PermissionModel will be added
type Models struct {
	Movies MovieModel
}

// Returns a Models struct containing the initialized MovieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}