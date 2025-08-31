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

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []MachineResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected empty list, got %d machines", len(response))
	}
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

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response MachineResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Name != reqBody.Name {
		t.Errorf("Expected name %s, got %s", reqBody.Name, response.Name)
	}
	if response.IPv4 != reqBody.IPv4 {
		t.Errorf("Expected IPv4 %s, got %s", reqBody.IPv4, response.IPv4)
	}
	if response.ID == 0 {
		t.Error("Expected non-zero ID")
	}
}

func TestCreateMachine_InvalidJSON(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateMachine_MissingFields(t *testing.T) {
	r := setupTestAPI(t)

	reqBody := CreateMachineRequest{Name: "test"} // Missing IPv4
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetMachine_NotFound(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetMachine_InvalidID(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
