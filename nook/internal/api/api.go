package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"log"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/repository"
)

// API holds repository dependencies for clean data access
type API struct {
	machineRepo   repository.MachineRepository
	sshKeyRepo    repository.SSHKeyRepository
	networkRepo   repository.NetworkRepository
	dhcpRangeRepo repository.DHCPRangeRepository
	ipLeaseRepo   repository.IPLeaseRepository
}

// ListMachines implements MachinesStore interface
func (a *API) ListMachines() ([]Machine, error) {
	machines, err := a.machineRepo.FindAll(context.Background())
	if err != nil {
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	var result []Machine
	for _, m := range machines {
		result = append(result, Machine{
			ID:        m.ID,
			Name:      m.Name,
			Hostname:  m.Hostname,
			IPv4:      m.IPv4,
			NetworkID: m.NetworkID,
		})
	}
	return result, nil
}

// CreateMachine implements MachinesStore interface
func (a *API) CreateMachine(m Machine) (Machine, error) {
	// Convert api.Machine to domain.Machine
	domainMachine := domain.Machine{
		ID:        m.ID,
		Name:      m.Name,
		Hostname:  m.Hostname,
		IPv4:      m.IPv4,
		NetworkID: m.NetworkID,
	}
	saved, err := a.machineRepo.Save(context.Background(), domainMachine)
	if err != nil {
		return Machine{}, err
	}

	// If network_id is provided but no IPv4, allocate IP after machine creation
	if m.NetworkID != nil && m.IPv4 == "" {
		lease, err := a.ipLeaseRepo.AllocateIPAddress(context.Background(), saved.ID, *m.NetworkID)
		if err != nil {
			// If IP allocation fails, delete the machine and return error
			if deleteErr := a.machineRepo.DeleteByID(context.Background(), saved.ID); deleteErr != nil {
				fmt.Printf("Warning: failed to delete machine after IP allocation failure: %v\n", deleteErr)
			}
			return Machine{}, fmt.Errorf("failed to allocate IP address: %w", err)
		}
		// Update the machine with the allocated IP
		saved.IPv4 = lease.IPAddress
		updated, err := a.machineRepo.Save(context.Background(), saved)
		if err != nil {
			// If update fails, deallocate the IP
			if deallocErr := a.ipLeaseRepo.DeallocateIPAddress(context.Background(), saved.ID, *m.NetworkID); deallocErr != nil {
				fmt.Printf("Warning: failed to deallocate IP after machine update failure: %v\n", deallocErr)
			}
			return Machine{}, err
		}
		saved = updated
	}

	// Convert back to api.Machine
	return Machine{
		ID:        saved.ID,
		Name:      saved.Name,
		Hostname:  saved.Hostname,
		IPv4:      saved.IPv4,
		NetworkID: saved.NetworkID,
	}, nil
}

// GetMachine implements MachinesStore interface
func (a *API) GetMachine(id int64) (*Machine, error) {
	machine, err := a.machineRepo.FindByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      machine.IPv4,
		NetworkID: machine.NetworkID,
	}, nil
}

// DeleteMachine implements MachinesStore interface
func (a *API) DeleteMachine(id int64) error {
	// First, get the machine to check if it has a network-allocated IP
	machine, err := a.machineRepo.FindByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // Machine doesn't exist, consider it deleted
		}
		return err
	}

	// If the machine has a network_id and IPv4, deallocate the IP
	if machine.NetworkID != nil && machine.IPv4 != "" {
		if deallocErr := a.ipLeaseRepo.DeallocateIPAddress(context.Background(), machine.ID, *machine.NetworkID); deallocErr != nil {
			// Log the error but don't fail the deletion
			fmt.Printf("Warning: failed to deallocate IP for machine %d: %v\n", machine.ID, deallocErr)
		}
	}

	// Delete the machine
	return a.machineRepo.DeleteByID(context.Background(), id)
}

// GetMachineByName implements MachinesStore interface
func (a *API) GetMachineByName(name string) (*Machine, error) {
	machine, err := a.machineRepo.FindByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      machine.IPv4,
		NetworkID: machine.NetworkID,
	}, nil
}

// AllocateIPAddress implements MachinesStore interface
func (a *API) AllocateIPAddress(machineID, networkID int64) (string, error) {
	lease, err := a.ipLeaseRepo.AllocateIPAddress(context.Background(), machineID, networkID)
	if err != nil {
		return "", err
	}
	return lease.IPAddress, nil
}

// DeallocateIPAddress implements MachinesStore interface
func (a *API) DeallocateIPAddress(machineID, networkID int64) error {
	return a.ipLeaseRepo.DeallocateIPAddress(context.Background(), machineID, networkID)
}

// CreateNetwork implements NetworksStore interface
func (a *API) CreateNetwork(network domain.Network) (domain.Network, error) {
	return a.networkRepo.Save(context.Background(), network)
}

