package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/repository"
	"github.com/jbweber/homelab/nook/internal/testutil"
)

func TestNetworks_NetworksHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_NetworksHandler")
	defer cleanup()

	// Create test data
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)

	network := domain.Network{
		Name:        "test-network",
		Bridge:      "br0",
		Subnet:      "192.168.1.0/24",
		Gateway:     "192.168.1.1",
		DNSServers:  "8.8.8.8,8.8.4.4",
		Description: "Test network",
	}

	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	_, err = dhcpRepo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	// Create handler
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request
	req := httptest.NewRequest("GET", "/api/v0/networks", nil)
	w := httptest.NewRecorder()

	// Call handler
	networks.NetworksHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []domain.Network
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 network, got %d", len(response))
	}

	if response[0].Name != network.Name {
		t.Errorf("Expected name %s, got %s", network.Name, response[0].Name)
	}
}

func TestNetworks_CreateNetworkHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_CreateNetworkHandler")
	defer cleanup()

	// Create handler
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request body
	requestBody := domain.Network{
		Name:        "new-network",
		Bridge:      "br1",
		Subnet:      "192.168.2.0/24",
		Gateway:     "192.168.2.1",
		DNSServers:  "8.8.8.8",
		Description: "New test network",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest("POST", "/api/v0/networks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	networks.CreateNetworkHandler(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response domain.Network
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != requestBody.Name {
		t.Errorf("Expected name %s, got %s", requestBody.Name, response.Name)
	}
}

func TestNetworks_GetNetworkHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_GetNetworkHandler")
	defer cleanup()

	// Create test data
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)

	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Create handler
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request with URL parameter
	req := httptest.NewRequest("GET", "/api/v0/networks/"+strconv.FormatInt(savedNetwork.ID, 10), nil)
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(savedNetwork.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	networks.GetNetworkHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response domain.Network
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != savedNetwork.ID {
		t.Errorf("Expected ID %d, got %d", savedNetwork.ID, response.ID)
	}
}

func TestNetworks_UpdateNetworkHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_UpdateNetworkHandler")
	defer cleanup()

	// Create test data
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)

	network := domain.Network{
		Name:        "test-network",
		Bridge:      "br0",
		Subnet:      "192.168.1.0/24",
		Description: "Original description",
	}

	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Create handler
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request body
	requestBody := domain.Network{
		Name:        "test-network",
		Bridge:      "br0",
		Subnet:      "192.168.1.0/24",
		Description: "Updated description",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request with URL parameter
	req := httptest.NewRequest("PATCH", "/api/v0/networks/"+strconv.FormatInt(savedNetwork.ID, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(savedNetwork.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	networks.UpdateNetworkHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response domain.Network
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Description != requestBody.Description {
		t.Errorf("Expected description %s, got %s", requestBody.Description, response.Description)
	}
}

func TestNetworks_DeleteNetworkHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_DeleteNetworkHandler")
	defer cleanup()

	// Create test data
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)

	network := domain.Network{
		Name:   "test-network",
		Bridge: "br0",
		Subnet: "192.168.1.0/24",
	}

	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Create handler
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request with URL parameter
	req := httptest.NewRequest("DELETE", "/api/v0/networks/"+strconv.FormatInt(savedNetwork.ID, 10), nil)
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(savedNetwork.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	networks.DeleteNetworkHandler(w, req)

	// Check response
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestNetworks_CreateDHCPRangeHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_CreateDHCPRangeHandler")
	defer cleanup()

	// Create test network
	networkRepo := repository.NewNetworkRepository(db)
	network := domain.Network{Name: "test-network", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Create handler
	dhcpRepo := repository.NewDHCPRangeRepository(db)
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request body
	requestBody := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest("POST", "/api/v0/networks/"+strconv.FormatInt(savedNetwork.ID, 10)+"/dhcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(savedNetwork.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	networks.CreateDHCPRangeHandler(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response domain.DHCPRange
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.StartIP != requestBody.StartIP {
		t.Errorf("Expected start IP %s, got %s", requestBody.StartIP, response.StartIP)
	}
}

func TestNetworks_GetNetworkDHCPRangesHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDBWithMigrations(t, "TestNetworks_GetNetworkDHCPRangesHandler")
	defer cleanup()

	// Create test network and DHCP ranges
	networkRepo := repository.NewNetworkRepository(db)
	dhcpRepo := repository.NewDHCPRangeRepository(db)

	network := domain.Network{Name: "test-network", Bridge: "br0", Subnet: "192.168.1.0/24"}
	savedNetwork, err := networkRepo.Save(context.Background(), network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	dhcpRange := domain.DHCPRange{
		NetworkID: savedNetwork.ID,
		StartIP:   "192.168.1.100",
		EndIP:     "192.168.1.150",
		LeaseTime: "24h",
	}

	_, err = dhcpRepo.Save(context.Background(), dhcpRange)
	if err != nil {
		t.Fatalf("Failed to save DHCP range: %v", err)
	}

	// Create handler
	machineRepo := repository.NewMachineRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ipLeaseRepo := repository.NewIPLeaseRepository(db)
	
	api := NewAPIWithRepos(machineRepo, sshKeyRepo, networkRepo, dhcpRepo, ipLeaseRepo)
	networks := NewNetworks(api)

	// Create request with URL parameter
	req := httptest.NewRequest("GET", "/api/v0/networks/"+strconv.FormatInt(savedNetwork.ID, 10)+"/dhcp", nil)
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(savedNetwork.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	networks.GetNetworkDHCPRangesHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []domain.DHCPRange
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 DHCP range, got %d", len(response))
	}

	if response[0].StartIP != dhcpRange.StartIP {
		t.Errorf("Expected start IP %s, got %s", dhcpRange.StartIP, response[0].StartIP)
	}
}
