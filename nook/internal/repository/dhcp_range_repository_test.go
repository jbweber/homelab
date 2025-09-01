package repository

import (
	"context"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/testutil"
)

func TestDHCPRangeRepository_Save(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_Save")
	defer cleanup()

	// Create a network first
	networkRepo := NewNetworkRepository(db)
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	// Test creating a new DHCP range
	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	saved, err := repo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	if saved.ID == 0 {
		t.Error("Expected DHCP range ID to be set")
	}
	if saved.StartIP != dhcpRange.StartIP {
		t.Errorf("Expected start IP %s, got %s", dhcpRange.StartIP, saved.StartIP)
	}

	// Test updating the DHCP range
	saved.LeaseTime = "48h"
	updated, err := repo.Save(context.Background(), saved)
	if err != nil {
		t.Fatalf("Failed to update DHCP range: %v", err)
	}

	if updated.LeaseTime != "48h" {
		t.Errorf("Expected updated lease time, got %s", updated.LeaseTime)
	}
}

func TestDHCPRangeRepository_FindByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_FindByID")
	defer cleanup()

	// Create a network first
	networkRepo := NewNetworkRepository(db)
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	// Create a test DHCP range
	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	saved, err := repo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	// Find the DHCP range by ID
	found, err := repo.FindByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to find DHCP range: %v", err)
	}

	if found.ID != saved.ID {
		t.Errorf("Expected ID %d, got %d", saved.ID, found.ID)
	}
	if found.StartIP != dhcpRange.StartIP {
		t.Errorf("Expected start IP %s, got %s", dhcpRange.StartIP, found.StartIP)
	}
}

func TestDHCPRangeRepository_FindByNetworkID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_FindByNetworkID")
	defer cleanup()

	// Create a network first
	networkRepo := NewNetworkRepository(db)
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	networkID := savedNetwork.ID

	// Create multiple DHCP ranges for the same network
	ranges := []domain.DHCPRange{
		{NetworkID: networkID, StartIP: "192.168.1.100", EndIP: "192.168.1.150", LeaseTime: "24h"},
		{NetworkID: networkID, StartIP: "192.168.1.200", EndIP: "192.168.1.250", LeaseTime: "12h"},
		{NetworkID: networkID, StartIP: "192.168.1.10", EndIP: "192.168.1.50", LeaseTime: "48h"},
	}

	for _, dhcpRange := range ranges {
		_, err := repo.Save(context.Background(), dhcpRange)
		if err != nil {
			t.Fatalf("Failed to save DHCP range: %v", err)
		}
	}

	// Find DHCP ranges by network ID
	found, err := repo.FindByNetworkID(context.Background(), networkID)
	if err != nil {
		t.Fatalf("Failed to find DHCP ranges: %v", err)
	}

	if len(found) != len(ranges) {
		t.Errorf("Expected %d DHCP ranges, got %d", len(ranges), len(found))
	}

	// Verify all ranges belong to the correct network
	for _, dhcpRange := range found {
		if dhcpRange.NetworkID != networkID {
			t.Errorf("Expected network ID %d, got %d", networkID, dhcpRange.NetworkID)
		}
	}
}

func TestDHCPRangeRepository_FindAll(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_FindAll")
	defer cleanup()

	// Create networks first
	networkRepo := NewNetworkRepository(db)
	network1 := domain.Network{Name: "network1", Bridge: "br0", Subnet: "192.168.1.0/24"}
	network2 := domain.Network{Name: "network2", Bridge: "br1", Subnet: "192.168.2.0/24"}
	savedNetwork1, err := networkRepo.Save(context.Background(), network1)
	if err != nil {
		t.Fatalf("Failed to save network1: %v", err)
	}
	savedNetwork2, err := networkRepo.Save(context.Background(), network2)
	if err != nil {
		t.Fatalf("Failed to save network2: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	// Create DHCP ranges for different networks
	ranges := []domain.DHCPRange{
		{NetworkID: savedNetwork1.ID, StartIP: "192.168.1.100", EndIP: "192.168.1.150", LeaseTime: "24h"},
		{NetworkID: savedNetwork2.ID, StartIP: "192.168.2.100", EndIP: "192.168.2.150", LeaseTime: "12h"},
		{NetworkID: savedNetwork1.ID, StartIP: "192.168.1.200", EndIP: "192.168.1.250", LeaseTime: "48h"},
	}

	for _, dhcpRange := range ranges {
		_, err := repo.Save(context.Background(), dhcpRange)
		if err != nil {
			t.Fatalf("Failed to save DHCP range: %v", err)
		}
	}

	// Find all DHCP ranges
	found, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("Failed to find DHCP ranges: %v", err)
	}

	if len(found) != len(ranges) {
		t.Errorf("Expected %d DHCP ranges, got %d", len(ranges), len(found))
	}
}

func TestDHCPRangeRepository_DeleteByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_DeleteByID")
	defer cleanup()

	// Create a network first
	networkRepo := NewNetworkRepository(db)
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	// Create a test DHCP range
	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	saved, err := repo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	// Delete the DHCP range
	err = repo.DeleteByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to delete DHCP range: %v", err)
	}

	// Verify it's deleted
	_, err = repo.FindByID(context.Background(), saved.ID)
	if err == nil {
		t.Error("Expected error when finding deleted DHCP range")
	}
}

func TestDHCPRangeRepository_ExistsByID(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestDHCPRangeRepository_ExistsByID")
	defer cleanup()

	// Create a network first
	networkRepo := NewNetworkRepository(db)
	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	repo := NewDHCPRangeRepository(db)

	// Test non-existent DHCP range
	exists, err := repo.ExistsByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected DHCP range to not exist")
	}

	// Create a DHCP range
	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	saved, err := repo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	// Test existing DHCP range
	exists, err = repo.ExistsByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected DHCP range to exist")
	}
}
