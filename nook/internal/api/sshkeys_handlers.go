package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

// SSHKeysStore defines the datastore interface for SSH key handlers
type SSHKeysStore interface {
	ListAllSSHKeys() ([]datastore.SSHKey, error)
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode ssh keys response: %v", err)
	}
}
