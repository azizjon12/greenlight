package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // Do not show in the output
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"` // Hide it if it is zero value
	Runtime   Runtime   `json:"runtime,omitzero"`
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}
