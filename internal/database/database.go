package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DB wraps the sql.DB to add custom functionality
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
