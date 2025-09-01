package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSSHKeyRepository_Save(t *testing.T) {
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_Save"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine first (SSH keys need a machine)
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	// Test Save (which should create a new SSH key)
	key := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey",
	}

	saved, err := repo.Save(ctx, key)
	require.NoError(t, err)
	assert.NotZero(t, saved.ID)
	assert.Equal(t, savedMachine.ID, saved.MachineID)
	assert.Equal(t, key.KeyText, saved.KeyText)
}

func TestSSHKeyRepository_FindByID(t *testing.T) {
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_FindByID"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine and SSH key first
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	key := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey",
	}
	savedKey, err := repo.Save(ctx, key)
	require.NoError(t, err)

	// Test FindByID
	found, err := repo.FindByID(ctx, savedKey.ID)
	require.NoError(t, err)
	assert.Equal(t, savedKey.ID, found.ID)
	assert.Equal(t, savedMachine.ID, found.MachineID)
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey", found.KeyText)

	// Test FindByID with non-existent ID
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSSHKeyRepository_FindByMachineID(t *testing.T) {
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_FindByMachineID"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	// Create multiple SSH keys for the machine
	key1 := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1",
	}
	_, err = repo.Save(ctx, key1)
	require.NoError(t, err)

	key2 := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2",
	}
	_, err = repo.Save(ctx, key2)
	require.NoError(t, err)

	// Test FindByMachineID
	keys, err := repo.FindByMachineID(ctx, savedMachine.ID)
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
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_FindAll"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine and some SSH keys
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	key1 := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1",
	}
	_, err = repo.Save(ctx, key1)
	require.NoError(t, err)

	key2 := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIkey2",
	}
	_, err = repo.Save(ctx, key2)
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
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_DeleteByID"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine and SSH key
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	key := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey",
	}
	savedKey, err := repo.Save(ctx, key)
	require.NoError(t, err)

	// Verify key exists
	exists, err := repo.ExistsByID(ctx, savedKey.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete the key
	err = repo.DeleteByID(ctx, savedKey.ID)
	require.NoError(t, err)

	// Verify key no longer exists
	exists, err = repo.ExistsByID(ctx, savedKey.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Try to find the deleted key
	_, err = repo.FindByID(ctx, savedKey.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSSHKeyRepository_ExistsByID(t *testing.T) {
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_ExistsByID"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine and SSH key
	machineRepo := NewMachineRepository(db)
	machine := domain.Machine{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	savedMachine, err := machineRepo.Save(ctx, machine)
	require.NoError(t, err)

	key := domain.SSHKey{
		MachineID: savedMachine.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey",
	}
	savedKey, err := repo.Save(ctx, key)
	require.NoError(t, err)

	// Test existing key
	exists, err := repo.ExistsByID(ctx, savedKey.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing key
	exists, err = repo.ExistsByID(ctx, 99999)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSSHKeyRepository_ErrorHandling(t *testing.T) {
	db, err := sql.Open("sqlite", testutil.NewTestDSN("TestSSHKeyRepository_ErrorHandling"))
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS machines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		hostname TEXT NOT NULL,
		ipv4 TEXT NOT NULL UNIQUE
	);`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		machine_id INTEGER NOT NULL,
		key_text TEXT NOT NULL,
		FOREIGN KEY (machine_id) REFERENCES machines(id)
	);`)
	require.NoError(t, err)

	repo := NewSSHKeyRepository(db)
	ctx := context.Background()

	// Test FindByID with non-existent key
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)

	// Test DeleteByID with non-existent key (should not error)
	err = repo.DeleteByID(ctx, 99999)
	assert.NoError(t, err) // SQLite DELETE on non-existent row doesn't error
}
