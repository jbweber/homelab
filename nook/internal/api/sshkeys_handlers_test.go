package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/migrations"
	"github.com/jbweber/homelab/nook/internal/repository"
	"github.com/jbweber/homelab/nook/internal/testutil"
	_ "modernc.org/sqlite"
)

// mockSSHKeysStore implements SSHKeysStore for testing error cases
type mockSSHKeysStore struct {
	getMachineError  error
	listSSHKeysError error
}

func (m *mockSSHKeysStore) ListAllSSHKeys() ([]SSHKey, error) {
	return nil, nil
}

func (m *mockSSHKeysStore) GetMachineByIPv4(ip string) (*Machine, error) {
	if m.getMachineError != nil {
		return nil, m.getMachineError
	}
	return &Machine{ID: 1, Name: "test", Hostname: "test", IPv4: ip}, nil
}

func (m *mockSSHKeysStore) ListSSHKeys(machineID int64) ([]SSHKey, error) {
	if m.listSSHKeysError != nil {
		return nil, m.listSSHKeysError
	}
	return []SSHKey{}, nil
}

func (m *mockSSHKeysStore) CreateSSHKey(machineID int64, keyText string) (*SSHKey, error) {
	return &SSHKey{ID: 1, MachineID: machineID, KeyText: keyText}, nil
}

func (m *mockSSHKeysStore) DeleteSSHKey(id int64) error {
	return nil
}

// setupSSHKeysTestRouter creates a test router with only SSH keys routes registered
func setupSSHKeysTestRouter(t *testing.T, store SSHKeysStore) *chi.Mux {
	r := chi.NewRouter()
	RegisterSSHKeysRoutes(r, store)
	return r
}

// setupSSHKeysTestAPI creates a test router with full API but focused on SSH keys testing
func setupSSHKeysTestAPI(t *testing.T) (*chi.Mux, *sql.DB) {
	// Create test database with migrations
	db, cleanup := testutil.SetupTestDB(t, "TestSSHKeys")
	t.Cleanup(cleanup)

	// Run migrations
	migrator := migrations.NewMigrator(db)
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	if err := migrator.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	r := chi.NewRouter()
	api := NewAPI(db)
	api.RegisterRoutes(r)

	return r, db
}

func TestSSHKeysHandler(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Optionally check body if expected
}

func TestSSHKeysHandler_Placeholder(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestPublicKeysHandler_Success(t *testing.T) {
	r, db := setupSSHKeysTestAPI(t)

	// Create repositories for test data setup
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create a machine
	reqBody := CreateMachineRequest{
		Name:     "ssh-machine",
		Hostname: "ssh-host",
		IPv4:     "192.168.1.170",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.170:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	// Insert SSH key for this machine using repository
	key := domain.SSHKey{
		MachineID: created.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey1",
	}
	_, err := sshKeyRepo.Save(ctx, key)
	require.NoError(t, err)

	key2 := domain.SSHKey{
		MachineID: created.ID,
		KeyText:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2",
	}
	_, err = sshKeyRepo.Save(ctx, key2)
	require.NoError(t, err)

	// Use X-Forwarded-For to trigger publicKeysHandler
	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.170")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	bodyStr := w2.Body.String()
	assert.Contains(t, bodyStr, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey1")
	assert.Contains(t, bodyStr, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2")
}

func TestPublicKeysHandler_NotFound(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99") // Not in test DB
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeysHandler_LookupError(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Should be 404 due to not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeysHandler_MalformedRemoteAddr(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	// No X-Forwarded-For, and malformed RemoteAddr
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unable to parse remote address")
}

func TestPublicKeysHandler_ListSSHKeysError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		listSSHKeysError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list SSH keys")
}

func TestPublicKeysHandler_GetMachineError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		getMachineError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to lookup machine by IP")
}

func TestPublicKeyByIdxHandler_Success(t *testing.T) {
	r, db := setupSSHKeysTestAPI(t)

	// Create repositories for test data setup
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create machine and keys
	reqBody := CreateMachineRequest{
		Name:     "idx-machine",
		Hostname: "idx-host",
		IPv4:     "192.168.1.180",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.180:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	key1 := domain.SSHKey{
		MachineID: created.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCidxkey1",
	}
	_, err := sshKeyRepo.Save(ctx, key1)
	require.NoError(t, err)

	key2 := domain.SSHKey{
		MachineID: created.ID,
		KeyText:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIidxkey2",
	}
	_, err = sshKeyRepo.Save(ctx, key2)
	require.NoError(t, err)

	// Get key at idx 1
	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/1", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.180")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIidxkey2")
}

func TestPublicKeyByIdxHandler_NotFound(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)

	// No machine for IP
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")

	// Create machine with unique name/IP, no keys
	reqBody := CreateMachineRequest{
		Name:     "idx-machine-unique",
		Hostname: "idx-host-unique",
		IPv4:     "192.168.1.182",
	}
	body, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.RemoteAddr = "192.168.1.182:12345"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&created))

	// Out of range idx
	req3 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req3.Header.Set("X-Forwarded-For", "192.168.1.182")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
	assert.Contains(t, w3.Body.String(), "key index out of range")
}

func TestPublicKeyByIdxHandler_InvalidIndex(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/invalid", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.170")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid key index")
}

func TestPublicKeyByIdxHandler_MalformedRemoteAddr(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	// No X-Forwarded-For, and malformed RemoteAddr
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unable to parse remote address")
}

func TestPublicKeyByIdxHandler_GetMachineError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		getMachineError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to lookup machine by IP")
}

func TestPublicKeyByIdxHandler_ListSSHKeysError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		listSSHKeysError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list SSH keys")
}

func TestPublicKeyOpenSSHHandler_Success(t *testing.T) {
	r, db := setupSSHKeysTestAPI(t)

	// Create repositories for test data setup
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	ctx := context.Background()

	// Create machine and keys
	reqBody := CreateMachineRequest{
		Name:     "openssh-machine",
		Hostname: "openssh-host",
		IPv4:     "192.168.1.190",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.190:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	key := domain.SSHKey{
		MachineID: created.ID,
		KeyText:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCopensshkey1",
	}
	_, err := sshKeyRepo.Save(ctx, key)
	require.NoError(t, err)

	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.190")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCopensshkey1")
}

func TestPublicKeyOpenSSHHandler_NotFound(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeyOpenSSHHandler_LookupError(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeyOpenSSHHandler_MalformedRemoteAddr(t *testing.T) {
	r, _ := setupSSHKeysTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	// No X-Forwarded-For, and malformed RemoteAddr
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unable to parse remote address")
}

func TestPublicKeyOpenSSHHandler_GetMachineError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		getMachineError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to lookup machine by IP")
}

func TestPublicKeyOpenSSHHandler_ListSSHKeysError(t *testing.T) {
	mockStore := &mockSSHKeysStore{
		listSSHKeysError: errors.New("database error"),
	}
	r := setupSSHKeysTestRouter(t, mockStore)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list SSH keys")
}
