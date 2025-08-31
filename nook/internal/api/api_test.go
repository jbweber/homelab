package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAPI(t *testing.T) *chi.Mux {
	// Create test datastore
	testDS, err := datastore.New(testutil.NewTestDSN("TestAPI"))
	if err != nil {
		t.Fatalf("Failed to create test datastore: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	api := NewAPI(testDS)
	api.RegisterRoutes(r)

	return r
}

func TestListMachines_Empty(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))

	assert.Len(t, response, 0)
}

func TestCreateMachine(t *testing.T) {
	r := setupTestAPI(t)

	reqBody := CreateMachineRequest{
		Name: "test-machine",
		IPv4: "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))

	assert.Equal(t, reqBody.Name, response.Name)
	assert.Equal(t, reqBody.IPv4, response.IPv4)
	assert.NotZero(t, response.ID)
}

func TestCreateMachine_InvalidJSON(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMachine_MissingFields(t *testing.T) {
	r := setupTestAPI(t)

	reqBody := CreateMachineRequest{Name: "test"} // Missing IPv4
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMachine_NotFound(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMachine_InvalidID(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
