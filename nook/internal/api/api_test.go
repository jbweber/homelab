package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
	"github.com/jbweber/homelab/nook/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSSHKeysStore implements SSHKeysStore for testing error cases
type mockSSHKeysStore struct {
	getMachineError  error
	listSSHKeysError error
}

func (m *mockSSHKeysStore) ListAllSSHKeys() ([]datastore.SSHKey, error) {
	return nil, nil
}

func (m *mockSSHKeysStore) GetMachineByIPv4(ip string) (*datastore.Machine, error) {
	if m.getMachineError != nil {
		return nil, m.getMachineError
	}
	return &datastore.Machine{ID: 1, Name: "test", Hostname: "test", IPv4: ip}, nil
}

func (m *mockSSHKeysStore) ListSSHKeys(machineID int64) ([]datastore.SSHKey, error) {
	if m.listSSHKeysError != nil {
		return nil, m.listSSHKeysError
	}
	return []datastore.SSHKey{}, nil
}

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

func TestInstanceIdentityDocumentHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/dynamic/instance-identity/document", nil)
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

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
	req := httptest.NewRequest("GET", "/user-data", nil)
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

func TestNoCloudNetworkConfigHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/network-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
	assert.Contains(t, w.Body.String(), "version: 2")
}

func TestNoCloudMetaDataHandler(t *testing.T) {
	r := setupTestAPI(t)
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
	assert.Contains(t, w.Body.String(), "machine not found for IP")
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
	// Simulate a lookup error by passing an invalid IP format
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Should be 404 due to not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestNetworksHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/networks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Optionally check body if expected
}

func TestSSHKeysHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Optionally check body if expected
}

func TestMetaDataDirectoryHandler(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/meta-data/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestMetaDataKeyHandler(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine for the test IP
	reqBody := CreateMachineRequest{
		Name:     "meta-key-machine",
		Hostname: "meta-key-host",
		IPv4:     "192.168.1.250",
	}
	body, _ := json.Marshal(reqBody)
	createReq := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.RemoteAddr = "192.168.1.250:12345"
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	// Now request the metadata key for the same IP
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.250")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// The output should be the instance-id value, e.g. "iid-00000001\n"
	assert.Regexp(t, `iid-\d{8}\n`, w.Body.String())
}

func TestInstanceIdentityDocumentHandler_MachineNotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/dynamic/instance-identity/document", nil)
	req.RemoteAddr = "203.0.113.99:12345" // IP not in test DB
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}
func TestNoCloudMetaDataHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unable to parse remote address")
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

func TestInstanceIdentityDocumentHandler_XForwardedFor(t *testing.T) {
	r := setupTestAPI(t)
	// Create a machine with a known IP
	reqBody := CreateMachineRequest{
		Name:     "xforwarded-machine",
		Hostname: "xforwarded-host",
		IPv4:     "192.168.1.251",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v0/machines", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.251:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now call instance identity with X-Forwarded-For
	req2 := httptest.NewRequest("GET", "/2021-01-03/dynamic/instance-identity/document", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.251")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var doc map[string]interface{}
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&doc))
	assert.Equal(t, "xforwarded-host", doc["hostname"])
	assert.Equal(t, "192.168.1.251", doc["privateIp"])
}

func TestInstanceIdentityDocumentHandler_IPNotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/dynamic/instance-identity/document", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99") // Not in test DB
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestInstanceIdentityDocumentHandler_LookupError(t *testing.T) {
	r := setupTestAPI(t)
	// Simulate a lookup error by passing an invalid IP format
	req := httptest.NewRequest("GET", "/2021-01-03/dynamic/instance-identity/document", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Should be 404 due to not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
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

func TestSSHKeysHandler_Placeholder(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/v0/ssh-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
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

func TestPublicKeysHandler_Success(t *testing.T) {
	r := setupTestAPI(t)
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

	// Insert SSH key for this machine
	ds, _ := datastore.New(testutil.NewTestDSN("TestAPI"))
	_, err := ds.CreateSSHKey(created.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey1")
	require.NoError(t, err)
	_, err = ds.CreateSSHKey(created.ID, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2")
	require.NoError(t, err)

	// Use X-Forwarded-For to trigger publicKeysHandler
	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.170")
	w2 := httptest.NewRecorder()
	api := NewAPI(ds)
	mux := chi.NewRouter()
	api.RegisterRoutes(mux)
	mux.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	bodyStr := w2.Body.String()
	assert.Contains(t, bodyStr, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey1")
	assert.Contains(t, bodyStr, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2")
}

func TestPublicKeysHandler_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99") // Not in test DB
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeysHandler_LookupError(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Should be 404 due to not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeysHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys", sshKeys.PublicKeysHandler)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys", sshKeys.PublicKeysHandler)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to lookup machine by IP")
}

func TestPublicKeyByIdxHandler_Success(t *testing.T) {
	ds, _ := datastore.New(testutil.NewTestDSN("TestAPI"))
	api := NewAPI(ds)
	mux := chi.NewRouter()
	api.RegisterRoutes(mux)

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
	mux.ServeHTTP(w, req)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	_, err := ds.CreateSSHKey(created.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCidxkey1")
	require.NoError(t, err)
	_, err = ds.CreateSSHKey(created.ID, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIidxkey2")
	require.NoError(t, err)

	// Get key at idx 1
	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/1", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.180")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIidxkey2")
}

func TestPublicKeyByIdxHandler_NotFound(t *testing.T) {
	ds, _ := datastore.New(testutil.NewTestDSN("TestAPI"))
	api := NewAPI(ds)
	mux := chi.NewRouter()
	api.RegisterRoutes(mux)

	// No machine for IP
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
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
	mux.ServeHTTP(w2, req2)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&created))

	// Out of range idx
	req3 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req3.Header.Set("X-Forwarded-For", "192.168.1.182")
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
	assert.Contains(t, w3.Body.String(), "key index out of range")
}

func TestPublicKeyByIdxHandler_InvalidIndex(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/invalid", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.170")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid key index")
}

func TestPublicKeyByIdxHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys/{idx}", sshKeys.PublicKeyByIdxHandler)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys/{idx}", sshKeys.PublicKeyByIdxHandler)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list SSH keys")
}

func TestPublicKeyOpenSSHHandler_Success(t *testing.T) {
	ds, _ := datastore.New(testutil.NewTestDSN("TestAPI"))
	api := NewAPI(ds)
	mux := chi.NewRouter()
	api.RegisterRoutes(mux)

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
	mux.ServeHTTP(w, req)
	var created MachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))

	_, err := ds.CreateSSHKey(created.ID, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCopensshkey1")
	require.NoError(t, err)

	req2 := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.190")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCopensshkey1")
}

func TestPublicKeyOpenSSHHandler_NotFound(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeyOpenSSHHandler_LookupError(t *testing.T) {
	r := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "machine not found for IP")
}

func TestPublicKeyOpenSSHHandler_MalformedRemoteAddr(t *testing.T) {
	r := setupTestAPI(t)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys/{idx}/openssh-key", sshKeys.PublicKeyOpenSSHHandler)
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
	sshKeys := NewSSHKeys(mockStore)
	r := chi.NewRouter()
	r.Get("/2021-01-03/meta-data/public-keys/{idx}/openssh-key", sshKeys.PublicKeyOpenSSHHandler)
	req := httptest.NewRequest("GET", "/2021-01-03/meta-data/public-keys/0/openssh-key", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list SSH keys")
}
