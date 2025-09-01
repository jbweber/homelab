package migrations

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigrator_RunMigrations(t *testing.T) {
	// Create a test database
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", "TestMigrator_RunMigrations")
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: failed to close test database: %v", closeErr)
		}
	}()

	// Create migrator
	migrator := NewMigrator(db)

	// Add initial migrations
	for _, migration := range GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	// Run migrations
	err = migrator.RunMigrations()
	require.NoError(t, err)

	// Verify current version
	version, err := migrator.GetCurrentVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(2), version)

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='machines'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='ssh_keys'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='networks'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='dhcp_ranges'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify schema_migrations table exists
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify migration was recorded
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 2 AND name = 'create_networks_table'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMigrator_AddMigration(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", "TestMigrator_AddMigration")
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: failed to close test database: %v", closeErr)
		}
	}()

	migrator := NewMigrator(db)

	// Add migrations out of order
	migrator.AddMigration(Migration{Version: 3, Name: "third"})
	migrator.AddMigration(Migration{Version: 1, Name: "first"})
	migrator.AddMigration(Migration{Version: 2, Name: "second"})

	// Verify they are sorted
	migrations := migrator.GetMigrations()
	assert.Equal(t, int64(1), migrations[0].Version)
	assert.Equal(t, int64(2), migrations[1].Version)
	assert.Equal(t, int64(3), migrations[2].Version)
}

func TestUpgradeExistingTables(t *testing.T) {
	// Create a test database with old schema
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", "TestUpgradeExistingTables")
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: failed to close test database: %v", closeErr)
		}
	}()

	// Create old schema (without created_at/updated_at columns)
	_, err = db.Exec(`
		CREATE TABLE machines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			hostname TEXT NOT NULL,
			ipv4 TEXT NOT NULL UNIQUE
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE ssh_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			machine_id INTEGER NOT NULL,
			key_text TEXT NOT NULL,
			FOREIGN KEY (machine_id) REFERENCES machines(id)
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec("INSERT INTO machines (name, hostname, ipv4) VALUES (?, ?, ?)", "test-machine", "test-host", "192.168.1.100")
	require.NoError(t, err)

	machineID := int64(1)
	_, err = db.Exec("INSERT INTO ssh_keys (machine_id, key_text) VALUES (?, ?)", machineID, "ssh-rsa test-key")
	require.NoError(t, err)

	// Run upgrade
	err = upgradeExistingTables(db)
	require.NoError(t, err)

	// Verify new columns exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('machines') WHERE name='created_at'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('machines') WHERE name='updated_at'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify data is preserved
	var name, hostname, ipv4 string
	err = db.QueryRow("SELECT name, hostname, ipv4 FROM machines WHERE id = ?", machineID).Scan(&name, &hostname, &ipv4)
	require.NoError(t, err)
	assert.Equal(t, "test-machine", name)
	assert.Equal(t, "test-host", hostname)
	assert.Equal(t, "192.168.1.100", ipv4)

	var keyText string
	err = db.QueryRow("SELECT key_text FROM ssh_keys WHERE machine_id = ?", machineID).Scan(&keyText)
	require.NoError(t, err)
	assert.Equal(t, "ssh-rsa test-key", keyText)
}
