package api

import (
	"fmt"
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
		fmt.Fprintln(w, k)
	}
}

// Handler stub for /latest/meta-data/{key}
func (a *API) metaDataKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	// For now, return placeholder. Next steps: map to actual machine fields.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[meta-data/%s placeholder]", key)
}
