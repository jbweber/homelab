package repository

import (
	"context"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
)

func TestIPLeaseRepository_NewIPLeaseRepository(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_NewIPLeaseRepository")
	defer cleanup()

	repo := NewIPLeaseRepository(db)
	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
}

func TestIPLeaseRepository_Save_Create(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_Save_Create")
	defer cleanup()

	// Create dependencies first
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	// Create network
	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	// Create machine
	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}

	savedLease, err := repo.Save(context.Background(), lease)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if savedLease.ID == 0 {
		t.Error("Expected ID to be assigned")
	}
	if savedLease.IPAddress != lease.IPAddress {
		t.Errorf("Expected IP address %s, got %s", lease.IPAddress, savedLease.IPAddress)
	}
}

func TestIPLeaseRepository_Save_Update(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_Save_Update")
	defer cleanup()

	// Create dependencies
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	// Create initial lease
	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	savedLease, _ := repo.Save(context.Background(), lease)

	// Update the lease
	savedLease.LeaseTime = "12h"
	updatedLease, err := repo.Save(context.Background(), savedLease)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedLease.LeaseTime != "12h" {
		t.Errorf("Expected lease time '12h', got '%s'", updatedLease.LeaseTime)
	}
}

func TestIPLeaseRepository_FindByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_FindByID")
	defer cleanup()

	// Create dependencies and lease
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	savedLease, _ := repo.Save(context.Background(), lease)

	// Find by ID
	foundLease, err := repo.FindByID(context.Background(), savedLease.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if foundLease.IPAddress != lease.IPAddress {
		t.Errorf("Expected IP address %s, got %s", lease.IPAddress, foundLease.IPAddress)
	}
}

func TestIPLeaseRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_FindByID_NotFound")
	defer cleanup()

	repo := NewIPLeaseRepository(db)

	_, err := repo.FindByID(context.Background(), 999)
	if err == nil {
		t.Fatal("Expected error for non-existent ID")
	}
}

func TestIPLeaseRepository_FindAll(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_FindAll")
	defer cleanup()

	// Create dependencies
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	// Create second machine for second lease
	machine2 := domain.Machine{Name: "test-machine-2", Hostname: "test-machine-2", IPv4: "10.0.0.2", NetworkID: &savedNetwork.ID}
	savedMachine2, err := machineRepo.Save(context.Background(), machine2)
	if err != nil {
		t.Fatalf("Failed to save machine2: %v", err)
	}

	// Create multiple leases
	lease1 := domain.IPAddressLease{MachineID: savedMachine.ID, NetworkID: savedNetwork.ID, IPAddress: "192.168.1.100", LeaseTime: "24h"}
	lease2 := domain.IPAddressLease{MachineID: savedMachine2.ID, NetworkID: savedNetwork.ID, IPAddress: "192.168.1.101", LeaseTime: "12h"}
	_, err1 := repo.Save(context.Background(), lease1)
	if err1 != nil {
		t.Fatalf("Failed to save lease1: %v", err1)
	}
	_, err2 := repo.Save(context.Background(), lease2)
	if err2 != nil {
		t.Fatalf("Failed to save lease2: %v", err2)
	}

	leases, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(leases) != 2 {
		t.Errorf("Expected 2 leases, got %d", len(leases))
	}
}

func TestIPLeaseRepository_DeleteByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_DeleteByID")
	defer cleanup()

	// Create dependencies and lease
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	savedLease, _ := repo.Save(context.Background(), lease)

	// Delete the lease
	err := repo.DeleteByID(context.Background(), savedLease.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(context.Background(), savedLease.ID)
	if err == nil {
		t.Fatal("Expected error when finding deleted lease")
	}
}

func TestIPLeaseRepository_ExistsByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_ExistsByID")
	defer cleanup()

	// Create dependencies and lease
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	savedLease, _ := repo.Save(context.Background(), lease)

	// Test exists
	exists, err := repo.ExistsByID(context.Background(), savedLease.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected lease to exist")
	}

	// Test not exists
	exists, err = repo.ExistsByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected lease to not exist")
	}
}

func TestIPLeaseRepository_FindByIPAddress(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_FindByIPAddress")
	defer cleanup()

	// Create dependencies and lease
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	savedLease, _ := repo.Save(context.Background(), lease)

	// Find by IP address
	foundLease, err := repo.FindByIPAddress(context.Background(), "192.168.1.100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if foundLease.ID != savedLease.ID {
		t.Errorf("Expected lease ID %d, got %d", savedLease.ID, foundLease.ID)
	}
}

func TestIPLeaseRepository_AllocateIPAddress(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_AllocateIPAddress")
	defer cleanup()

	// Create dependencies
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	dhcpRepo := NewDHCPRangeRepository(db)
	repo := NewIPLeaseRepository(db)

	// Create network
	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	// Create machine
	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	// Create DHCP range
	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.110",
		LeaseTime: "24h",
	}
	dhcpRepo.Save(context.Background(), dhcpRange)

	// Allocate IP
	lease, err := repo.AllocateIPAddress(context.Background(), savedMachine.ID, savedNetwork.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if lease == nil {
		t.Fatal("Expected non-nil lease")
	}
	if lease.MachineID != savedMachine.ID {
		t.Errorf("Expected machine ID %d, got %d", savedMachine.ID, lease.MachineID)
	}
	if lease.NetworkID != savedNetwork.ID {
		t.Errorf("Expected network ID %d, got %d", savedNetwork.ID, lease.NetworkID)
	}
}

func TestIPLeaseRepository_DeallocateIPAddress(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_DeallocateIPAddress")
	defer cleanup()

	// Create dependencies
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	dhcpRepo := NewDHCPRangeRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.110",
		LeaseTime: "24h",
	}
	dhcpRepo.Save(context.Background(), dhcpRange)

	// Allocate IP first
	repo.AllocateIPAddress(context.Background(), savedMachine.ID, savedNetwork.ID)

	// Deallocate IP
	err := repo.DeallocateIPAddress(context.Background(), savedMachine.ID, savedNetwork.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify deallocation - should be able to find by machine ID and get empty result
	leases, err := repo.FindByMachineID(context.Background(), savedMachine.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(leases) != 0 {
		t.Errorf("Expected 0 leases after deallocation, got %d", len(leases))
	}
}

func TestIPLeaseRepository_IsIPAddressAvailable(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestIPLeaseRepository_IsIPAddressAvailable")
	defer cleanup()

	// Create dependencies
	networkRepo := NewNetworkRepository(db)
	machineRepo := NewMachineRepository(db)
	repo := NewIPLeaseRepository(db)

	network := domain.Network{Name: "test-net", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, _ := networkRepo.Save(context.Background(), network)

	machine := domain.Machine{Name: "test-machine", Hostname: "test-machine", IPv4: "10.0.0.1", NetworkID: &savedNetwork.ID}
	savedMachine, _ := machineRepo.Save(context.Background(), machine)

	// Test available IP
	available, err := repo.IsIPAddressAvailable(context.Background(), savedNetwork.ID, "192.168.1.100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !available {
		t.Error("Expected IP to be available")
	}

	// Create lease for the IP
	lease := domain.IPAddressLease{
		MachineID: savedMachine.ID,
		NetworkID: savedNetwork.ID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}
	repo.Save(context.Background(), lease)

	// Test unavailable IP
	available, err = repo.IsIPAddressAvailable(context.Background(), savedNetwork.ID, "192.168.1.100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if available {
		t.Error("Expected IP to be unavailable")
	}
}