// GetNetwork implements NetworksStore interface
func (a *API) GetNetwork(id int64) (domain.Network, error) {
	return a.networkRepo.FindByID(context.Background(), id)
}

// GetNetworkByName implements NetworksStore interface
func (a *API) GetNetworkByName(name string) (domain.Network, error) {
	return a.networkRepo.FindByName(context.Background(), name)
}

// ListNetworks implements NetworksStore interface
func (a *API) ListNetworks() ([]domain.Network, error) {
	return a.networkRepo.FindAll(context.Background())
}

// UpdateNetwork implements NetworksStore interface
func (a *API) UpdateNetwork(network domain.Network) (domain.Network, error) {
	return a.networkRepo.Save(context.Background(), network)
}

// DeleteNetwork implements NetworksStore interface
func (a *API) DeleteNetwork(id int64) error {
	return a.networkRepo.DeleteByID(context.Background(), id)
}

// GetDHCPRanges implements NetworksStore interface
func (a *API) GetDHCPRanges(networkID int64) ([]domain.DHCPRange, error) {
	return a.networkRepo.GetDHCPRanges(context.Background(), networkID)
}

// CreateDHCPRange implements NetworksStore interface
func (a *API) CreateDHCPRange(dhcpRange domain.DHCPRange) (domain.DHCPRange, error) {
	return a.dhcpRangeRepo.Save(context.Background(), dhcpRange)
}

// DeleteDHCPRange implements NetworksStore interface
func (a *API) DeleteDHCPRange(id int64) error {
	return a.dhcpRangeRepo.DeleteByID(context.Background(), id)
}

// ListAllSSHKeys implements SSHKeysStore interface
func (a *API) ListAllSSHKeys() ([]SSHKey, error) {
	keys, err := a.sshKeyRepo.FindAll(context.Background())
	if err != nil {
		return nil, err
	}
	// Convert domain.SSHKey to api.SSHKey
	var result []SSHKey
	for _, k := range keys {
		result = append(result, SSHKey{
			ID:        k.ID,
			MachineID: k.MachineID,
			KeyText:   k.KeyText,
		})
	}
	return result, nil
}

// CreateSSHKey implements SSHKeysStore interface
func (a *API) CreateSSHKey(machineID int64, keyText string) (*SSHKey, error) {
	key, err := a.sshKeyRepo.CreateForMachine(context.Background(), machineID, keyText)
	if err != nil {
		return nil, err
	}
	// Convert domain.SSHKey to api.SSHKey
	return &SSHKey{
		ID:        key.ID,
		MachineID: key.MachineID,
		KeyText:   key.KeyText,
	}, nil
}

// DeleteSSHKey implements SSHKeysStore interface
func (a *API) DeleteSSHKey(id int64) error {
	return a.sshKeyRepo.DeleteByID(context.Background(), id)
}

// GetMachineByIPv4 implements MetaDataStore interface
func (a *API) GetMachineByIPv4(ipv4 string) (*Machine, error) {
	machine, err := a.machineRepo.FindByIPv4(context.Background(), ipv4)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}, nil
}

// extractClientIP extracts the client IP from the request, preferring X-Forwarded-For header
// over RemoteAddr. Returns an error if the IP cannot be parsed.
func extractClientIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("unable to parse remote address: %w", err)
		}
	}
	return ip, nil
}

// updateMachineHandler handles PATCH /api/v0/machines/{id}.
//
// Request: JSON body with fields "name", "hostname", "ipv4".
// Validates ID, required fields, and IPv4 format. Returns 400 for invalid input, 404 if not found, 500 for DB errors.
// Response: 200 OK with updated machine, or error JSON.
func (a *API) updateMachineHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid machine ID"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	var req CreateMachineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	// Validate required fields
	if req.Name == "" || req.Hostname == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Name and Hostname are required"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	// Validate IPv4 format if provided
	if req.IPv4 != nil && *req.IPv4 != "" {
		if net.ParseIP(*req.IPv4) == nil || !isIPv4(*req.IPv4) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid IPv4 address format"}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			return
		}
	}

	// Check if machine exists
	machine, err := a.machineRepo.FindByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine not found"}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get machine"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	// Update machine fields
	machine.Name = req.Name
	machine.Hostname = req.Hostname
	if req.IPv4 != nil && *req.IPv4 != "" {
		machine.IPv4 = *req.IPv4
	}

	updated, err := a.machineRepo.Save(context.Background(), machine)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update machine"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	response := MachineResponse{
		ID:       updated.ID,
		Name:     updated.Name,
		Hostname: updated.Hostname,
		IPv4:     &updated.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode update response: %v", err)
	}
}

