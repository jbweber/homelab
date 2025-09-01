package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/migrations"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetMachineByName_MissingName(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/machines/name/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMachineHandler_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/machines/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMachineHandler_InvalidID(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/machines/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMachine_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("DELETE", "/api/v0/machines/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetMachineByIPv4_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/machines/ipv4/203.0.113.99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMachineByName_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/machines/name/no-such-machine", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func setupTestAPI(t *testing.T) *chi.Mux {
	// Create test database with migrations
	db, cleanup := testutil.SetupTestDB(t, "TestAPI")
	t.Cleanup(cleanup)

	// Run migrations
	migrator := migrations.NewMigrator(db)
	for _, migration := range migrations.GetInitialMigrations() {
		migrator.AddMigration(migration)
	}

	if err := migrator.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	api := NewAPI(db)
	api.RegisterRoutes(r)

	return r
}

func TestListMachines_Empty(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v0/machines", nil)
	req.RemoteAddr = "192.168.1.100:12345"
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
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))

	assert.Equal(t, reqBody.Name, response.Name)
	assert.Equal(t, reqBody.Hostname, response.Hostname)
	assert.Equal(t, reqBody.IPv4, response.IPv4)
	assert.NotZero(t, response.ID)
}

func TestCreateMachine_InvalidJSON(t *testing.T) {
	r := setupTestAPI(t)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMachine_MissingFields(t *testing.T) {
	r := setupTestAPI(t)

	reqBody := CreateMachineRequest{Name: "test"} // Missing Hostname and IPv4
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMachine_InvalidIPv4(t *testing.T) {
	r := setupTestAPI(t)

	reqBody := CreateMachineRequest{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "invalid-ip",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid IPv4 address format")
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

func TestNoCloudUserDataHandler(t *testing.T) {
	r := setupTestAPI(t)

	// Create a machine with the IP that will be making the request
	reqBody := CreateMachineRequest{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.0.2.1",
	}
	body, _ := json.Marshal(reqBody)

	createReq := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.RemoteAddr = "192.0.2.1:12345"
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	req := httptest.NewRequest("GET", "/user-data", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Optionally check body if expected
}

func TestNoCloudVendorDataHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/vendor-data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestNoCloudMetaDataHandler(t *testing.T) {
	r := setupTestAPI(t)

	// Create a machine with the IP that will be making the request
	reqBody := CreateMachineRequest{
		Name:     "test-machine",
		Hostname: "test-host",
		IPv4:     "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	createReq := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.RemoteAddr = "192.168.1.100:12345"
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	// Now test the metadata handler
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestNoCloudMetaDataHandler_IPNotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "203.0.113.99:12345" // IP not in test DB
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found")
}

func TestNoCloudMetaDataHandler_XForwardedFor(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine with a known IP
	reqBody := CreateMachineRequest{
		Name:     "meta-xforwarded",
		Hostname: "meta-xhost",
		IPv4:     "192.168.1.222",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.222:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now call meta-data with X-Forwarded-For
	req2 := httptest.NewRequest("GET", "/meta-data", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.222")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEmpty(t, w2.Body.String())
	assert.Contains(t, w2.Body.String(), "meta-xhost")
}

func TestNoCloudMetaDataHandler_LookupError(t *testing.T) {
	r := setupTestAPI(t)
	// Simulate an invalid IP format - should return 400 Bad Request due to IP validation
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid IP address format")
}

func TestNetworksHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/networks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Optionally check body if expected
}



func TestNoCloudMetaDataHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unable to determine client IP address")
}

func TestCreateMachine_DuplicateName(t *testing.T) {
	r := setupTestAPI(t)
	// First create a machine
	reqBody := CreateMachineRequest{
		Name:     "dup-machine",
		Hostname: "dup-host",
		IPv4:     "192.168.1.101",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.101:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to create another machine with the same name
	req2 := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.RemoteAddr = "192.168.1.102:12345"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
	assert.Contains(t, w2.Body.String(), "IPv4 address already exists")
}

func TestDeleteMachine_InvalidID(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("DELETE", "/api/v0/machines/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMachine_Success(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "delete-machine",
		Hostname: "delete-host",
		IPv4:     "192.168.1.180",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.180:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var created MachineResponse
	err := json.NewDecoder(w.Body).Decode(&created)
	require.NoError(t, err)

	// Delete the machine
	deleteReq := httptest.NewRequest("DELETE", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), nil)
	deleteW := httptest.NewRecorder()
	r.ServeHTTP(deleteW, deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	// Verify it's gone
	getReq := httptest.NewRequest("GET", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	assert.Equal(t, http.StatusNotFound, getW.Code)
}

func TestGetMachineByName_Valid(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "find-by-name",
		Hostname: "find-host",
		IPv4:     "192.168.1.150",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.150:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now lookup by name
	req2 := httptest.NewRequest("GET", "/api/v0/machines/name/find-by-name", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var resp MachineResponse
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&resp))
	assert.Equal(t, reqBody.Name, resp.Name)
	assert.Equal(t, reqBody.Hostname, resp.Hostname)
	assert.Equal(t, reqBody.IPv4, resp.IPv4)
}

func TestGetMachineByIPv4_Valid(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "find-by-ipv4",
		Hostname: "find-host",
		IPv4:     "192.168.1.151",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.151:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now lookup by IPv4
	req2 := httptest.NewRequest("GET", "/api/v0/machines/ipv4/192.168.1.151", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var resp MachineResponse
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&resp))
	assert.Equal(t, reqBody.Name, resp.Name)
	assert.Equal(t, reqBody.Hostname, resp.Hostname)
	assert.Equal(t, reqBody.IPv4, resp.IPv4)
}

func TestListMachines_NonEmpty(t *testing.T) {
	r := setupTestAPI(t)
	// Create two machines
	reqBody1 := CreateMachineRequest{
		Name:     "machine1",
		Hostname: "host1",
		IPv4:     "192.168.1.201",
	}
	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.RemoteAddr = "192.168.1.201:12345"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	reqBody2 := CreateMachineRequest{
		Name:     "machine2",
		Hostname: "host2",
		IPv4:     "192.168.1.202",
	}
	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.RemoteAddr = "192.168.1.202:12345"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// List machines
	reqList := httptest.NewRequest("GET", "/api/v0/machines", nil)
	reqList.RemoteAddr = "192.168.1.201:12345"
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)
	var response []MachineResponse
	require.NoError(t, json.NewDecoder(wList.Body).Decode(&response))
	// Check that both expected machines are present
	found1, found2 := false, false
	for _, m := range response {
		if m.Name == "machine1" && m.Hostname == "host1" && m.IPv4 == "192.168.1.201" {
			found1 = true
		}
		if m.Name == "machine2" && m.Hostname == "host2" && m.IPv4 == "192.168.1.202" {
			found2 = true
		}
	}
	assert.True(t, found1, "machine1 not found in response")
	assert.True(t, found2, "machine2 not found in response")
}

func TestGetMachineHandler_Valid(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine
	reqBody := CreateMachineRequest{
		Name:     "get-machine",
		Hostname: "get-host",
		IPv4:     "192.168.1.210",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.210:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	// Get by ID
	reqGet := httptest.NewRequest("GET", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	var resp MachineResponse
	require.NoError(t, json.NewDecoder(wGet.Body).Decode(&resp))
	assert.Equal(t, created.ID, resp.ID)
	assert.Equal(t, reqBody.Name, resp.Name)
	assert.Equal(t, reqBody.Hostname, resp.Hostname)
	assert.Equal(t, reqBody.IPv4, resp.IPv4)
}







// End of Tests
func TestNetworksHandler_Placeholder(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/networks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[networks endpoint placeholder]", w.Body.String())
}

func TestUpdateMachineHandler_Success(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "update-machine",
		Hostname: "update-host",
		IPv4:     "192.168.1.160",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var created MachineResponse
	err := json.NewDecoder(w.Body).Decode(&created)
	require.NoError(t, err)

	// Update the machine
	updateBody := CreateMachineRequest{
		Name:     "updated-machine",
		Hostname: "updated-host",
		IPv4:     "192.168.1.161",
	}
	updateJSON, _ := json.Marshal(updateBody)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), bytes.NewReader(updateJSON))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusOK, patchW.Code)
	var updated MachineResponse
	err = json.NewDecoder(patchW.Body).Decode(&updated)
	require.NoError(t, err)
	assert.Equal(t, updateBody.Name, updated.Name)
	assert.Equal(t, updateBody.Hostname, updated.Hostname)
	assert.Equal(t, updateBody.IPv4, updated.IPv4)
}

func TestUpdateMachineHandler_InvalidID(t *testing.T) {
	r := setupTestAPI(t)
	updateBody := CreateMachineRequest{
		Name:     "updated-machine",
		Hostname: "updated-host",
		IPv4:     "192.168.1.161",
	}
	updateJSON, _ := json.Marshal(updateBody)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/invalid", bytes.NewReader(updateJSON))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusBadRequest, patchW.Code)
	assert.Contains(t, patchW.Body.String(), "Invalid machine ID")
}

func TestUpdateMachineHandler_InvalidJSON(t *testing.T) {
	r := setupTestAPI(t)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/1", bytes.NewReader([]byte("invalid json")))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusBadRequest, patchW.Code)
	assert.Contains(t, patchW.Body.String(), "Invalid JSON")
}

func TestUpdateMachineHandler_InvalidIPv4(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "update-machine",
		Hostname: "update-host",
		IPv4:     "192.168.1.160",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.160:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var created MachineResponse
	err := json.NewDecoder(w.Body).Decode(&created)
	require.NoError(t, err)

	// Update with invalid IPv4
	updateBody := CreateMachineRequest{
		Name:     "updated-machine",
		Hostname: "updated-host",
		IPv4:     "invalid-ip",
	}
	updateJSON, _ := json.Marshal(updateBody)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), bytes.NewReader(updateJSON))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusBadRequest, patchW.Code)
	assert.Contains(t, patchW.Body.String(), "Invalid IPv4 address format")
}

func TestUpdateMachineHandler_MissingFields(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine first
	reqBody := CreateMachineRequest{
		Name:     "update-machine",
		Hostname: "update-host",
		IPv4:     "192.168.1.160",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var created MachineResponse
	err := json.NewDecoder(w.Body).Decode(&created)
	require.NoError(t, err)

	// Missing fields
	updateBody := CreateMachineRequest{Name: ""}
	updateJSON, _ := json.Marshal(updateBody)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/"+strconv.Itoa(int(created.ID)), bytes.NewReader(updateJSON))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusBadRequest, patchW.Code)
}

func TestUpdateMachineHandler_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	updateBody := CreateMachineRequest{
		Name:     "updated-machine",
		Hostname: "updated-host",
		IPv4:     "192.168.1.161",
	}
	updateJSON, _ := json.Marshal(updateBody)
	patchReq := httptest.NewRequest("PATCH", "/api/v0/machines/99999", bytes.NewReader(updateJSON))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	assert.Equal(t, http.StatusNotFound, patchW.Code)
}
