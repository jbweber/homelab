package api

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// MetaDataStore describes the datastore methods needed for metadata endpoints.
type MetaDataStore interface {
	GetMachineByIPv4(ipv4 string) (*Machine, error)
	// Add more methods here as needed for other metadata endpoints
}

// MetaData holds dependencies and handler methods for /meta-data* endpoints.
type MetaData struct {
	store MetaDataStore
}

// NewMetaData creates a new MetaData instance with the given store.
func NewMetaData(store MetaDataStore) *MetaData {
	return &MetaData{store: store}
}

// NoCloudMetaDataHandler serves NoCloud-compatible metadata based on requestor IP (refactored for MetaData).
func (m *MetaData) NoCloudMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := extractClientIP(r)
	if err != nil {
		log.Printf("failed to extract client IP: %v", err)
		http.Error(w, "unable to determine client IP address", http.StatusBadRequest)
		return
	}

	// Validate IP format
	if net.ParseIP(ip) == nil {
		log.Printf("invalid IP address format: %s", ip)
		http.Error(w, "invalid IP address format", http.StatusBadRequest)
		return
	}

	machine, err := m.store.GetMachineByIPv4(ip)
	if err != nil {
		log.Printf("failed to lookup machine by IP %s: %v", ip, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		log.Printf("machine not found for IP: %s", ip)
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	instanceID := fmt.Sprintf("iid-%08d", machine.ID)
	// Use proper YAML format for NoCloud compatibility
	meta := fmt.Sprintf(`instance-id: %s
hostname: %s
local-hostname: %s
local-ipv4: %s
public-hostname: %s
security-groups: default
`,
		instanceID,
		machine.Hostname,
		machine.Hostname,
		machine.IPv4,
		machine.Hostname,
	)

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(meta)); err != nil {
		log.Printf("failed to write meta-data response: %v", err)
	}
}

// MetaDataDirectoryHandler serves a directory listing for /meta-data/ (refactored for MetaData).
func (m *MetaData) MetaDataDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	// NoCloud metadata directory listing
	dir := `instance-id
hostname
local-hostname
local-ipv4
public-hostname
security-groups
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(dir)); err != nil {
		log.Printf("failed to write meta-data directory response: %v", err)
	}
}

// MetaDataKeyHandler serves individual metadata keys for /meta-data/{key} (refactored for MetaData).
func (m *MetaData) MetaDataKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		log.Printf("empty metadata key requested")
		http.Error(w, "metadata key is required", http.StatusBadRequest)
		return
	}

	ip, err := extractClientIP(r)
	if err != nil {
		log.Printf("failed to extract client IP for key %s: %v", key, err)
		http.Error(w, "unable to determine client IP address", http.StatusBadRequest)
		return
	}

	// Validate IP format
	if net.ParseIP(ip) == nil {
		log.Printf("invalid IP address format for key %s: %s", key, ip)
		http.Error(w, "invalid IP address format", http.StatusBadRequest)
		return
	}

	machine, err := m.store.GetMachineByIPv4(ip)
	if err != nil {
		log.Printf("failed to lookup machine by IP %s for key %s: %v", ip, key, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		log.Printf("machine not found for IP %s requesting key %s", ip, key)
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	var value string
	switch key {
	case "instance-id":
		value = fmt.Sprintf("iid-%08d", machine.ID)
	case "hostname", "local-hostname", "public-hostname":
		value = machine.Hostname
	case "local-ipv4":
		value = machine.IPv4
	case "security-groups":
		value = "default"
	default:
		log.Printf("unknown metadata key requested: %s", key)
		http.Error(w, "unknown metadata key", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(value + "\n")); err != nil {
		log.Printf("failed to write meta-data key %s response: %v", key, err)
	}
}
