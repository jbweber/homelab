package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all API endpoints to the given chi router.
func RegisterRoutes(r chi.Router) {
	r.Get("/2021-01-03/dynamic/instance-identity/document", instanceIdentityHandler)
	r.Get("/2021-01-03/meta-data/public-keys", publicKeysHandler)
	r.Get("/2021-01-03/meta-data/public-keys/{idx}", publicKeyIndexHandler)
	r.Get("/2021-01-03/meta-data/public-keys/{idx}/openssh-key", publicKeyOpenSSHHandler)
	r.Get("/latest/api/token", apiTokenHandler)
	r.Get("/meta-data", metaDataHandler)
	r.Get("/user-data", userDataHandler)
	r.Get("/vendor-data", vendorDataHandler)
	r.Get("/api/v0/machines", machinesHandler)
	r.Get("/api/v0/networks", networksHandler)
	r.Get("/api/v0/ssh-keys", sshKeysHandler)
}

func instanceIdentityHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[instance-identity document placeholder]"))
}

func publicKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[public-keys placeholder]"))
}

func publicKeyIndexHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "Invalid key index", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[public-keys/%d placeholder]", idx)
}

func publicKeyOpenSSHHandler(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "Invalid key index", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[public-keys/%d/openssh-key placeholder]", idx)
}

func apiTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[api token placeholder]"))
}

func metaDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[meta-data placeholder]"))
}

func userDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[user-data placeholder]"))
}

func vendorDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[vendor-data placeholder]"))
}

func machinesHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[machines endpoint placeholder]"))
}

func networksHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[networks endpoint placeholder]"))
}

func sshKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[ssh-keys endpoint placeholder]"))
}
