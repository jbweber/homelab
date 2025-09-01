package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockMetaDataStore struct {
	machine *Machine
	err     error
}

func (m *mockMetaDataStore) GetMachineByIPv4(ipv4 string) (*Machine, error) {
	return m.machine, m.err
}

func TestNoCloudMetaDataHandler_Success(t *testing.T) {
	store := &mockMetaDataStore{
		machine: &Machine{ID: 42, Name: "test", Hostname: "testhost", IPv4: "1.2.3.4"},
	}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	meta.NoCloudMetaDataHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/yaml; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/yaml; charset=utf-8', got '%s'", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	expectedContent := `instance-id: iid-00000042
hostname: testhost
local-hostname: testhost
local-ipv4: 1.2.3.4
public-hostname: testhost
security-groups: default
`
	if string(body) != expectedContent {
		t.Errorf("unexpected response body:\nexpected:\n%s\ngot:\n%s", expectedContent, string(body))
	}
}

func TestNoCloudMetaDataHandler_NotFound(t *testing.T) {
	store := &mockMetaDataStore{machine: nil, err: nil}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	meta.NoCloudMetaDataHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestNoCloudMetaDataHandler_Error(t *testing.T) {
	store := &mockMetaDataStore{machine: nil, err: errors.New("fail")}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	meta.NoCloudMetaDataHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_Success(t *testing.T) {
	store := &mockMetaDataStore{
		machine: &Machine{ID: 42, Name: "test", Hostname: "testhost", IPv4: "1.2.3.4"},
	}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "instance-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_UnknownKey(t *testing.T) {
	store := &mockMetaDataStore{
		machine: &Machine{ID: 42, Name: "test", Hostname: "testhost", IPv4: "1.2.3.4"},
	}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/unknown", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "unknown")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_Error(t *testing.T) {
	store := &mockMetaDataStore{machine: nil, err: errors.New("fail")}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "instance-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestNoCloudMetaDataHandler_BadRemoteAddr(t *testing.T) {
	store := &mockMetaDataStore{}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "malformed-addr"
	w := httptest.NewRecorder()
	meta.NoCloudMetaDataHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_BadRemoteAddr(t *testing.T) {
	store := &mockMetaDataStore{}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.RemoteAddr = "malformed-addr"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "instance-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_InvalidIP(t *testing.T) {
	store := &mockMetaDataStore{}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.RemoteAddr = "999.999.999.999:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "instance-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_EmptyKey(t *testing.T) {
	store := &mockMetaDataStore{}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestMetaDataKeyHandler_Success_ContentType(t *testing.T) {
	store := &mockMetaDataStore{
		machine: &Machine{ID: 42, Name: "test", Hostname: "testhost", IPv4: "1.2.3.4"},
	}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/instance-id", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "instance-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	w := httptest.NewRecorder()
	meta.MetaDataKeyHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/plain; charset=utf-8', got '%s'", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	expected := "iid-00000042\n"
	if string(body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}

func TestMetaDataDirectoryHandler_Success(t *testing.T) {
	store := &mockMetaDataStore{}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data/", nil)
	w := httptest.NewRecorder()
	meta.MetaDataDirectoryHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/plain; charset=utf-8', got '%s'", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	expected := `instance-id
hostname
local-hostname
local-ipv4
public-hostname
security-groups
`
	if string(body) != expected {
		t.Errorf("unexpected directory listing:\nexpected:\n%s\ngot:\n%s", expected, string(body))
	}
}
