package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/migrations"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// setupIntegrationTestAPI creates a test API server with a real database
func setupIntegrationTestAPI(t *testing.T) (*API, func()) {
	t.Helper()

	// Setup test database
	db, cleanup := testutil.SetupTestDB(t, "integration_test")

	// Run migrations
	migrator := migrations.NewMigrator(db)
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	if err := migrator.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create API
	api := NewAPI(db)

	return api, cleanup
}

// TestMetaDataIntegration_Basic tests basic metadata functionality
func TestMetaDataIntegration_Basic(t *testing.T) {
	api, cleanup := setupIntegrationTestAPI(t)
	defer cleanup()

	r := chi.NewRouter()
	api.RegisterRoutes(r)

	// Test 1: Get metadata directory
	t.Run("MetaDataDirectory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

		body := w.Body.String()
		assert.Contains(t, body, "instance-id")
		assert.Contains(t, body, "hostname")
		assert.Contains(t, body, "local-ipv4")
	})

	// Test 2: Test with non-existent IP
	t.Run("NonExistentIP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data", nil)
		req.RemoteAddr = "203.0.113.99:12345" // Non-existent IP
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "machine not found")
	})

	// Test 3: Test with invalid IP format
	t.Run("InvalidIP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data", nil)
		req.Header.Set("X-Forwarded-For", "invalid-ip-format")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid IP address format")
	})

	// Test 4: Test with invalid IP format
	t.Run("InvalidIP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data", nil)
		req.Header.Set("X-Forwarded-For", "invalid-ip-format")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid IP address format")
	})
}

// TestMetaDataIntegration_WithMachine tests metadata with an actual machine in the database
func TestMetaDataIntegration_WithMachine(t *testing.T) {
	// Setup test database
	db, cleanup := testutil.SetupTestDB(t, "with_machine_test")
	defer cleanup()

	// Run migrations
	migrator := migrations.NewMigrator(db)
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}
	if err := migrator.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Insert a test machine
	_, err := db.Exec("INSERT INTO machines (name, hostname, ipv4) VALUES (?, ?, ?)",
		"test-machine", "test-host", "192.168.1.50")
	if err != nil {
		t.Fatalf("Failed to insert test machine: %v", err)
	}

	// Create API with this database
	api := NewAPI(db)
	r := chi.NewRouter()
	api.RegisterRoutes(r)

	// Test with existing machine
	t.Run("ExistingMachine", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/yaml; charset=utf-8", w.Header().Get("Content-Type"))

		body := w.Body.String()
		assert.Contains(t, body, "hostname: test-host")
		assert.Contains(t, body, "local-ipv4: 192.168.1.50")
	})

	// Test unknown key with existing machine
	t.Run("UnknownKeyWithMachine", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/meta-data/unknown-key", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "unknown metadata key")
	})
}

// TestMetaDataIntegration_VendorData tests vendor-data endpoint
func TestMetaDataIntegration_VendorData(t *testing.T) {
	api, cleanup := setupIntegrationTestAPI(t)
	defer cleanup()

	r := chi.NewRouter()
	api.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/vendor-data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Vendor data is currently empty, so we just check the response
}

// TestMetaDataIntegration_NetworkConfig tests network-config endpoint
func TestMetaDataIntegration_NetworkConfig(t *testing.T) {
	api, cleanup := setupIntegrationTestAPI(t)
	defer cleanup()

	r := chi.NewRouter()
	api.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/network-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Network config is currently empty, so we just check the response
}
