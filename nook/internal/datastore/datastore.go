package datastore

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Datastore wraps the SQLite DB connection
// and provides methods for machine metadata management.
type Datastore struct {
	DB *sql.DB
}

// New creates and initializes the database connection.
func New(path string) (*Datastore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &Datastore{DB: db}, nil
}

// Close closes the database connection.
func (ds *Datastore) Close() error {
	return ds.DB.Close()
}
