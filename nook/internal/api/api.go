package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

// API holds the datastore dependency
type API struct {
	ds *datastore.Datastore
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
	json.NewEncoder(w).Encode(doc)
}

// NewAPI creates a new API instance with the given datastore
func NewAPI(ds *datastore.Datastore) *API {
	return &API{ds: ds}
}

// RegisterRoutes registers all API endpoints to the given chi router.
func (a *API) RegisterRoutes(r chi.Router) {
	r.Get("/meta-data", a.noCloudMetaDataHandler)
	r.Get("/meta-data/", a.metaDataDirectoryHandler)
	r.Get("/meta-data/{key}", a.metaDataKeyHandler)
	r.Get("/user-data", a.noCloudUserDataHandler)
	r.Get("/vendor-data", a.noCloudVendorDataHandler)
	r.Get("/network-config", a.noCloudNetworkConfigHandler)

	r.Get("/api/v0/machines", a.listMachinesHandler)
	r.Post("/api/v0/machines", a.createMachineHandler)
	r.Get("/api/v0/machines/{id}", a.getMachineHandler)
	r.Delete("/api/v0/machines/{id}", a.deleteMachineHandler)
	r.Get("/api/v0/machines/name/{name}", a.getMachineByNameHandler)
	r.Get("/api/v0/machines/ipv4/{ipv4}", a.getMachineByIPv4Handler)
	r.Get("/api/v0/networks", networksHandler)
	r.Get("/api/v0/ssh-keys", sshKeysHandler)
	r.Get("/2021-01-03/dynamic/instance-identity/document", a.instanceIdentityDocumentHandler)
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
	w.Write([]byte(userData))
}

// noCloudVendorDataHandler serves NoCloud-compatible vendor-data
func (a *API) noCloudVendorDataHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve empty vendor-data
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
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
	w.Write([]byte(networkConfig))
}

// noCloudMetaDataHandler serves NoCloud-compatible metadata based on requestor IP
func (a *API) noCloudMetaDataHandler(w http.ResponseWriter, r *http.Request) {
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

	// Format instance-id as iid-XXXXXXXX (8-digit zero-padded ID)
	instanceID := fmt.Sprintf("iid-%08d", machine.ID)

	// Compose NoCloud metadata (YAML)
	meta := fmt.Sprintf(
		"instance-id: %s\n"+
			"hostname: %s\n"+
			"local-hostname: %s\n"+
			"local-ipv4: %s\n"+
			"public-hostname: %s\n"+
			"security-groups: default\n",
		instanceID,
		machine.Hostname,
		machine.Hostname,
		machine.IPv4,
		machine.Hostname,
	)

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(meta))
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
	json.NewEncoder(w).Encode(response)
}

// createMachineHandler handles POST /api/v0/machines
func (a *API) createMachineHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateMachineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Hostname == "" || req.IPv4 == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Name, Hostname, and IPv4 are required"})
		fmt.Printf("[ERROR] missing required fields in machine creation: %+v\n", req)
		return
	}

	// Validate IPv4 format
	if net.ParseIP(req.IPv4) == nil || !isIPv4(req.IPv4) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid IPv4 address format"})
		fmt.Printf("[ERROR] invalid IPv4 address: %s\n", req.IPv4)
		return
	}

	// Check for duplicate IPv4
	existing, _ := a.ds.GetMachineByIPv4(req.IPv4)
	if existing != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "A machine with this IPv4 address already exists"})
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
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)})
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
	json.NewEncoder(w).Encode(response)
}

// getMachineHandler handles GET /api/v0/machines/{id}
func (a *API) getMachineHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid machine ID"})
		return
	}

	machine, err := a.ds.GetMachine(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)})
		return
	}

	if machine == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine not found"})
		return
	}

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deleteMachineHandler handles DELETE /api/v0/machines/{id}
func (a *API) deleteMachineHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid machine ID"})
		return
	}

	err = a.ds.DeleteMachine(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to delete machine: %v", err)})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getMachineByNameHandler handles GET /api/v0/machines/name/{name}
func (a *API) getMachineByNameHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine name is required"})
		return
	}

	machine, err := a.ds.GetMachineByName(name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)})
		return
	}

	if machine == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine not found"})
		return
	}

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getMachineByIPv4Handler handles GET /api/v0/machines/ipv4/{ipv4}
func (a *API) getMachineByIPv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := chi.URLParam(r, "ipv4")
	if ipv4 == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "IPv4 address is required"})
		return
	}

	machine, err := a.ds.GetMachineByIPv4(ipv4)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to get machine: %v", err)})
		return
	}

	if machine == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Machine not found"})
		return
	}

	response := MachineResponse{
		ID:       machine.ID,
		Name:     machine.Name,
		Hostname: machine.Hostname,
		IPv4:     machine.IPv4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func networksHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[networks endpoint placeholder]"))
}

func sshKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[ssh-keys endpoint placeholder]"))
}

// isIPv4 checks if a string is a valid IPv4 address
func isIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}
