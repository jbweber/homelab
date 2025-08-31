package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

// MachinesStore defines the datastore interface for machine handlers
type MachinesStore interface {
	ListMachines() ([]datastore.Machine, error)
	CreateMachine(datastore.Machine) (datastore.Machine, error)
	GetMachine(id int64) (*datastore.Machine, error)
	DeleteMachine(id int64) error
	GetMachineByName(name string) (*datastore.Machine, error)
	GetMachineByIPv4(ipv4 string) (*datastore.Machine, error)
}

// Machines groups machine handlers for testability
type Machines struct {
	store MachinesStore
}

func NewMachines(store MachinesStore) *Machines {
	return &Machines{store: store}
}

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

func (m *Machines) ListMachinesHandler(w http.ResponseWriter, r *http.Request) {
	machines, err := m.store.ListMachines()
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

func (m *Machines) CreateMachineHandler(w http.ResponseWriter, r *http.Request) {
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
	existing, _ := m.store.GetMachineByIPv4(req.IPv4)
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

	created, err := m.store.CreateMachine(machine)
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

func (m *Machines) GetMachineHandler(w http.ResponseWriter, r *http.Request) {
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

	machine, err := m.store.GetMachine(id)
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

func (m *Machines) DeleteMachineHandler(w http.ResponseWriter, r *http.Request) {
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

	err = m.store.DeleteMachine(id)
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

func (m *Machines) GetMachineByNameHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	machine, err := m.store.GetMachineByName(name)
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

func (m *Machines) GetMachineByIPv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := chi.URLParam(r, "ipv4")

	machine, err := m.store.GetMachineByIPv4(ipv4)
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

func isIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}
