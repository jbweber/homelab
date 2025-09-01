package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbweber/homelab/nook/internal/datastore"
	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
)

func TestMachineRepository_Save(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_Save"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
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
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_FindByID"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
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
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_FindByName"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	_, err = repo.Save(ctx, machine)
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
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_FindByIPv4"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
	ctx := context.Background()

	// Create a machine
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}

	_, err = repo.Save(ctx, machine)
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
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_FindAll"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
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

	_, err = repo.Save(ctx, machine1)
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
	ds, err := datastore.New(testutil.NewTestDSN("TestMachineRepository_DeleteByID"))
	require.NoError(t, err)

	repo := NewMachineRepository(ds.DB)
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
