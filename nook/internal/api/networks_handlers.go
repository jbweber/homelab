package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NetworksStore defines the datastore interface for network handlers (placeholder for future expansion)
type NetworksStore interface {
	// Add methods as needed for real network data
	CreateNetwork(name string) error
	DeleteNetwork(name string) error
}

// Networks groups network handlers for testability
type Networks struct {
	store NetworksStore
}

func NewNetworks(store NetworksStore) *Networks {
	return &Networks{store: store}
}

func (n *Networks) NetworksHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("[networks endpoint placeholder]")); err != nil {
		log.Printf("failed to write networks endpoint placeholder: %v", err)
	}
}

func (n *Networks) CreateNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if n.store == nil {
		http.Error(w, "networks store not implemented", http.StatusNotImplemented)
		return
	}
	if err := n.store.CreateNetwork(req.Name); err != nil {
		http.Error(w, "failed to create network", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(`{"message": "network created"}`)); err != nil {
		log.Printf("failed to write create network response: %v", err)
	}
}

func (n *Networks) DeleteNetworkHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "network name is required", http.StatusBadRequest)
		return
	}

	if n.store == nil {
		http.Error(w, "networks store not implemented", http.StatusNotImplemented)
		return
	}

	if err := n.store.DeleteNetwork(name); err != nil {
		http.Error(w, "failed to delete network", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
