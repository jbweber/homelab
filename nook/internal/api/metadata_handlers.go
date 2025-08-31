package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler for /latest/meta-data/ directory listing
func (a *API) metaDataDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	// For now, return a static list of keys. Later, generate dynamically from machine model.
	keys := []string{
		"instance-id",
		"hostname",
		"local-ipv4",
		"public-keys/",
		"public-hostname",
		"placement/",
		"security-groups",
		"user-data",
	}
	w.WriteHeader(http.StatusOK)
	for _, k := range keys {
		if _, err := fmt.Fprintln(w, k); err != nil {
			log.Printf("failed to write metadata key: %v", err)
		}
	}
}

// Handler stub for /latest/meta-data/{key}
func (a *API) metaDataKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	// For now, return placeholder. Next steps: map to actual machine fields.
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "[meta-data/%s placeholder]", key); err != nil {
		log.Printf("failed to write metadata key placeholder: %v", err)
	}
}