// NewAPI creates a new API instance with repositories initialized from the datastore
func NewAPI(db *sql.DB) *API {
	return &API{
		machineRepo:   repository.NewMachineRepository(db),
		sshKeyRepo:    repository.NewSSHKeyRepository(db),
		networkRepo:   repository.NewNetworkRepository(db),
		dhcpRangeRepo: repository.NewDHCPRangeRepository(db),
		ipLeaseRepo:   repository.NewIPLeaseRepository(db),
	}
}

// NewAPIWithRepos creates a new API instance with specific repositories for testing
func NewAPIWithRepos(machineRepo repository.MachineRepository, sshKeyRepo repository.SSHKeyRepository, networkRepo repository.NetworkRepository, dhcpRangeRepo repository.DHCPRangeRepository, ipLeaseRepo repository.IPLeaseRepository) *API {
	return &API{
		machineRepo:   machineRepo,
		sshKeyRepo:    sshKeyRepo,
		networkRepo:   networkRepo,
		dhcpRangeRepo: dhcpRangeRepo,
		ipLeaseRepo:   ipLeaseRepo,
	}
}

// RegisterRoutes registers all API endpoints to the given chi router.
func (a *API) RegisterRoutes(r chi.Router) {

	// Metadata endpoints group
	meta := NewMetaData(a)
	r.Get("/meta-data", meta.NoCloudMetaDataHandler)
	r.Get("/user-data", a.noCloudUserDataHandler)
	r.Get("/vendor-data", a.noCloudVendorDataHandler)

	// Machines endpoints group
	machines := NewMachines(a)
	r.Route("/api/v0/machines", func(r chi.Router) {
		r.Get("/", machines.ListMachinesHandler)
		r.Post("/", machines.CreateMachineHandler)
		r.Get("/{id}", machines.GetMachineHandler)
		r.Delete("/{id}", machines.DeleteMachineHandler)
		r.Get("/name/{name}", machines.GetMachineByNameHandler)
		r.Get("/ipv4/{ipv4}", machines.GetMachineByIPv4Handler)
		r.Patch("/{id}", a.updateMachineHandler)
	})

	// Networks endpoints group
	networks := NewNetworks(a)
	r.Route("/api/v0/networks", func(r chi.Router) {
		r.Get("/", networks.NetworksHandler)
		r.Post("/", networks.CreateNetworkHandler)
		r.Get("/{id}", networks.GetNetworkHandler)
		r.Patch("/{id}", networks.UpdateNetworkHandler)
		r.Delete("/{id}", networks.DeleteNetworkHandler)
		r.Get("/{id}/dhcp", networks.GetNetworkDHCPRangesHandler)
		r.Post("/{id}/dhcp", networks.CreateDHCPRangeHandler)
		r.Delete("/dhcp/{rangeId}", networks.DeleteDHCPRangeHandler)
	})

	// SSH keys endpoints group - registered by the SSH keys module
	RegisterSSHKeysRoutes(r, a)
}
func (a *API) noCloudUserDataHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := extractClientIP(r)
	if err != nil {
		log.Printf("failed to extract client IP: %v", err)
		http.Error(w, "unable to determine client IP address", http.StatusBadRequest)
		return
	}

	// Validate IP format
	if net.ParseIP(ip) == nil {
		log.Printf("invalid IP address format: %s", ip)
		http.Error(w, "invalid IP address format", http.StatusBadRequest)
		return
	}

	machine, err := a.machineRepo.FindByIPv4(context.Background(), ip)
	if err != nil {
		log.Printf("failed to lookup machine by IP %s: %v", ip, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if machine.ID == 0 {
		log.Printf("machine not found for IP: %s", ip)
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	// Get SSH keys
	keys, err := a.sshKeyRepo.FindByMachineID(context.Background(), machine.ID)
	if err != nil {
		log.Printf("failed to list SSH keys for machine %d: %v", machine.ID, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Build user-data
	userData := fmt.Sprintf(`#cloud-config
hostname: %s
manage_etc_hosts: true
`, machine.Hostname)

	if len(keys) > 0 {
		userData += "ssh_authorized_keys:\n"
		for _, key := range keys {
			userData += fmt.Sprintf("  - %s\n", key.KeyText)
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(userData)); err != nil {
		log.Printf("failed to write user data: %v", err)
	}
}

// noCloudVendorDataHandler serves NoCloud-compatible vendor-data
func (a *API) noCloudVendorDataHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve empty vendor-data
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("")); err != nil {
		log.Printf("failed to write vendor data: %v", err)
	}
}

// noCloudNetworkConfigHandler serves NoCloud-compatible network-config
func (a *API) noCloudNetworkConfigHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve a basic network config
	networkConfig := `version: 2
ethernets:
  eth0:
	dhcp4: true
`
	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(networkConfig)); err != nil {
		log.Printf("failed to write network config: %v", err)
	}
}

// isIPv4 checks if a string is a valid IPv4 address
