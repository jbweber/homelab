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
	if err := migrateDatabase(db); err != nil {
		return nil, err
	}
	return db, nil
}

// migrateDatabase creates tables for machines and ssh_keys if they do not exist.
func migrateDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY(machine_id) REFERENCES machines(id)
	);`)
	return err
}
