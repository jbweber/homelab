package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbweber/homelab/nook/internal/migrations"
	_ "modernc.org/sqlite"
)

// Config holds all configuration for the nook service
type Config struct {
	DBPath string
	Port   string
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		DBPath: "~/nook/data/nook.db",
		Port:   "8080",
	}
}

// InitializeDatabase creates and configures the database connection
func (c *Config) InitializeDatabase() (*sql.DB, error) {
	dbPath := c.expandPath(c.DBPath)

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Apply performance optimizations
	OptimizeDatabaseConnection(db)

	if err := ApplyPragmaOptimizations(db); err != nil {
		return nil, fmt.Errorf("failed to apply performance optimizations: %w", err)
	}

	// Run migrations
	if err := c.runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// expandPath expands ~ to home directory
func (c *Config) expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Return original path if we can't get home dir
		return path
	}

	return filepath.Join(homeDir, path[2:])
}

// runMigrations runs all database migrations
func (c *Config) runMigrations(db *sql.DB) error {
	migrator := migrations.NewMigrator(db)

	// Add all migrations
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	// Run migrations
	if err := migrator.RunMigrations(); err != nil {
		return err
	}

	return nil
}
