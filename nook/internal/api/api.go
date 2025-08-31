package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"log"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

// ...existing code...

// API holds the datastore dependency
type API struct {
	ds *datastore.Datastore
}

// publicKeyByIdxHandler serves a specific SSH public key by index for the requesting machine.
// GET /2021-01-03/meta-data/public-keys/{idx}
// Returns: text/plain, key line or error.
func (a *API) publicKeyByIdxHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 {
		http.Error(w, "invalid key index", http.StatusBadRequest)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	machine, err := a.ds.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}

	keys, err := a.ds.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}
	if idx >= len(keys) {
		http.Error(w, "key index out of range", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprintln(w, keys[idx].KeyText); err != nil {
		log.Printf("failed to write public key by idx: %v", err)
	}
}

// publicKeyOpenSSHHandler serves the OpenSSH-formatted key for a specific index.
// GET /2021-01-03/meta-data/public-keys/{idx}/openssh-key
// Returns: text/plain, OpenSSH key or error.
func (a *API) publicKeyOpenSSHHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 {
		http.Error(w, "invalid key index", http.StatusBadRequest)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	machine, err := a.ds.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}

	keys, err := a.ds.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}
	if idx >= len(keys) {
		http.Error(w, "key index out of range", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprintln(w, keys[idx].KeyText); err != nil {
		log.Printf("failed to write public key OpenSSH: %v", err)
	}
}

// publicKeysHandler serves a list of SSH public keys for the requesting machine (cloud-init format).
//
// GET /2021-01-03/meta-data/public-keys
// Returns: text/plain, one key per line, or error.
func (a *API) publicKeysHandler(w http.ResponseWriter, r *http.Request) {
	// Extract requestor IP, prefer X-Forwarded-For if present
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	// Lookup machine by IPv4
	machine, err := a.ds.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}

	// Fetch SSH keys for this machine
	keys, err := a.ds.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}

	// Format: one key per line (cloud-init expects this)
	w.Header().Set("Content-Type", "text/plain")
	for _, k := range keys {
		if _, err := fmt.Fprintln(w, k.KeyText); err != nil {
			log.Printf("failed to write public key: %v", err)
		}
	}
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
	machine, err := a.ds.GetMachine(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get machine"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}
	if machine == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine not found"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	// Update machine fields
	machine.Name = req.Name
	machine.Hostname = req.Hostname
	machine.IPv4 = req.IPv4

	updated, err := a.ds.UpdateMachine(*machine)
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

// instanceIdentityDocumentHandler serves EC2-compatible instance identity document
func (a *API) instanceIdentityDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Extract requestor IP, prefer X-Forwarded-For if present
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Printf("[ERROR] unable to parse remote address: %v\n", err)
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	// Lookup machine by IPv4
	machine, err := a.ds.GetMachineByIPv4(ip)
	if err != nil {
		fmt.Printf("[ERROR] failed to lookup machine by IP %s: %v\n", ip, err)
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		fmt.Printf("[ERROR] machine not found for IP %s\n", ip)
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}

	// Compose EC2-compatible instance identity document
	doc := map[string]interface{}{
		"instanceId":       fmt.Sprintf("iid-%08d", machine.ID),
		"privateIp":        machine.IPv4,
		"hostname":         machine.Hostname,
		"region":           "local-nocloud",        // NoCloud does not have regions
		"availabilityZone": "local-nocloud-az",     // NoCloud does not have AZs
		"architecture":     "x86_64",               // Default, could be made dynamic
		"imageId":          "n/a",                  // Not tracked
		"accountId":        "n/a",                  // Not tracked
		"instanceType":     "nocloud",              // Default type
		"pendingTime":      "2025-08-31T00:00:00Z", // Static for now
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		log.Printf("failed to encode instance identity document: %v", err)
	}
}

// NewAPI creates a new API instance with the given datastore
func NewAPI(ds *datastore.Datastore) *API {
	return &API{ds: ds}
}

// RegisterRoutes registers all API endpoints to the given chi router.
func (a *API) RegisterRoutes(r chi.Router) {

	// Metadata endpoints group
	meta := NewMetaData(a.ds)
	r.Get("/meta-data", meta.NoCloudMetaDataHandler)
	r.Get("/meta-data/", meta.MetaDataDirectoryHandler)
	r.Get("/meta-data/{key}", meta.MetaDataKeyHandler)
	r.Get("/user-data", a.noCloudUserDataHandler)
	r.Get("/vendor-data", a.noCloudVendorDataHandler)
	r.Get("/network-config", a.noCloudNetworkConfigHandler)

	// Machines endpoints group
	machines := NewMachines(a.ds)
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
	r.Route("/api/v0/networks", func(r chi.Router) {
		r.Get("/", networksHandler)
	})

	// SSH keys endpoints group
	r.Route("/api/v0/ssh-keys", func(r chi.Router) {
		r.Get("/", a.sshKeysHandler)
	})

	// EC2-compatible and public-keys endpoints group
	r.Route("/2021-01-03", func(r chi.Router) {
		r.Get("/dynamic/instance-identity/document", a.instanceIdentityDocumentHandler)
		r.Route("/meta-data/public-keys", func(r chi.Router) {
			r.Get("/", a.publicKeysHandler)
			r.Get("/{idx}", a.publicKeyByIdxHandler)
			r.Get("/{idx}/openssh-key", a.publicKeyOpenSSHHandler)
		})
	})
}

// noCloudUserDataHandler serves NoCloud-compatible user-data
func (a *API) noCloudUserDataHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve a static cloud-init user-data script
	fmt.Println("[DEBUG] noCloudUserDataHandler called")
	userData := `#cloud-config
hostname: default-host
manage_etc_hosts: true
`
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

func networksHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("[networks endpoint placeholder]")); err != nil {
		log.Printf("failed to write networks endpoint placeholder: %v", err)
	}
}

// sshKeysHandler serves a list of all SSH keys in the system as JSON.
func (a *API) sshKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := a.ds.ListAllSSHKeys()
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}

	type SSHKeyResponse struct {
		ID        int64  `json:"id"`
		MachineID int64  `json:"machine_id"`
		KeyText   string `json:"key_text"`
	}
	resp := make([]SSHKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = SSHKeyResponse{
			ID:        k.ID,
			MachineID: k.MachineID,
			KeyText:   k.KeyText,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode ssh keys response: %v", err)
	}
}

// isIPv4 checks if a string is a valid IPv4 address
