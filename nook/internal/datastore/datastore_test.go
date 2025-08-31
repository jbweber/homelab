package datastore

import (
	"fmt"
	"testing"

	"github.com/jbweber/homelab/nook/internal/testutil"
)

// testDSN returns a unique in-memory SQLite DSN for each test.
// This ensures tests do not share state and remain independent.
func testDSN(testID string) string {
	// Use a unique name per test, but still use shared cache for driver compatibility.
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", testID)
}

func TestNew_InMemory(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestNew_InMemory"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	if ds.DB == nil {
		t.Fatal("expected DB to be initialized")
	}
	// Check that tables exist by attempting a simple query
	_, err = ds.DB.Query("SELECT id, name, ipv4 FROM machines")
	if err != nil {
		t.Fatalf("machines table not found: %v", err)
	}
	_, err = ds.DB.Query("SELECT id, machine_id, key_text FROM ssh_keys")
	if err != nil {
		t.Fatalf("ssh_keys table not found: %v", err)
	}
}

func TestCreateMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestCreateMachine"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machine := Machine{
		Name: "test-machine",
		IPv4: "192.168.1.100",
	}
	created, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero machine ID")
	}
	if created.Name != machine.Name {
		t.Fatalf("expected machine name %q, got %q", machine.Name, created.Name)
	}
	if created.IPv4 != machine.IPv4 {
		t.Fatalf("expected machine IPv4 %q, got %q", machine.IPv4, created.IPv4)
	}

	// Test validation: missing name
	invalid := Machine{IPv4: "192.168.1.101"}
	_, err = ds.CreateMachine(invalid)
	if err == nil {
		t.Error("expected error for missing name")
	}
	// Test validation: missing IPv4
	invalid = Machine{Name: "no-ip"}
	_, err = ds.CreateMachine(invalid)
	if err == nil {
		t.Error("expected error for missing IPv4")
	}
}

func TestGetMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachine"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machine := Machine{
		Name: "get-machine",
		IPv4: "10.0.0.1",
	}
	created, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}
	got, err := ds.GetMachine(created.ID)
	if err != nil {
		t.Fatalf("failed to get machine: %v", err)
	}
	if got == nil {
		t.Fatal("expected machine to exist")
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Name != created.Name {
		t.Fatalf("expected name %q, got %q", created.Name, got.Name)
	}
}

func TestListMachines(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestListMachines"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machines := []Machine{
		{Name: "machine1", IPv4: "10.0.0.1"},
		{Name: "machine2", IPv4: "10.0.0.2"},
		{Name: "machine3", IPv4: "10.0.0.3"},
	}
	for _, m := range machines {
		_, err := ds.CreateMachine(m)
		if err != nil {
			t.Fatalf("failed to create machine %q: %v", m.Name, err)
		}
	}
	got, err := ds.ListMachines()
	if err != nil {
		t.Fatalf("failed to list machines: %v", err)
	}
	if len(got) != len(machines) {
		t.Fatalf("expected %d machines, got %d", len(machines), len(got))
	}
	names := map[string]bool{}
	for _, m := range got {
		names[m.Name] = true
	}
	for _, m := range machines {
		if !names[m.Name] {
			t.Errorf("machine %q not found in list", m.Name)
		}
	}
}

func TestDeleteMachine(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestDeleteMachine"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machine := Machine{Name: "delete-me", IPv4: "10.0.0.99"}
	created, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}
	// Delete the machine
	err = ds.DeleteMachine(created.ID)
	if err != nil {
		t.Fatalf("failed to delete machine: %v", err)
	}
	// Try to get the deleted machine
	got, err := ds.GetMachine(created.ID)
	if err != nil {
		t.Fatalf("error when getting deleted machine: %v", err)
	}
	if got != nil {
		t.Error("expected machine to be deleted, but it still exists")
	}
	// Deleting a non-existent machine should not error
	err = ds.DeleteMachine(created.ID)
	if err != nil {
		t.Errorf("expected no error when deleting non-existent machine, got: %v", err)
	}
}

func TestGetMachineByName(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachineByName"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machine := Machine{Name: "find-me", IPv4: "10.0.0.42"}
	created, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}
	got, err := ds.GetMachineByName(machine.Name)
	if err != nil {
		t.Fatalf("failed to get machine by name: %v", err)
	}
	if got == nil {
		t.Fatal("expected machine to exist")
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Name != created.Name {
		t.Fatalf("expected name %q, got %q", created.Name, got.Name)
	}
	if got.IPv4 != created.IPv4 {
		t.Fatalf("expected IPv4 %q, got %q", created.IPv4, got.IPv4)
	}
	// Test not found
	missing, err := ds.GetMachineByName("does-not-exist")
	if err != nil {
		t.Fatalf("unexpected error for missing machine: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for missing machine, got non-nil")
	}
}

func TestGetMachineByIPv4(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetMachineByIPv4"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}
	machine := Machine{Name: "ip-machine", IPv4: "192.168.1.200"}
	created, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}
	got, err := ds.GetMachineByIPv4(machine.IPv4)
	if err != nil {
		t.Fatalf("failed to get machine by IPv4: %v", err)
	}
	if got == nil {
		t.Fatal("expected machine to exist")
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Name != created.Name {
		t.Fatalf("expected name %q, got %q", created.Name, got.Name)
	}
	if got.IPv4 != created.IPv4 {
		t.Fatalf("expected IPv4 %q, got %q", created.IPv4, got.IPv4)
	}
	// Test not found
	missing, err := ds.GetMachineByIPv4("10.0.0.254")
	if err != nil {
		t.Fatalf("unexpected error for missing machine: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for missing machine, got non-nil")
	}
}

func TestCreateSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestCreateSSHKey"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}

	// First, create a machine to associate the SSH key with
	machine := Machine{
		Name: "ssh-machine",
		IPv4: "192.168.1.150",
	}
	createdMachine, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}

	// Now create an SSH key for that machine
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	if err != nil {
		t.Fatalf("failed to create SSH key: %v", err)
	}
	if keyID == 0 {
		t.Fatal("expected non-zero SSH key ID")
	}

	// Verify the key was stored correctly
	retrievedKey, err := ds.GetSSHKey(keyID)
	if err != nil {
		t.Fatalf("failed to retrieve SSH key: %v", err)
	}
	if retrievedKey == nil {
		t.Fatal("expected SSH key to exist")
	}
	if retrievedKey.ID != keyID {
		t.Fatalf("expected key ID %d, got %d", keyID, retrievedKey.ID)
	}
	if retrievedKey.MachineID != createdMachine.ID {
		t.Fatalf("expected machine ID %d, got %d", createdMachine.ID, retrievedKey.MachineID)
	}
	if retrievedKey.KeyText != keyText {
		t.Fatalf("expected key text %q, got %q", keyText, retrievedKey.KeyText)
	}
}

func TestListSSHKeys(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestListSSHKeys"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}

	// Create a machine
	machine := Machine{
		Name: "list-keys-machine",
		IPv4: "192.168.1.160",
	}
	createdMachine, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}

	// Initially should have no keys
	keys, err := ds.ListSSHKeys(createdMachine.ID)
	if err != nil {
		t.Fatalf("failed to list SSH keys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}

	// Add some keys
	key1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey1 user@test"
	key2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkey2 user@test"

	id1, err := ds.CreateSSHKey(createdMachine.ID, key1)
	if err != nil {
		t.Fatalf("failed to create first SSH key: %v", err)
	}
	id2, err := ds.CreateSSHKey(createdMachine.ID, key2)
	if err != nil {
		t.Fatalf("failed to create second SSH key: %v", err)
	}

	// List keys again
	keys, err = ds.ListSSHKeys(createdMachine.ID)
	if err != nil {
		t.Fatalf("failed to list SSH keys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	// Verify keys are in order (by ID)
	if keys[0].ID != id1 || keys[1].ID != id2 {
		t.Error("keys not returned in correct order")
	}
	if keys[0].KeyText != key1 || keys[1].KeyText != key2 {
		t.Error("key texts don't match")
	}
	if keys[0].MachineID != createdMachine.ID || keys[1].MachineID != createdMachine.ID {
		t.Error("machine IDs don't match")
	}
}

func TestGetSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestGetSSHKey"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}

	// Create a machine
	machine := Machine{
		Name: "get-key-machine",
		IPv4: "192.168.1.170",
	}
	createdMachine, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}

	// Create an SSH key
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCgetkey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	if err != nil {
		t.Fatalf("failed to create SSH key: %v", err)
	}

	// Get the key
	retrievedKey, err := ds.GetSSHKey(keyID)
	if err != nil {
		t.Fatalf("failed to get SSH key: %v", err)
	}
	if retrievedKey == nil {
		t.Fatal("expected SSH key to exist")
	}
	if retrievedKey.ID != keyID {
		t.Fatalf("expected key ID %d, got %d", keyID, retrievedKey.ID)
	}
	if retrievedKey.MachineID != createdMachine.ID {
		t.Fatalf("expected machine ID %d, got %d", createdMachine.ID, retrievedKey.MachineID)
	}
	if retrievedKey.KeyText != keyText {
		t.Fatalf("expected key text %q, got %q", keyText, retrievedKey.KeyText)
	}

	// Test getting non-existent key
	nonExistentKey, err := ds.GetSSHKey(99999)
	if err != nil {
		t.Fatalf("unexpected error for non-existent key: %v", err)
	}
	if nonExistentKey != nil {
		t.Error("expected nil for non-existent key")
	}
}

func TestDeleteSSHKey(t *testing.T) {
	ds, err := New(testutil.NewTestDSN("TestDeleteSSHKey"))
	if err != nil {
		t.Fatalf("failed to create datastore: %v", err)
	}

	// Create a machine
	machine := Machine{
		Name: "delete-key-machine",
		IPv4: "192.168.1.180",
	}
	createdMachine, err := ds.CreateMachine(machine)
	if err != nil {
		t.Fatalf("failed to create machine: %v", err)
	}

	// Create an SSH key
	keyText := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCdeletekey user@test"
	keyID, err := ds.CreateSSHKey(createdMachine.ID, keyText)
	if err != nil {
		t.Fatalf("failed to create SSH key: %v", err)
	}

	// Verify key exists
	retrievedKey, err := ds.GetSSHKey(keyID)
	if err != nil {
		t.Fatalf("failed to get SSH key: %v", err)
	}
	if retrievedKey == nil {
		t.Fatal("expected SSH key to exist before deletion")
	}

	// Delete the key
	err = ds.DeleteSSHKey(keyID)
	if err != nil {
		t.Fatalf("failed to delete SSH key: %v", err)
	}

	// Verify key no longer exists
	deletedKey, err := ds.GetSSHKey(keyID)
	if err != nil {
		t.Fatalf("failed to get SSH key after deletion: %v", err)
	}
	if deletedKey != nil {
		t.Error("expected SSH key to be deleted")
	}

	// Test deleting non-existent key (should not error)
	err = ds.DeleteSSHKey(99999)
	if err != nil {
		t.Fatalf("unexpected error deleting non-existent key: %v", err)
	}
}
