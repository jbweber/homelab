package config

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.DBPath != "~/nook/data/nook.db" {
		t.Errorf("Expected DBPath '~/nook/data/nook.db', got '%s'", config.DBPath)
	}

	if config.Port != "8080" {
		t.Errorf("Expected Port '8080', got '%s'", config.Port)
	}
}

func TestConfig_expandPath_WithTilde(t *testing.T) {
	config := NewConfig()

	path := "~/test/path"
	expanded := config.expandPath(path)

	if strings.HasPrefix(expanded, "~/") {
		t.Errorf("Expected path to be expanded, got '%s'", expanded)
	}

	if !strings.HasSuffix(expanded, "test/path") {
		t.Errorf("Expected expanded path to end with 'test/path', got '%s'", expanded)
	}
}

func TestConfig_expandPath_WithoutTilde(t *testing.T) {
	config := NewConfig()

	path := "/absolute/path"
	expanded := config.expandPath(path)

	if expanded != path {
		t.Errorf("Expected path to remain unchanged, got '%s'", expanded)
	}
}

func TestConfig_expandPath_RelativePath(t *testing.T) {
	config := NewConfig()

	path := "relative/path"
	expanded := config.expandPath(path)

	if expanded != path {
		t.Errorf("Expected path to remain unchanged, got '%s'", expanded)
	}
}

func TestConfig_InitializeDatabase_Success(t *testing.T) {
	config := NewConfig()

	// Use a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.DBPath = filepath.Join(tempDir, "test.db")

	db, err := config.InitializeDatabase()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	// Verify database works
	err = db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Verify foreign keys are enabled
	var fkEnabled bool
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Errorf("Failed to check foreign keys: %v", err)
	}
	if !fkEnabled {
		t.Error("Expected foreign keys to be enabled")
	}
}

func TestConfig_InitializeDatabase_DirectoryCreation(t *testing.T) {
	config := NewConfig()

	// Use a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set path to a nested directory that doesn't exist
	config.DBPath = filepath.Join(tempDir, "nested", "path", "test.db")

	db, err := config.InitializeDatabase()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer db.Close()

	// Verify the nested directory was created
	dbDir := filepath.Dir(config.DBPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		t.Errorf("Expected directory to be created: %s", dbDir)
	}
}

func TestConfig_InitializeDatabase_InvalidPath(t *testing.T) {
	config := NewConfig()

	// Use an invalid path (cannot create directory in root on most systems without privileges)
	config.DBPath = "/root/invalid/nook.db"

	db, err := config.InitializeDatabase()
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Expected error for invalid path")
	}

	if !strings.Contains(err.Error(), "failed to create database directory") {
		t.Errorf("Expected directory creation error, got: %v", err)
	}
}

func TestConfig_runMigrations_Success(t *testing.T) {
	config := NewConfig()

	// Create a temporary database for testing
	tempDir, err := os.MkdirTemp("", "nook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	err = config.runMigrations(db)
	if err != nil {
		t.Errorf("Expected no error running migrations, got %v", err)
	}

	// Verify that migration tracking table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	if err != nil {
		t.Errorf("Expected schema_migrations table to exist: %v", err)
	}
}

func TestConfig_runMigrations_DatabaseError(t *testing.T) {
	config := NewConfig()

	// Create a closed database to force an error
	tempDir, err := os.MkdirTemp("", "nook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	db.Close() // Close the database to force errors

	err = config.runMigrations(db)
	if err == nil {
		t.Fatal("Expected error running migrations on closed database")
	}
}
