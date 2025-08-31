package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

// API holds the datastore dependency
type API struct {
	ds *datastore.Datastore
}

// NewAPI creates a new API instance with the given datastore
func NewAPI(ds *datastore.Datastore) *API {
	return &API{ds: ds}
}

// RegisterRoutes registers all API endpoints to the given chi router.
func (a *API) RegisterRoutes(r chi.Router) {
	r.Get("/2021-01-03/dynamic/instance-identity/document", a.instanceIdentityHandler)
	r.Get("/2021-01-03/meta-data/public-keys", publicKeysHandler)
	r.Get("/2021-01-03/meta-data/public-keys/{idx}", publicKeyIndexHandler)
	r.Get("/2021-01-03/meta-data/public-keys/{idx}/openssh-key", publicKeyOpenSSHHandler)
	r.Get("/latest/api/token", apiTokenHandler)
	r.Get("/meta-data", metaDataHandler)
	r.Get("/user-data", userDataHandler)
	r.Get("/vendor-data", vendorDataHandler)

	// Machine management endpoints
	r.Get("/api/v0/machines", a.listMachinesHandler)
	r.Post("/api/v0/machines", a.createMachineHandler)
	r.Get("/api/v0/machines/{id}", a.getMachineHandler)
	r.Delete("/api/v0/machines/{id}", a.deleteMachineHandler)
	r.Get("/api/v0/machines/name/{name}", a.getMachineByNameHandler)
	r.Get("/api/v0/machines/ipv4/{ipv4}", a.getMachineByIPv4Handler)

	r.Get("/api/v0/networks", networksHandler)
	r.Get("/api/v0/ssh-keys", sshKeysHandler)
}

func (a *API) instanceIdentityHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[instance-identity document placeholder]"))
}

func publicKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[public-keys placeholder]"))
}

func publicKeyIndexHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "Invalid key index", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[public-keys/%d placeholder]", idx)
}

func publicKeyOpenSSHHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "Invalid key index", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[public-keys/%d/openssh-key placeholder]", idx)
}

func apiTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[api token placeholder]"))
}

func metaDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[meta-data placeholder]"))
}

func userDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[user-data placeholder]"))
}

func vendorDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[vendor-data placeholder]"))
}

// Machine request/response types
type CreateMachineRequest struct {
	Name string `json:"name"`
	IPv4 string `json:"ipv4"`
}

type MachineResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	IPv4 string `json:"ipv4"`
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
			ID:   machine.ID,
			Name: machine.Name,
			IPv4: machine.IPv4,
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

	if req.Name == "" || req.IPv4 == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Name and IPv4 are required"})
		return
	}

	machine := datastore.Machine{
		Name: req.Name,
		IPv4: req.IPv4,
	}

	created, err := a.ds.CreateMachine(machine)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)})
		return
	}

	response := MachineResponse{
		ID:   created.ID,
		Name: created.Name,
		IPv4: created.IPv4,
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
		ID:   machine.ID,
		Name: machine.Name,
		IPv4: machine.IPv4,
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
		ID:   machine.ID,
		Name: machine.Name,
		IPv4: machine.IPv4,
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
		ID:   machine.ID,
		Name: machine.Name,
		IPv4: machine.IPv4,
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
