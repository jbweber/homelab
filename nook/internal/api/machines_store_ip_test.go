package api

import (
	"context"
	"errors"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
)

type mockIPLeaseRepo struct {
	err error
}

func (m *mockIPLeaseRepo) Save(ctx context.Context, lease domain.IPAddressLease) (domain.IPAddressLease, error) {
	return domain.IPAddressLease{}, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) FindByID(ctx context.Context, id int64) (domain.IPAddressLease, error) {
	return domain.IPAddressLease{}, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) FindAll(ctx context.Context) ([]domain.IPAddressLease, error) {
	return []domain.IPAddressLease{}, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) DeleteByID(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockIPLeaseRepo) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) FindByMachineID(ctx context.Context, machineID int64) ([]domain.IPAddressLease, error) {
	return []domain.IPAddressLease{}, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) FindByNetworkID(ctx context.Context, networkID int64) ([]domain.IPAddressLease, error) {
	return []domain.IPAddressLease{}, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) FindByIPAddress(ctx context.Context, ipAddress string) (*domain.IPAddressLease, error) {
	return nil, errors.New("not implemented")
}

func (m *mockIPLeaseRepo) AllocateIPAddress(ctx context.Context, machineID, networkID int64) (*domain.IPAddressLease, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.IPAddressLease{
		ID:        1,
		MachineID: machineID,
		NetworkID: networkID,
		IPAddress: "192.168.1.100",
		LeaseTime: "24h",
	}, nil
}

func (m *mockIPLeaseRepo) DeallocateIPAddress(ctx context.Context, machineID, networkID int64) error {
	return m.err
}

func (m *mockIPLeaseRepo) IsIPAddressAvailable(ctx context.Context, networkID int64, ipAddress string) (bool, error) {
	return false, errors.New("not implemented")
}

func TestAPI_AllocateIPAddress_Success(t *testing.T) {
	mockRepo := &mockIPLeaseRepo{}
	api := &API{ipLeaseRepo: mockRepo}

	ip, err := api.AllocateIPAddress(1, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ip != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", ip)
	}
}

func TestAPI_AllocateIPAddress_Error(t *testing.T) {
	mockRepo := &mockIPLeaseRepo{err: errors.New("allocation error")}
	api := &API{ipLeaseRepo: mockRepo}

	ip, err := api.AllocateIPAddress(1, 1)
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	if ip != "" {
		t.Errorf("Expected empty IP on error, got '%s'", ip)
	}
}

func TestAPI_DeallocateIPAddress_Success(t *testing.T) {
	mockRepo := &mockIPLeaseRepo{}
	api := &API{ipLeaseRepo: mockRepo}

	err := api.DeallocateIPAddress(1, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAPI_DeallocateIPAddress_Error(t *testing.T) {
	mockRepo := &mockIPLeaseRepo{err: errors.New("deallocation error")}
	api := &API{ipLeaseRepo: mockRepo}

	err := api.DeallocateIPAddress(1, 1)
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}
