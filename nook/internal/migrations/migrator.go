package migrations

import (
	"database/sql"
	"fmt"
	"sort"
)

// Migration represents a database migration with up and down functions
type Migration struct {
	Version int64
	Name    string
	Up      func(*sql.DB) error
	Down    func(*sql.DB) error
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: []Migration{},
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})
}

// RunMigrations runs all pending migrations
func (m *Migrator) RunMigrations() error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if migration.Version > currentVersion {
			if err := m.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getCurrentVersion returns the current migration version
func (m *Migrator) getCurrentVersion() (int64, error) {
	var version int64
	err := m.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(migration Migration) error {
	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			// Log error but don't fail if transaction is already committed
		}
	}()

	// Run the migration
	if err := migration.Up(m.db); err != nil {
		return err
	}

	// Record the migration
	_, err = tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", migration.Version, migration.Name)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// GetCurrentVersion returns the current migration version (public method)
func (m *Migrator) GetCurrentVersion() (int64, error) {
	return m.getCurrentVersion()
}

// GetMigrations returns all registered migrations
func (m *Migrator) GetMigrations() []Migration {
	return m.migrations
}
