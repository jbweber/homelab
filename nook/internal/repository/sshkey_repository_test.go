package repository

import (
	"context"
	"testing"

	"github.com/jbweber/homelab/nook/internal/datastore"
	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHKeyRepository_Save(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_Save"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine first (SSH keys need a machine)
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	// Test Save (which should create a new SSH key)
	key := domain.SSHKey{
		MachineID: machine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey",
	}

	saved, err := repo.Save(ctx, key)
	require.NoError(t, err)
	assert.NotZero(t, saved.ID)
	assert.Equal(t, machine.ID, saved.MachineID)
	assert.Equal(t, key.KeyText, saved.KeyText)
}

func TestSSHKeyRepository_FindByID(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_FindByID"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine and SSH key first
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	keyID, err := ds.CreateSSHKey(machine.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey")
	require.NoError(t, err)

	// Test FindByID
	found, err := repo.FindByID(ctx, keyID)
	require.NoError(t, err)
	assert.Equal(t, keyID, found.ID)
	assert.Equal(t, machine.ID, found.MachineID)
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey", found.KeyText)

	// Test FindByID with non-existent ID
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSSHKeyRepository_FindByMachineID(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_FindByMachineID"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	// Create multiple SSH keys for the machine
	_, err = ds.CreateSSHKey(machine.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1")
	require.NoError(t, err)
	_, err = ds.CreateSSHKey(machine.ID, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2")
	require.NoError(t, err)

	// Test FindByMachineID
	keys, err := repo.FindByMachineID(ctx, machine.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1", keys[0].KeyText)
	assert.Equal(t, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2", keys[1].KeyText)

	// Test FindByMachineID with non-existent machine
	keys, err = repo.FindByMachineID(ctx, 99999)
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSSHKeyRepository_FindAll(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_FindAll"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine and some SSH keys
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	_, err = ds.CreateSSHKey(machine.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1")
	require.NoError(t, err)
	_, err = ds.CreateSSHKey(machine.ID, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2")
	require.NoError(t, err)

	// Test FindAll
	keys, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	// Should be ordered by ID
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1", keys[0].KeyText)
	assert.Equal(t, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2", keys[1].KeyText)
}

func TestSSHKeyRepository_DeleteByID(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_DeleteByID"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine and SSH key
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	keyID, err := ds.CreateSSHKey(machine.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey")
	require.NoError(t, err)

	// Verify key exists
	exists, err := repo.ExistsByID(ctx, keyID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete the key
	err = repo.DeleteByID(ctx, keyID)
	require.NoError(t, err)

	// Verify key no longer exists
	exists, err = repo.ExistsByID(ctx, keyID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Try to find the deleted key
	_, err = repo.FindByID(ctx, keyID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSSHKeyRepository_ExistsByID(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_ExistsByID"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Create a machine and SSH key
	machine, err := ds.CreateMachine(datastore.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	})
	require.NoError(t, err)

	keyID, err := ds.CreateSSHKey(machine.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey")
	require.NoError(t, err)

	// Test existing key
	exists, err := repo.ExistsByID(ctx, keyID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing key
	exists, err = repo.ExistsByID(ctx, 99999)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSSHKeyRepository_ErrorHandling(t *testing.T) {
	ds, err := datastore.New(testutil.NewTestDSN("TestSSHKeyRepository_ErrorHandling"))
	require.NoError(t, err)

	repo := NewSSHKeyRepository(ds.DB)
	ctx := context.Background()

	// Test FindByID with non-existent key
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)

	// Test DeleteByID with non-existent key (should not error)
	err = repo.DeleteByID(ctx, 99999)
	assert.NoError(t, err) // SQLite DELETE on non-existent row doesn't error
}
