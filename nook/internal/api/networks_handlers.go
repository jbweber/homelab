package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/domain"
)

// NetworksStore defines the datastore interface for network handlers
type NetworksStore interface {
	CreateNetwork(network domain.Network) (domain.Network, error)
	GetNetwork(id int64) (domain.Network, error)
	GetNetworkByName(name string) (domain.Network, error)
	ListNetworks() ([]domain.Network, error)
	UpdateNetwork(network domain.Network) (domain.Network, error)
	DeleteNetwork(id int64) error
	GetDHCPRanges(networkID int64) ([]domain.DHCPRange, error)
	CreateDHCPRange(dhcpRange domain.DHCPRange) (domain.DHCPRange, error)
	DeleteDHCPRange(id int64) error
}

// Networks groups network handlers for testability
type Networks struct {
	store NetworksStore
}

func NewNetworks(store NetworksStore) *Networks {
	return &Networks{store: store}
}

// NetworksHandler returns all networks
func (n *Networks) NetworksHandler(w http.ResponseWriter, r *http.Request) {
	networks, err := n.store.ListNetworks()
	if err != nil {
		log.Printf("failed to list networks: %v", err)
		http.Error(w, "failed to list networks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(networks); err != nil {
		log.Printf("failed to encode networks: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateNetworkHandler creates a new network
func (n *Networks) CreateNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var network domain.Network
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if network.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if network.Bridge == "" {
		http.Error(w, "bridge is required", http.StatusBadRequest)
		return
	}
	if network.Subnet == "" {
		http.Error(w, "subnet is required", http.StatusBadRequest)
		return
	}

	createdNetwork, err := n.store.CreateNetwork(network)
	if err != nil {
		log.Printf("failed to create network: %v", err)
		http.Error(w, "failed to create network", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdNetwork); err != nil {
		log.Printf("failed to encode created network: %v", err)
	}
}

// GetNetworkHandler gets a network by ID
func (n *Networks) GetNetworkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "network ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid network ID", http.StatusBadRequest)
		return
	}

	network, err := n.store.GetNetwork(id)
	if err != nil {
		log.Printf("failed to get network: %v", err)
		http.Error(w, "network not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(network); err != nil {
		log.Printf("failed to encode network: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateNetworkHandler updates a network
func (n *Networks) UpdateNetworkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "network ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid network ID", http.StatusBadRequest)
		return
	}

	var network domain.Network
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	network.ID = id
	updatedNetwork, err := n.store.UpdateNetwork(network)
	if err != nil {
		log.Printf("failed to update network: %v", err)
		http.Error(w, "failed to update network", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedNetwork); err != nil {
		log.Printf("failed to encode updated network: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteNetworkHandler deletes a network
func (n *Networks) DeleteNetworkHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "network ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid network ID", http.StatusBadRequest)
		return
	}

	if err := n.store.DeleteNetwork(id); err != nil {
		log.Printf("failed to delete network: %v", err)
		http.Error(w, "failed to delete network", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetNetworkDHCPRangesHandler gets DHCP ranges for a network
func (n *Networks) GetNetworkDHCPRangesHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "network ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid network ID", http.StatusBadRequest)
		return
	}

	ranges, err := n.store.GetDHCPRanges(id)
	if err != nil {
		log.Printf("failed to get DHCP ranges: %v", err)
		http.Error(w, "failed to get DHCP ranges", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ranges); err != nil {
		log.Printf("failed to encode DHCP ranges: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateDHCPRangeHandler creates a DHCP range for a network
func (n *Networks) CreateDHCPRangeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "network ID is required", http.StatusBadRequest)
		return
	}

	networkID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid network ID", http.StatusBadRequest)
		return
	}

	var dhcpRange domain.DHCPRange
	if err := json.NewDecoder(r.Body).Decode(&dhcpRange); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	dhcpRange.NetworkID = networkID
	if dhcpRange.StartIP == "" {
		http.Error(w, "start_ip is required", http.StatusBadRequest)
		return
	}
	if dhcpRange.EndIP == "" {
		http.Error(w, "end_ip is required", http.StatusBadRequest)
		return
	}

	createdRange, err := n.store.CreateDHCPRange(dhcpRange)
	if err != nil {
		log.Printf("failed to create DHCP range: %v", err)
		http.Error(w, "failed to create DHCP range", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdRange); err != nil {
		log.Printf("failed to encode created DHCP range: %v", err)
	}
}

// DeleteDHCPRangeHandler deletes a DHCP range
func (n *Networks) DeleteDHCPRangeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "rangeId")
	if idStr == "" {
		http.Error(w, "DHCP range ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid DHCP range ID", http.StatusBadRequest)
		return
	}

	if err := n.store.DeleteDHCPRange(id); err != nil {
		log.Printf("failed to delete DHCP range: %v", err)
		http.Error(w, "failed to delete DHCP range", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
