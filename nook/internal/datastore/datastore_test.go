package datastore

import (
	"testing"

	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_InMemory(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestNew_InMemory"))
	require.NoError(t, err)
	require.NotNil(t, ds.DB)
	// Check that tables exist by attempting a simple query
	_, err = ds.DB.Query("SELECT id, name, ipv4 FROM machines")
	require.NoError(t, err)
	_, err = ds.DB.Query("SELECT id, machine_id, key_text FROM ssh_keys")
	require.NoError(t, err)
}

func TestCreateMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestCreateMachine"))
	require.NoError(t, err)
	machine := Machine{
		Name: "test-machine",
		IPv4: "192.168.1.100",
	}
	created, err := ds.CreateMachine(machine)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, machine.Name, created.Name)
	assert.Equal(t, machine.IPv4, created.IPv4)

	// Test validation: missing name
	invalid := Machine{IPv4: "192.168.1.101"}
	_, err = ds.CreateMachine(invalid)
	assert.Error(t, err)
	// Test validation: missing IPv4
	invalid = Machine{Name: "no-ip"}
	_, err = ds.CreateMachine(invalid)
	assert.Error(t, err)
}

func TestGetMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachine"))
	require.NoError(t, err)
	machine := Machine{
		Name: "get-machine",
		IPv4: "10.0.0.1",
	}
	created, err := ds.CreateMachine(machine)
	require.NoError(t, err)
	got, err := ds.GetMachine(created.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
}

func TestListMachines(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestListMachines"))
	require.NoError(t, err)
	machines := []Machine{
		{Name: "machine1", IPv4: "10.0.0.1"},
		{Name: "machine2", IPv4: "10.0.0.2"},
		{Name: "machine3", IPv4: "10.0.0.3"},
	}
	for _, m := range machines {
		_, err := ds.CreateMachine(m)
		require.NoError(t, err)
	}
	got, err := ds.ListMachines()
	require.NoError(t, err)
	assert.Len(t, got, len(machines))
	names := map[string]bool{}
	for _, m := range got {
		names[m.Name] = true
	}
	for _, m := range machines {
		assert.True(t, names[m.Name], "machine %q not found in list", m.Name)
	}
}

func TestDeleteMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestDeleteMachine"))
	require.NoError(t, err)
	machine := Machine{Name: "delete-me", IPv4: "10.0.0.99"}
	created, err := ds.CreateMachine(machine)
	require.NoError(t, err)
	// Delete the machine
	err = ds.DeleteMachine(created.ID)
	require.NoError(t, err)
	// Try to get the deleted machine
	got, err := ds.GetMachine(created.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
	// Deleting a non-existent machine should not error
	err = ds.DeleteMachine(created.ID)
	assert.NoError(t, err)
}

func TestGetMachineByName(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachineByName"))
	require.NoError(t, err)
	machine := Machine{Name: "find-me", IPv4: "10.0.0.42"}
	created, err := ds.CreateMachine(machine)
	require.NoError(t, err)
	got, err := ds.GetMachineByName(machine.Name)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
	assert.Equal(t, created.IPv4, got.IPv4)
	// Test not found
	missing, err := ds.GetMachineByName("does-not-exist")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestGetMachineByIPv4(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachineByIPv4"))
	require.NoError(t, err)
	machine := Machine{Name: "ip-machine", IPv4: "192.168.1.200"}
	created, err := ds.CreateMachine(machine)
	require.NoError(t, err)
	got, err := ds.GetMachineByIPv4(machine.IPv4)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
	assert.Equal(t, created.IPv4, got.IPv4)
	// Test not found
	missing, err := ds.GetMachineByIPv4("10.0.0.254")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestCreateSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestCreateSSHKey"))
	require.NoError(t, err)

	// First, create a machine to associate the SSH key with
	machine := Machine{
		Name: "ssh-machine",
		IPv4: "192.168.1.150",
	}
	createdMachine, err := ds.CreateMachine(machine)
	require.NoError(t, err)

	// Now create an SSH key for that machine
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	require.NoError(t, err)
	assert.NotZero(t, keyID)

	// Verify the key was stored correctly
	retrievedKey, err := ds.GetSSHKey(keyID)
	require.NoError(t, err)
	require.NotNil(t, retrievedKey)
	assert.Equal(t, keyID, retrievedKey.ID)
	assert.Equal(t, createdMachine.ID, retrievedKey.MachineID)
	assert.Equal(t, keyText, retrievedKey.KeyText)
}

func TestListSSHKeys(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestListSSHKeys"))
	require.NoError(t, err)

	// Create a machine
	machine := Machine{
		Name: "list-keys-machine",
		IPv4: "192.168.1.160",
	}
	createdMachine, err := ds.CreateMachine(machine)
	require.NoError(t, err)

	// Initially should have no keys
	keys, err := ds.ListSSHKeys(createdMachine.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 0)

	// Add some keys
	key1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1 user@test"
	key2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey2 user@test"

	id1, err := ds.CreateSSHKey(createdMachine.ID, key1)
	require.NoError(t, err)
	id2, err := ds.CreateSSHKey(createdMachine.ID, key2)
	require.NoError(t, err)

	// List keys again
	keys, err = ds.ListSSHKeys(createdMachine.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	// Verify keys are in order (by ID)
	assert.Equal(t, id1, keys[0].ID)
	assert.Equal(t, id2, keys[1].ID)
	assert.Equal(t, key1, keys[0].KeyText)
	assert.Equal(t, key2, keys[1].KeyText)
	assert.Equal(t, createdMachine.ID, keys[0].MachineID)
	assert.Equal(t, createdMachine.ID, keys[1].MachineID)
}

func TestGetSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetSSHKey"))
	require.NoError(t, err)

	// Create a machine
	machine := Machine{
		Name: "get-key-machine",
		IPv4: "192.168.1.170",
	}
	createdMachine, err := ds.CreateMachine(machine)
	require.NoError(t, err)

	// Create an SSH key
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCgetkey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	require.NoError(t, err)

	// Get the key
	retrievedKey, err := ds.GetSSHKey(keyID)
	require.NoError(t, err)
	require.NotNil(t, retrievedKey)
	assert.Equal(t, keyID, retrievedKey.ID)
	assert.Equal(t, createdMachine.ID, retrievedKey.MachineID)
	assert.Equal(t, keyText, retrievedKey.KeyText)

	// Test getting non-existent key
	nonExistentKey, err := ds.GetSSHKey(99999)
	require.NoError(t, err)
	assert.Nil(t, nonExistentKey)
}

func TestDeleteSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestDeleteSSHKey"))
	require.NoError(t, err)

	// Create a machine
	machine := Machine{
		Name: "delete-key-machine",
		IPv4: "192.168.1.180",
	}
	createdMachine, err := ds.CreateMachine(machine)
	require.NoError(t, err)

	// Create an SSH key
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCdeletekey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	require.NoError(t, err)

	// Verify key exists
	retrievedKey, err := ds.GetSSHKey(keyID)
	require.NoError(t, err)
	require.NotNil(t, retrievedKey)

	// Delete the key
	err = ds.DeleteSSHKey(keyID)
	require.NoError(t, err)

	// Verify key no longer exists
	deletedKey, err := ds.GetSSHKey(keyID)
	require.NoError(t, err)
	assert.Nil(t, deletedKey)

	// Test deleting non-existent key (should not error)
	err = ds.DeleteSSHKey(99999)
	assert.NoError(t, err)
}
