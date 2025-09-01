package repository

import (
	"context"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
)

func TestNetworkRepository_Save(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_Save")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Test creating a new network
	network := domain.Network{
		Name:        "test-network",
		Bridge:      "br0",
		Subnet:      "192.168.1.0/24",
		Gateway:     "192.168.1.1",
		DNSServers:  "8.8.8.8,8.8.4.4",
		Description: "Test network",
	}

	saved, err := repo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	if saved.ID == 0 {
		t.Error("Expected network ID to be set")
	}
	if saved.Name != network.Name {
		t.Errorf("Expected name %s, got %s", network.Name, saved.Name)
	}

	// Test updating the network
	saved.Description = "Updated test network"
	updated, err := repo.Save(context.Background(), saved)
	if err != nil {
		t.Fatalf("Failed to update network: %v", err)
	}

	if updated.Description != "Updated test network" {
		t.Errorf("Expected updated description, got %s", updated.Description)
	}
}

func TestNetworkRepository_FindByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_FindByID")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Create a test network
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	saved, err := repo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Find the network by ID
	found, err := repo.FindByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to find network: %v", err)
	}

	if found.ID != saved.ID {
		t.Errorf("Expected ID %d, got %d", saved.ID, found.ID)
	}
	if found.Name != network.Name {
		t.Errorf("Expected name %s, got %s", network.Name, found.Name)
	}
}

func TestNetworkRepository_FindByName(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_FindByName")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Create a test network
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	_, err := repo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Find the network by name
	found, err := repo.FindByName(context.Background(), "test-network")
	if err != nil {
		t.Fatalf("Failed to find network: %v", err)
	}

	if found.Name != network.Name {
		t.Errorf("Expected name %s, got %s", network.Name, found.Name)
	}
}

func TestNetworkRepository_FindAll(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_FindAll")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Create multiple test networks
	networks := []domain.Network{
		{Name: "network1", Bridge: "br0", Subnet: "192.168.1.0/24"},
		{Name: "network2", Bridge: "br1", Subnet: "192.168.2.0/24"},
		{Name: "network3", Bridge: "br2", Subnet: "192.168.3.0/24"},
	}

	for _, network := range networks {
		_, err := repo.Save(context.Background(), network)
		if err != nil {
			t.Fatalf("Failed to save network: %v", err)
		}
	}

	// Find all networks
	found, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("Failed to find networks: %v", err)
	}

	if len(found) != len(networks) {
		t.Errorf("Expected %d networks, got %d", len(networks), len(found))
	}
}

func TestNetworkRepository_DeleteByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_DeleteByID")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Create a test network
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	saved, err := repo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Delete the network
	err = repo.DeleteByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to delete network: %v", err)
	}

	// Verify it's deleted
	_, err = repo.FindByID(context.Background(), saved.ID)
	if err == nil {
		t.Error("Expected error when finding deleted network")
	}
}

func TestNetworkRepository_GetDHCPRanges(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_GetDHCPRanges")
	defer cleanup()

	networkRepo := NewNetworkRepository(db)
	dhcpRepo := NewDHCPRangeRepository(db)

	// Create a test network
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Create DHCP ranges for the network
	ranges := []domain.DHCPRange{
		{NetworkID: savedNetwork.ID, StartIP: "192.168.1.100", EndIP: "192.168.1.150", LeaseTime: "24h"},
		{NetworkID: savedNetwork.ID, StartIP: "192.168.1.200", EndIP: "192.168.1.250", LeaseTime: "12h"},
	}

	for _, dhcpRange := range ranges {
		_, err := dhcpRepo.Save(context.Background(), dhcpRange)
		if err != nil {
			t.Fatalf("Failed to save DHCP range: %v", err)
		}
	}

	// Get DHCP ranges for the network
	foundRanges, err := networkRepo.GetDHCPRanges(context.Background(), savedNetwork.ID)
	if err != nil {
		t.Fatalf("Failed to get DHCP ranges: %v", err)
	}

	if len(foundRanges) != len(ranges) {
		t.Errorf("Expected %d DHCP ranges, got %d", len(ranges), len(foundRanges))
	}
}

func TestNetworkRepository_ExistsByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworkRepository_ExistsByID")
	defer cleanup()

	repo := NewNetworkRepository(db)

	// Test non-existent network
	exists, err := repo.ExistsByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected network to not exist")
	}

	// Create a network
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	saved, err := repo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Test existing network
	exists, err = repo.ExistsByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected network to exist")
	}
}
