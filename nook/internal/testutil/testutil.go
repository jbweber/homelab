package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// CleanupTestDB removes the test database file
func CleanupTestDB(dsn string) error {
	// Extract file path from DSN
	if len(dsn) < 5 || dsn[:5] != "file:" {
		return fmt.Errorf("invalid DSN format")
	}

	path := dsn[5:]
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Skip cleanup for in-memory databases
	if path == ":memory:" || strings.Contains(dsn, "mode=memory") {
		return nil
	}

	// Only attempt to remove if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clean up
	}

	return os.Remove(path)
}

// SetupTestDB creates and returns a test database connection
func SetupTestDB(t *testing.T, testName string) (*sql.DB, func()) {
	dsn := NewTestDSN(testName)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	cleanup := func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: failed to close test database: %v", closeErr)
		}
		if cleanupErr := CleanupTestDB(dsn); cleanupErr != nil {
			t.Logf("Warning: failed to cleanup test database: %v", cleanupErr)
		}
	}

	return db, cleanup
}
