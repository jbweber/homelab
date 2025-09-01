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
	machineRepo repository.MachineRepository
	sshKeyRepo  repository.SSHKeyRepository
}

// machineStoreAdapter adapts MachinesStore interface
type machineStoreAdapter struct {
	repo repository.MachineRepository
}

func (a *machineStoreAdapter) ListMachines() ([]Machine, error) {
	machines, err := a.repo.FindAll(context.Background())
	if err != nil {
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	var result []Machine
	for _, m := range machines {
		result = append(result, Machine{
			ID:       m.ID,
			Name:     m.Name,
			Hostname: m.Hostname,
			IPv4:     m.IPv4,
		})
	}
	return result, nil
}

func (a *machineStoreAdapter) CreateMachine(m Machine) (Machine, error) {
	// Convert api.Machine to domain.Machine
	domainMachine := domain.Machine{
		ID:       m.ID,
		Name:     m.Name,
		Hostname: m.Hostname,
		IPv4:     m.IPv4,
	}
	saved, err := a.repo.Save(context.Background(), domainMachine)
	if err != nil {
		return Machine{}, err
	}
	// Convert back to api.Machine
	return Machine{
		ID:       saved.ID,
		Name:     saved.Name,
		Hostname: saved.Hostname,
		IPv4:     saved.IPv4,
	}, nil
}

func (a *machineStoreAdapter) GetMachine(id int64) (*Machine, error) {
	machine, err := a.repo.FindByID(context.Background(), id)
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

func (a *machineStoreAdapter) DeleteMachine(id int64) error {
	return a.repo.DeleteByID(context.Background(), id)
}

func (a *machineStoreAdapter) GetMachineByName(name string) (*Machine, error) {
	machine, err := a.repo.FindByName(context.Background(), name)
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

func (a *machineStoreAdapter) GetMachineByIPv4(ipv4 string) (*Machine, error) {
	machine, err := a.repo.FindByIPv4(context.Background(), ipv4)
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

// sshKeysStoreAdapter adapts repositories to SSHKeysStore interface
type sshKeysStoreAdapter struct {
	sshKeyRepo  repository.SSHKeyRepository
	machineRepo repository.MachineRepository
}

func (a *sshKeysStoreAdapter) ListAllSSHKeys() ([]SSHKey, error) {
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

func (a *sshKeysStoreAdapter) GetMachineByIPv4(ip string) (*Machine, error) {
	machine, err := a.machineRepo.FindByIPv4(context.Background(), ip)
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

func (a *sshKeysStoreAdapter) CreateSSHKey(machineID int64, keyText string) (*SSHKey, error) {
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

func (a *sshKeysStoreAdapter) DeleteSSHKey(id int64) error {
	return a.sshKeyRepo.DeleteByID(context.Background(), id)
}

// simpleNetworksStore is a simple implementation of NetworksStore
type simpleNetworksStore struct{}

func (s *simpleNetworksStore) CreateNetwork(name string) error {
	// For now, just log the network creation
	log.Printf("Network created: %s", name)
	return nil
}

func (s *simpleNetworksStore) DeleteNetwork(name string) error {
	// For now, just log the network deletion
	log.Printf("Network deleted: %s", name)
	return nil
}

// metaDataStoreAdapter adapts MachineRepository to MetaDataStore interface
type metaDataStoreAdapter struct {
	machineRepo repository.MachineRepository
}

func (a *metaDataStoreAdapter) GetMachineByIPv4(ipv4 string) (*Machine, error) {
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
	if req.Name == "" || req.Hostname == "" || req.IPv4 == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Name, Hostname, and IPv4 are required"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	// Validate IPv4 format
	if net.ParseIP(req.IPv4) == nil || !isIPv4(req.IPv4) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid IPv4 address format"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
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
	machine.IPv4 = req.IPv4

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
		IPv4:     updated.IPv4,
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
		machineRepo: repository.NewMachineRepository(db),
		sshKeyRepo:  repository.NewSSHKeyRepository(db),
	}
}

// RegisterRoutes registers all API endpoints to the given chi router.
func (a *API) RegisterRoutes(r chi.Router) {

	// Metadata endpoints group
	metaAdapter := &metaDataStoreAdapter{machineRepo: a.machineRepo}
	meta := NewMetaData(metaAdapter)
	r.Get("/meta-data", meta.NoCloudMetaDataHandler)
	r.Get("/user-data", a.noCloudUserDataHandler)
	r.Get("/vendor-data", a.noCloudVendorDataHandler)

	// Machines endpoints group
	machineAdapter := &machineStoreAdapter{repo: a.machineRepo}
	machines := NewMachines(machineAdapter)
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
	networks := NewNetworks(&simpleNetworksStore{})
	r.Route("/api/v0/networks", func(r chi.Router) {
		r.Get("/", networks.NetworksHandler)
		r.Post("/", networks.CreateNetworkHandler)
		r.Delete("/{name}", networks.DeleteNetworkHandler)
	})

	// SSH keys endpoints group - registered by the SSH keys module
	sshKeysAdapter := &sshKeysStoreAdapter{
		sshKeyRepo:  a.sshKeyRepo,
		machineRepo: a.machineRepo,
	}
	RegisterSSHKeysRoutes(r, sshKeysAdapter)
}
func (a *API) noCloudUserDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] noCloudUserDataHandler called")
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

	log.Printf("Serving user-data for machine %s (IP: %s), keys count: %d", machine.Name, ip, len(keys))
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
