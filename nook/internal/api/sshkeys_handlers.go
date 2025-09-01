package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// SSHKey represents an SSH public key associated with a machine
type SSHKey struct {
	ID        int64  // Unique identifier
	MachineID int64  // Foreign key to Machine
	KeyText   string // Public SSH key text
}

// SSHKeysStore defines the datastore interface for SSH key handlers
type SSHKeysStore interface {
	ListAllSSHKeys() ([]SSHKey, error)
	GetMachineByIPv4(ip string) (*Machine, error)
	CreateSSHKey(machineID int64, keyText string) (*SSHKey, error)
	DeleteSSHKey(id int64) error
}

// SSHKeys groups SSH key handlers for testability
type SSHKeys struct {
	store SSHKeysStore
}

func NewSSHKeys(store SSHKeysStore) *SSHKeys {
	return &SSHKeys{store: store}
}

// SSHKeyResponse represents the JSON response for SSH key operations
type SSHKeyResponse struct {
	ID        int64  `json:"id"`
	MachineID int64  `json:"machine_id"`
	KeyText   string `json:"key_text"`
}

func (s *SSHKeys) SSHKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := s.store.ListAllSSHKeys()
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
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
	w.WriteHeader(http.StatusOK)
	if len(resp) == 0 {
		if _, err := w.Write([]byte("[]\n")); err != nil {
			log.Printf("failed to write empty ssh keys array: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode ssh keys response: %v", err)
	}
}

func (s *SSHKeys) CreateSSHKeyHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MachineID int64  `json:"machine_id"`
		KeyText   string `json:"key_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.KeyText == "" {
		http.Error(w, "key_text is required", http.StatusBadRequest)
		return
	}
	key, err := s.store.CreateSSHKey(req.MachineID, req.KeyText)
	if err != nil {
		http.Error(w, "failed to create SSH key", http.StatusInternalServerError)
		return
	}
	resp := SSHKeyResponse{
		ID:        key.ID,
		MachineID: key.MachineID,
		KeyText:   key.KeyText,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode create ssh key response: %v", err)
	}
}

func (s *SSHKeys) DeleteSSHKeyHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid SSH key ID", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteSSHKey(id)
	if err != nil {
		http.Error(w, "failed to delete SSH key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterSSHKeysRoutes registers all SSH key related routes on the provided router.
// This function encapsulates all SSH key route registration logic, making it easier
// to test and maintain SSH key functionality independently.
func RegisterSSHKeysRoutes(r chi.Router, store SSHKeysStore) {
	sshKeys := NewSSHKeys(store)

	// API v0 SSH keys endpoints
	r.Route("/api/v0/ssh-keys", func(r chi.Router) {
		r.Get("/", sshKeys.SSHKeysHandler)
		r.Post("/", sshKeys.CreateSSHKeyHandler)
		r.Delete("/{id}", sshKeys.DeleteSSHKeyHandler)
	})

	// EC2-compatible public keys endpoints - removed as not needed for nocloud
}
