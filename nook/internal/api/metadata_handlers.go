package api

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

// MetaDataStore describes the datastore methods needed for metadata endpoints.
type MetaDataStore interface {
	GetMachineByIPv4(ipv4 string) (*datastore.Machine, error)
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
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	machine, err := m.store.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
		return
	}

	instanceID := fmt.Sprintf("iid-%08d", machine.ID)
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
	if _, err := w.Write([]byte(meta)); err != nil {
		log.Printf("failed to write meta-data: %v", err)
	}
}

// MetaDataDirectoryHandler serves a directory listing for /meta-data/ (refactored for MetaData).
func (m *MetaData) MetaDataDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	// For now, serve a static directory listing
	dir := "instance-id\nhostname\nlocal-hostname\nlocal-ipv4\npublic-hostname\nsecurity-groups\n"
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(dir)); err != nil {
		log.Printf("failed to write meta-data directory: %v", err)
	}
}

// MetaDataKeyHandler serves individual metadata keys for /meta-data/{key} (refactored for MetaData).
func (m *MetaData) MetaDataKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to parse remote address", http.StatusBadRequest)
			return
		}
	}

	machine, err := m.store.GetMachineByIPv4(ip)
	if err != nil {
		http.Error(w, "failed to lookup machine by IP", http.StatusInternalServerError)
		return
	}
	if machine == nil {
		http.Error(w, "machine not found for IP", http.StatusNotFound)
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
		http.Error(w, "unknown metadata key", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(value + "\n")); err != nil {
		log.Printf("failed to write meta-data key: %v", err)
	}
}
