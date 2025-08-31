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
	r.Route("/api/v0/machines", func(r chi.Router) {
		r.Get("/", a.listMachinesHandler)
		r.Post("/", a.createMachineHandler)
		r.Get("/{id}", a.getMachineHandler)
		r.Delete("/{id}", a.deleteMachineHandler)
		r.Get("/name/{name}", a.getMachineByNameHandler)
		r.Get("/ipv4/{ipv4}", a.getMachineByIPv4Handler)
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

// Machine request/response types
type CreateMachineRequest struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	IPv4     string `json:"ipv4"`
}

type MachineResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	IPv4     string `json:"ipv4"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// listMachinesHandler handles GET /api/v0/machines
func (a *API) listMachinesHandler(w http.ResponseWriter, r *http.Request) {
	machines, err := a.ds.ListMachines()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]MachineResponse, len(machines))
	for i, machine := range machines {
		response[i] = MachineResponse{
			ID:       machine.ID,
			Name:     machine.Name,
			Hostname: machine.Hostname,
			IPv4:     machine.IPv4,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode machines response: %v", err)
	}
}

// createMachineHandler handles POST /api/v0/machines
func (a *API) createMachineHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Printf("[ERROR] missing required fields in machine creation: %+v\n", req)
		return
	}

	// Validate IPv4 format
	if net.ParseIP(req.IPv4) == nil || !isIPv4(req.IPv4) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid IPv4 address format"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		fmt.Printf("[ERROR] invalid IPv4 address: %s\n", req.IPv4)
		return
	}

	// Check for duplicate IPv4
	existing, _ := a.ds.GetMachineByIPv4(req.IPv4)
	if existing != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "A machine with this IPv4 address already exists"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		fmt.Printf("[ERROR] duplicate IPv4 address: %s\n", req.IPv4)
		return
	}

	machine := datastore.Machine{
		Name:     req.Name,
		Hostname: req.Hostname,
		IPv4:     req.IPv4,
	}

	created, err := a.ds.CreateMachine(machine)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		fmt.Printf("[ERROR] failed to create machine: %v\n", err)
		return
	}

	response := MachineResponse{
		ID:       created.ID,
		Name:     created.Name,
		Hostname: created.Hostname,
		IPv4:     created.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode machines response: %v", err)
	}
}

// getMachineHandler handles GET /api/v0/machines/{id}
func (a *API) getMachineHandler(w http.ResponseWriter, r *http.Request) {
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

	machine, err := a.ds.GetMachine(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)}); err != nil {
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

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode machine by name response: %v", err)
	}
}

// deleteMachineHandler handles DELETE /api/v0/machines/{id}
func (a *API) deleteMachineHandler(w http.ResponseWriter, r *http.Request) {
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

	err = a.ds.DeleteMachine(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to delete machine: %v", err)}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getMachineByNameHandler handles GET /api/v0/machines/name/{name}
func (a *API) getMachineByNameHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	machine, err := a.ds.GetMachineByName(name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)}); err != nil {
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

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode network response: %v", err)
	}
}

// getMachineByIPv4Handler handles GET /api/v0/machines/ipv4/{ipv4}
func (a *API) getMachineByIPv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := chi.URLParam(r, "ipv4")

	machine, err := a.ds.GetMachineByIPv4(ipv4)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)}); err != nil {
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

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode ssh keys response: %v", err)
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
func isIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}
