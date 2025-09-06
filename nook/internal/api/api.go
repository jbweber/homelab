package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		r.Patch("/{id}", machines.UpdateMachineHandler)
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
	var userData string

	if err != nil || machine.ID == 0 {
		// Machine not found - provide basic user data without machine-specific config
		log.Printf("machine not found for IP %s, providing basic user data", ip)
		userData = `#cloud-config
manage_etc_hosts: true
`
	} else {
		// Machine found - get SSH keys and build full user data
		keys, err := a.sshKeyRepo.FindByMachineID(context.Background(), machine.ID)
		if err != nil {
			log.Printf("failed to list SSH keys for machine %d: %v", machine.ID, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		userData = fmt.Sprintf(`#cloud-config
hostname: %s
manage_etc_hosts: true
`, machine.Hostname)

		if len(keys) > 0 {
			userData += "ssh_authorized_keys:\n"
			for _, key := range keys {
				userData += fmt.Sprintf("  - %s\n", key.KeyText)
			}
		}
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(userData)); err != nil {
		log.Printf("failed to write user data: %v", err)
	}
}

// noCloudVendorDataHandler serves NoCloud-compatible vendor-data
func (a *API) noCloudVendorDataHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve empty vendor-data
	w.Header().Set("Content-Type", "text/yaml")
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
