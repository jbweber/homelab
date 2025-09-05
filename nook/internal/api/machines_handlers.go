package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Machine represents a virtual machine in the system
type Machine struct {
	ID        int64  // Unique identifier
	Name      string // Machine name
	Hostname  string // Machine hostname
	IPv4      string // IPv4 address
	NetworkID *int64 // Network ID for dynamic IP allocation (optional)
}

// MachinesStore defines the datastore interface for machine handlers
type MachinesStore interface {
	ListMachines() ([]Machine, error)
	CreateMachine(Machine) (Machine, error)
	GetMachine(id int64) (*Machine, error)
	DeleteMachine(id int64) error
	GetMachineByName(name string) (*Machine, error)
	GetMachineByIPv4(ipv4 string) (*Machine, error)
	AllocateIPAddress(machineID, networkID int64) (string, error)
	DeallocateIPAddress(machineID, networkID int64) error
}

// Machines groups machine handlers for testability
type Machines struct {
	store MachinesStore
}

func NewMachines(store MachinesStore) *Machines {
	return &Machines{store: store}
}

type CreateMachineRequest struct {
	Name      string  `json:"name"`
	Hostname  string  `json:"hostname"`
	IPv4      *string `json:"ipv4,omitempty"`       // Optional: for static IP assignment
	NetworkID *int64  `json:"network_id,omitempty"` // Optional: if provided, allocate IP from this network
}

type MachineResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Hostname  string  `json:"hostname"`
	IPv4      *string `json:"ipv4,omitempty"`
	NetworkID *int64  `json:"network_id,omitempty"`
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
			ID:        machine.ID,
			Name:      machine.Name,
			Hostname:  machine.Hostname,
			IPv4:      &machine.IPv4,
			NetworkID: machine.NetworkID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode machines response: %v", err)
	}
}

func (m *Machines) CreateMachineHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateMachineRequest
	var allocatedIP string
	var machine Machine
	var created Machine
	var response MachineResponse
	var err error

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
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
		fmt.Printf("[ERROR] missing required fields in machine creation: %+v\n", req)
		return
	}

	// Handle different IP assignment scenarios
	if req.NetworkID != nil && req.IPv4 != nil {
		// Both network_id and ipv4 provided - this is invalid
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Cannot specify both network_id and ipv4. Choose one or neither."}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		fmt.Printf("[ERROR] both network_id and ipv4 specified: %+v\n", req)
		return
	} else if req.NetworkID != nil {
		// Network-based IP allocation - IP will be allocated by the store
		machine = Machine{
			Name:      req.Name,
			Hostname:  req.Hostname,
			IPv4:      "", // Will be allocated by the store
			NetworkID: req.NetworkID,
		}

		created, err = m.store.CreateMachine(machine)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			fmt.Printf("[ERROR] failed to create machine: %v\n", err)
			return
		}
	} else if req.IPv4 != nil {
		// Static IP provided
		allocatedIP = *req.IPv4
		// Validate static IP format
		if net.ParseIP(allocatedIP) == nil || !isIPv4(allocatedIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid IPv4 address format"}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			fmt.Printf("[ERROR] invalid IPv4 address: %s\n", allocatedIP)
			return
		}
		// Check for duplicate static IP
		existing, _ := m.store.GetMachineByIPv4(allocatedIP)
		if existing != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "A machine with this IPv4 address already exists"}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			fmt.Printf("[ERROR] duplicate IPv4 address: %s\n", allocatedIP)
			return
		}

		// Create machine with static IP
		machine = Machine{
			Name:      req.Name,
			Hostname:  req.Hostname,
			IPv4:      allocatedIP,
			NetworkID: nil, // Static IPs don't use networks
		}

		created, err = m.store.CreateMachine(machine)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			fmt.Printf("[ERROR] failed to create machine: %v\n", err)
			return
		}
	} else {
		// No IP assignment - create machine with empty IP
		machine = Machine{
			Name:      req.Name,
			Hostname:  req.Hostname,
			IPv4:      "",
			NetworkID: nil,
		}

		created, err = m.store.CreateMachine(machine)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create machine: %v", err)}); err != nil {
				log.Printf("failed to encode error response: %v", err)
			}
			fmt.Printf("[ERROR] failed to create machine: %v\n", err)
			return
		}
	}

	// Prepare response
	response = MachineResponse{
		ID:        created.ID,
		Name:      created.Name,
		Hostname:  created.Hostname,
		IPv4:      &created.IPv4,
		NetworkID: created.NetworkID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode machine response: %v", err)
	}

	fmt.Printf("Created machine with ID: %d\n", created.ID)
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
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      &machine.IPv4,
		NetworkID: machine.NetworkID,
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
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      &machine.IPv4,
		NetworkID: machine.NetworkID,
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
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      &machine.IPv4,
		NetworkID: machine.NetworkID,
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

// UpdateMachineHandler handles PATCH /api/v0/machines/{id}.
//
// Request: JSON body with fields "name", "hostname", "ipv4".
// Validates ID, required fields, and IPv4 format. Returns 400 for invalid input, 404 if not found, 500 for DB errors.
// Response: 200 OK with updated machine, or error JSON.
func (m *Machines) UpdateMachineHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get the machine via store interface
	machine, err := m.store.GetMachine(id)
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
	if req.IPv4 != nil && *req.IPv4 != "" {
		machine.IPv4 = *req.IPv4
	}

	// Save via store interface
	updated, err := m.store.CreateMachine(*machine) // CreateMachine handles both create and update
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update machine"}); err != nil {
			log.Printf("failed to encode error response: %v", err)
		}
		return
	}

	response := MachineResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		Hostname:  updated.Hostname,
		IPv4:      &updated.IPv4,
		NetworkID: updated.NetworkID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode update response: %v", err)
	}
}
