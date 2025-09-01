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
		db.Close()
		CleanupTestDB(dsn)
	}

	return db, cleanup
}
