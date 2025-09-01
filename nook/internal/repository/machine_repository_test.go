package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/migrations"
	"github.com/jbweber/homelab/nook/internal/testutil"
	_ "modernc.org/sqlite"
)

func setupTestDBWithMigrations(t *testing.T, testName string) (*sql.DB, func()) {
	db, cleanup := testutil.SetupTestDB(t, testName)

	// Run migrations
	migrator := migrations.NewMigrator(db)
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	if err := migrator.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db, cleanup
}

func TestMachineRepository_Save(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_Save")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	saved, err := repo.Save(ctx, machine)
	require.NoError(t, err)
	assert.NotZero(t, saved.ID)
	assert.Equal(t, "test-machine", saved.Name)
	assert.Equal(t, "test-host", saved.Hostname)
	assert.Equal(t, "192.168.1.100", saved.IPv4)
}

func TestMachineRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_FindByID")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	saved, err := repo.Save(ctx, machine)
	require.NoError(t, err)

	// Find the machine
	found, err := repo.FindByID(ctx, saved.ID)
	require.NoError(t, err)
	assert.Equal(t, saved.ID, found.ID)
	assert.Equal(t, "test-machine", found.Name)
	assert.Equal(t, "test-host", found.Hostname)
	assert.Equal(t, "192.168.1.100", found.IPv4)

	// Test not found
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMachineRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_FindByName")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	_, err := repo.Save(ctx, machine)
	require.NoError(t, err)

	// Find by name
	found, err := repo.FindByName(ctx, "test-machine")
	require.NoError(t, err)
	assert.Equal(t, "test-machine", found.Name)
	assert.Equal(t, "test-host", found.Hostname)
	assert.Equal(t, "192.168.1.100", found.IPv4)

	// Test not found
	_, err = repo.FindByName(ctx, "non-existent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMachineRepository_FindByIPv4(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_FindByIPv4")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	_, err := repo.Save(ctx, machine)
	require.NoError(t, err)

	// Find by IPv4
	found, err := repo.FindByIPv4(ctx, "192.168.1.100")
	require.NoError(t, err)
	assert.Equal(t, "test-machine", found.Name)
	assert.Equal(t, "test-host", found.Hostname)
	assert.Equal(t, "192.168.1.100", found.IPv4)

	// Test not found
	_, err = repo.FindByIPv4(ctx, "192.168.1.999")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMachineRepository_FindAll(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_FindAll")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create multiple machines
	machine1 := domain.Machine{
		Name:     "machine1",
		Hostname: "host1",
		IPv4:     "192.168.1.100",
	}
	machine2 := domain.Machine{
		Name:     "machine2",
		Hostname: "host2",
		IPv4:     "192.168.1.101",
	}

	_, err := repo.Save(ctx, machine1)
	require.NoError(t, err)
	_, err = repo.Save(ctx, machine2)
	require.NoError(t, err)

	// Find all
	machines, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, machines, 2)

	// Check that both machines are present
	names := make([]string, len(machines))
	for i, m := range machines {
		names[i] = m.Name
	}
	assert.Contains(t, names, "machine1")
	assert.Contains(t, names, "machine2")
}

func TestMachineRepository_DeleteByID(t *testing.T) {
	db, cleanup := setupTestDBWithMigrations(t, "TestMachineRepository_DeleteByID")
	defer cleanup()

	repo := NewMachineRepository(db)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	saved, err := repo.Save(ctx, machine)
	require.NoError(t, err)

	// Verify it exists
	exists, err := repo.ExistsByID(ctx, saved.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete it
	err = repo.DeleteByID(ctx, saved.ID)
	require.NoError(t, err)

	// Verify it's gone
	exists, err = repo.ExistsByID(ctx, saved.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Try to find it
	_, err = repo.FindByID(ctx, saved.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}
