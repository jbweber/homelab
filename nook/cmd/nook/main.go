//go:build !test

// Code coverage for main is ignored for now. TODO: Add integration tests for main entrypoint.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jbweber/homelab/nook/internal/api"
	"github.com/jbweber/homelab/nook/internal/migrations"
	_ "modernc.org/sqlite"
)

func main() {
	// Initialize database
	db, err := initializeDatabase("nook.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register API routes
	api := api.NewAPI(db)
	api.RegisterRoutes(r)

	// Health check endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "Nook web service is running!"); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	fmt.Println("Starting Nook web service on :8080...")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// initializeDatabase creates a new SQLite database and runs migrations.
func initializeDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *sql.DB) error {
	migrator := migrations.NewMigrator(db)

	// Add all migrations
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	// Run migrations
	if err := migrator.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
