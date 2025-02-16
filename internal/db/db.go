package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("error running migrations: %v", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %v", err)
	}

	// Get list of migration files
	files, err := filepath.Glob("internal/db/migrations/*.sql")
	if err != nil {
		return fmt.Errorf("error finding migration files: %v", err)
	}
	sort.Strings(files)

	// Get applied migrations
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %v", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("error scanning migration version: %v", err)
		}
		applied[version] = true
	}

	// Apply new migrations
	for _, file := range files {
		version := getMigrationVersion(file)
		if applied[version] {
			continue
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %v", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error applying migration %s: %v", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("error recording migration %s: %v", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error committing migration %s: %v", file, err)
		}
	}

	return nil
}

func getMigrationVersion(filename string) int {
	base := filepath.Base(filename)
	version := strings.Split(base, "_")[0]
	var num int
	fmt.Sscanf(version, "%d", &num)
	return num
}
