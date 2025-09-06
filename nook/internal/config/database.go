package config

import (
	"database/sql"
	"time"
)

// OptimizeDatabaseConnection applies performance optimizations to the database connection
func OptimizeDatabaseConnection(db *sql.DB) {
	// Set connection pool parameters for optimal performance
	db.SetMaxOpenConns(10)                 // Limit concurrent connections
	db.SetMaxIdleConns(5)                  // Keep some connections alive
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections periodically
	db.SetConnMaxIdleTime(1 * time.Minute) // Close idle connections after 1 minute
}

// ApplyPragmaOptimizations applies SQLite-specific performance pragmas
func ApplyPragmaOptimizations(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",    // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous = NORMAL",  // Balance between safety and performance
		"PRAGMA cache_size = 10000",    // Increase cache size (10MB)
		"PRAGMA temp_store = MEMORY",   // Store temporary tables in memory
		"PRAGMA mmap_size = 268435456", // 256MB memory mapping
		"PRAGMA optimize",              // Enable query optimizer
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}
