package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jbweber/homelab/nook/internal/datastore"
)

type mockMetaDataStore struct {
	machine *datastore.Machine
	err     error
}

func (m *mockMetaDataStore) GetMachineByIPv4(ipv4 string) (*datastore.Machine, error) {
	return m.machine, m.err
}

func TestNoCloudMetaDataHandler_Success(t *testing.T) {
	store := &mockMetaDataStore{
		machine: &datastore.Machine{ID: 42, Hostname: "testhost", IPv4: "1.2.3.4"},
	}
	meta := NewMetaData(store)
	req := httptest.NewRequest("GET", "/meta-data", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	meta.NoCloudMetaDataHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
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
		machine: &datastore.Machine{ID: 42, Hostname: "testhost", IPv4: "1.2.3.4"},
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
		machine: &datastore.Machine{ID: 42, Hostname: "testhost", IPv4: "1.2.3.4"},
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
