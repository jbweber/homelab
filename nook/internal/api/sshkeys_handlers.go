package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

// SSHKeysStore defines the datastore interface for SSH key handlers
type SSHKeysStore interface {
	ListAllSSHKeys() ([]datastore.SSHKey, error)
	GetMachineByIPv4(ip string) (*datastore.Machine, error)
	ListSSHKeys(machineID int64) ([]datastore.SSHKey, error)
}

// PublicKeysHandler handles /2021-01-03/meta-data/public-keys
func (s *SSHKeys) PublicKeysHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := extractClientIP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	machine, err := s.store.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}
	keys, err := s.store.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	for _, k := range keys {
		if _, err := fmt.Fprintln(w, k.KeyText); err != nil {
			log.Printf("failed to write public key: %v", err)
		}
	}
}

// PublicKeyByIdxHandler handles /2021-01-03/meta-data/public-keys/{idx}
func (s *SSHKeys) PublicKeyByIdxHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid key index", http.StatusBadRequest)
		return
	}
	ip, err := extractClientIP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	machine, err := s.store.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}
	keys, err := s.store.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}
	if idx < 0 || idx >= len(keys) {
		http.Error(w, "key index out of range", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprintln(w, keys[idx].KeyText); err != nil {
		log.Printf("failed to write public key by idx: %v", err)
	}
}

// PublicKeyOpenSSHHandler handles /2021-01-03/meta-data/public-keys/{idx}/openssh-key
func (s *SSHKeys) PublicKeyOpenSSHHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid key index", http.StatusBadRequest)
		return
	}
	ip, err := extractClientIP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	machine, err := s.store.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}
	keys, err := s.store.ListSSHKeys(machine.ID)
	if err != nil {
		http.Error(w, "failed to list SSH keys", http.StatusInternalServerError)
		return
	}
	if idx < 0 || idx >= len(keys) {
		http.Error(w, "key index out of range", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprintln(w, keys[idx].KeyText); err != nil {
		log.Printf("failed to write public key OpenSSH: %v", err)
	}
}

// SSHKeys groups SSH key handlers for testability
type SSHKeys struct {
	store SSHKeysStore
}

func NewSSHKeys(store SSHKeysStore) *SSHKeys {
	return &SSHKeys{store: store}
}

func (s *SSHKeys) SSHKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := s.store.ListAllSSHKeys()
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

// RegisterSSHKeysRoutes registers all SSH key related routes on the provided router.
// This function encapsulates all SSH key route registration logic, making it easier
// to test and maintain SSH key functionality independently.
func RegisterSSHKeysRoutes(r chi.Router, store SSHKeysStore) {
	sshKeys := NewSSHKeys(store)

	// API v0 SSH keys endpoints
	r.Route("/api/v0/ssh-keys", func(r chi.Router) {
		r.Get("/", sshKeys.SSHKeysHandler)
	})

	// EC2-compatible public keys endpoints
	r.Route("/2021-01-03/meta-data/public-keys", func(r chi.Router) {
		r.Get("/", sshKeys.PublicKeysHandler)
		r.Get("/{idx}", sshKeys.PublicKeyByIdxHandler)
		r.Get("/{idx}/openssh-key", sshKeys.PublicKeyOpenSSHHandler)
	})
}
