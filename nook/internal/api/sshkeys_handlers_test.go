package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockSSHKeysStore struct {
	sshKeys []SSHKey
	err     error
}

func (m *mockSSHKeysStore) ListAllSSHKeys() ([]SSHKey, error) {
	return m.sshKeys, m.err
}

func (m *mockSSHKeysStore) GetMachineByIPv4(ip string) (*Machine, error) {
	return nil, nil // Not used in SSH key handlers
}

func (m *mockSSHKeysStore) CreateSSHKey(machineID int64, keyText string) (*SSHKey, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := &SSHKey{
		ID:        int64(len(m.sshKeys) + 1),
		MachineID: machineID,
		KeyText:   keyText,
	}
	m.sshKeys = append(m.sshKeys, *key)
	return key, nil
}

func (m *mockSSHKeysStore) DeleteSSHKey(id int64) error {
	if m.err != nil {
		return m.err
	}
	// Find and remove the key
	for i, key := range m.sshKeys {
		if key.ID == id {
			m.sshKeys = append(m.sshKeys[:i], m.sshKeys[i+1:]...)
			return nil
		}
	}
	return nil // Key not found, but don't error
}

func TestSSHKeys_SSHKeysHandler_Empty(t *testing.T) {
	store := &mockSSHKeysStore{sshKeys: []SSHKey{}}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()

	sshKeys.SSHKeysHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := w.Body.String()
	expected := "[]\n"
	if body != expected {
		t.Errorf("Expected body %q, got %q", expected, body)
	}
}

func TestSSHKeys_SSHKeysHandler_Success(t *testing.T) {
	store := &mockSSHKeysStore{
		sshKeys: []SSHKey{
			{ID: 1, MachineID: 1, KeyText: "ssh-rsa AAAAB3NzaC1yc2E..."},
			{ID: 2, MachineID: 1, KeyText: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."},
		},
	}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()

	sshKeys.SSHKeysHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []SSHKeyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 SSH keys, got %d", len(response))
	}

	if response[0].ID != 1 || response[0].MachineID != 1 {
		t.Errorf("Unexpected first SSH key: %+v", response[0])
	}
}

func TestSSHKeys_SSHKeysHandler_Error(t *testing.T) {
	store := &mockSSHKeysStore{err: errors.New("database error")}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()

	sshKeys.SSHKeysHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestSSHKeys_CreateSSHKeyHandler_Success(t *testing.T) {
	store := &mockSSHKeysStore{sshKeys: []SSHKey{}}
	sshKeys := NewSSHKeys(store)

	requestBody := map[string]interface{}{
		"machine_id": 1,
		"key_text":   "ssh-rsa AAAAB3NzaC1yc2E...",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v0/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	sshKeys.CreateSSHKeyHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response SSHKeyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.MachineID != 1 {
		t.Errorf("Expected MachineID 1, got %d", response.MachineID)
	}
	if response.KeyText != "ssh-rsa AAAAB3NzaC1yc2E..." {
		t.Errorf("Expected key text to match, got %s", response.KeyText)
	}
}

func TestSSHKeys_CreateSSHKeyHandler_InvalidJSON(t *testing.T) {
	store := &mockSSHKeysStore{sshKeys: []SSHKey{}}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("POST", "/api/v0/ssh-keys", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	sshKeys.CreateSSHKeyHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSSHKeys_CreateSSHKeyHandler_MissingKeyText(t *testing.T) {
	store := &mockSSHKeysStore{sshKeys: []SSHKey{}}
	sshKeys := NewSSHKeys(store)

	requestBody := map[string]interface{}{
		"machine_id": 1,
		"key_text":   "",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v0/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	sshKeys.CreateSSHKeyHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSSHKeys_CreateSSHKeyHandler_StoreError(t *testing.T) {
	store := &mockSSHKeysStore{err: errors.New("store error")}
	sshKeys := NewSSHKeys(store)

	requestBody := map[string]interface{}{
		"machine_id": 1,
		"key_text":   "ssh-rsa AAAAB3NzaC1yc2E...",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v0/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	sshKeys.CreateSSHKeyHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestSSHKeys_DeleteSSHKeyHandler_Success(t *testing.T) {
	store := &mockSSHKeysStore{
		sshKeys: []SSHKey{
			{ID: 1, MachineID: 1, KeyText: "ssh-rsa AAAAB3NzaC1yc2E..."},
		},
	}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("DELETE", "/api/v0/ssh-keys/1", nil)
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	sshKeys.DeleteSSHKeyHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestSSHKeys_DeleteSSHKeyHandler_InvalidID(t *testing.T) {
	store := &mockSSHKeysStore{sshKeys: []SSHKey{}}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("DELETE", "/api/v0/ssh-keys/invalid", nil)
	w := httptest.NewRecorder()

	// Set up chi context with invalid ID
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	sshKeys.DeleteSSHKeyHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSSHKeys_DeleteSSHKeyHandler_StoreError(t *testing.T) {
	store := &mockSSHKeysStore{err: errors.New("store error")}
	sshKeys := NewSSHKeys(store)

	req := httptest.NewRequest("DELETE", "/api/v0/ssh-keys/1", nil)
	w := httptest.NewRecorder()

	// Set up chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	sshKeys.DeleteSSHKeyHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewSSHKeys(t *testing.T) {
	store := &mockSSHKeysStore{}
	sshKeys := NewSSHKeys(store)

	if sshKeys == nil {
		t.Error("Expected NewSSHKeys to return non-nil")
	}
	if sshKeys.store != store {
		t.Error("Expected store to be set correctly")
	}
}

func TestRegisterSSHKeysRoutes(t *testing.T) {
	store := &mockSSHKeysStore{}
	r := chi.NewRouter()

	// This should not panic and should register routes
	RegisterSSHKeysRoutes(r, store)

	// Test that routes were registered by making requests
	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v0/ssh-keys"},
		{"POST", "/api/v0/ssh-keys"},
		{"DELETE", "/api/v0/ssh-keys/1"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// We expect the routes to be registered (not 404)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s %s not registered", tc.method, tc.path)
		}
	}
}
