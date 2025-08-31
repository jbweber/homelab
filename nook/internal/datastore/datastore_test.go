package datastore

import (
	"fmt"
	"testing"
)

// testDSN returns a unique in-memory SQLite DSN for each test.
// This ensures tests do not share state and remain independent.
func testDSN(testID string) string {
	// Use a unique name per test, but still use shared cache for driver compatibility.
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", testID)
}

func TestNew_InMemory(t *testing.T) {
	ds, err := New(testDSN("TestNew_InMemory"))
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
	ds, err := New(testDSN("TestCreateMachine"))
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
	ds, err := New(testDSN("TestGetMachine"))
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
	ds, err := New(testDSN("TestListMachines"))
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
	ds, err := New(testDSN("TestDeleteMachine"))
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
	ds, err := New(testDSN("TestGetMachineByName"))
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
	ds, err := New(testDSN("TestGetMachineByIPv4"))
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
