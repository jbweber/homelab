package testutil

import (
	"testing"
)

func TestSetupTestDB(t *testing.T) {
	db, cleanup := SetupTestDB(t, "TestSetupTestDB")
	defer cleanup()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	// Verify database connection works
	err := db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Test that we can execute a query
	var result string
	err = db.QueryRow("SELECT 'test'").Scan(&result)
	if err != nil {
		t.Errorf("Test query failed: %v", err)
	}
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}
}

func TestSetupTestDBWithMigrations(t *testing.T) {
	db, cleanup := SetupTestDBWithMigrations(t, "TestSetupTestDBWithMigrations")
	defer cleanup()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	// Verify database connection works
	err := db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Verify migration tables exist (schema_migrations should be created by migrator)
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	if err != nil {
		t.Errorf("Expected schema_migrations table to exist: %v", err)
	}

	// Verify main application tables exist (machines, networks, etc.)
	tables := []string{"machines", "networks", "dhcp_ranges", "ssh_keys", "ip_address_leases"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Errorf("Error checking for table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("Expected table %s to exist", table)
		}
	}
}

func TestCleanupTestDB(t *testing.T) {
	// Test cleanup with in-memory database (should not error)
	dsn := NewTestDSN("test-cleanup")
	err := CleanupTestDB(dsn)
	if err != nil {
		t.Errorf("CleanupTestDB should not error on in-memory database: %v", err)
	}

	// Test cleanup with invalid DSN
	err = CleanupTestDB("invalid-dsn")
	if err == nil {
		t.Error("Expected error for invalid DSN")
	}
}

func TestSetupTestDB_MultipleInstances(t *testing.T) {
	// Test that we can create multiple test databases without conflicts
	db1, cleanup1 := SetupTestDB(t, "TestSetupTestDB_MultipleInstances_1")
	defer cleanup1()

	db2, cleanup2 := SetupTestDB(t, "TestSetupTestDB_MultipleInstances_2")
	defer cleanup2()

	// Both should work independently
	err1 := db1.Ping()
	err2 := db2.Ping()

	if err1 != nil {
		t.Errorf("First database failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second database failed: %v", err2)
	}

	// They should be separate instances
	if db1 == db2 {
		t.Error("Expected different database instances")
	}
}

func TestSetupTestDBWithMigrations_TableCreation(t *testing.T) {
	db, cleanup := SetupTestDBWithMigrations(t, "TestSetupTestDBWithMigrations_TableCreation")
	defer cleanup()

	// Test that we can insert data into created tables
	_, err := db.Exec("INSERT INTO networks (name, bridge, subnet) VALUES (?, ?, ?)", "test-net", "br0", "192.168.1.0/24")
	if err != nil {
		t.Errorf("Failed to insert into networks table: %v", err)
	}

	// Test that we can query the data back
	var name, bridge, subnet string
	err = db.QueryRow("SELECT name, bridge, subnet FROM networks WHERE name = ?", "test-net").Scan(&name, &bridge, &subnet)
	if err != nil {
		t.Errorf("Failed to query from networks table: %v", err)
	}

	if name != "test-net" || bridge != "br0" || subnet != "192.168.1.0/24" {
		t.Errorf("Unexpected data: name=%s, bridge=%s, subnet=%s", name, bridge, subnet)
	}
}

func TestCleanupTestDB_IdempotentCalls(t *testing.T) {
	dsn := NewTestDSN("test-idempotent")

	// Multiple cleanup calls should not panic or error
	err1 := CleanupTestDB(dsn)
	err2 := CleanupTestDB(dsn) // Second call should be safe
	err3 := CleanupTestDB(dsn) // Third call should be safe

	if err1 != nil {
		t.Errorf("First cleanup call failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second cleanup call failed: %v", err2)
	}
	if err3 != nil {
		t.Errorf("Third cleanup call failed: %v", err3)
	}
}
