package api

import (
	"context"
	"errors"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
)

type mockNetworkRepo struct {
	networks []domain.Network
	err      error
}

func (m *mockNetworkRepo) Save(ctx context.Context, network domain.Network) (domain.Network, error) {
	return network, nil
}

func (m *mockNetworkRepo) FindByID(ctx context.Context, id int64) (domain.Network, error) {
	return domain.Network{}, nil
}

func (m *mockNetworkRepo) FindByName(ctx context.Context, name string) (domain.Network, error) {
	if m.err != nil {
		return domain.Network{}, m.err
	}
	for _, network := range m.networks {
		if network.Name == name {
			return network, nil
		}
	}
	return domain.Network{}, errors.New("network not found")
}

func (m *mockNetworkRepo) FindByBridge(ctx context.Context, bridge string) (domain.Network, error) {
	return domain.Network{}, errors.New("not implemented")
}

func (m *mockNetworkRepo) FindAll(ctx context.Context) ([]domain.Network, error) {
	return m.networks, nil
}

func (m *mockNetworkRepo) DeleteByID(ctx context.Context, id int64) error {
	return nil
}

func (m *mockNetworkRepo) GetDHCPRanges(ctx context.Context, networkID int64) ([]domain.DHCPRange, error) {
	return []domain.DHCPRange{}, nil
}

func (m *mockNetworkRepo) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return false, errors.New("not implemented")
}

type mockDHCPRangeRepo struct {
	err error
}

func (m *mockDHCPRangeRepo) Save(ctx context.Context, dhcpRange domain.DHCPRange) (domain.DHCPRange, error) {
	return dhcpRange, nil
}

func (m *mockDHCPRangeRepo) FindByID(ctx context.Context, id int64) (domain.DHCPRange, error) {
	return domain.DHCPRange{}, errors.New("not implemented")
}

func (m *mockDHCPRangeRepo) FindAll(ctx context.Context) ([]domain.DHCPRange, error) {
	return []domain.DHCPRange{}, errors.New("not implemented")
}

func (m *mockDHCPRangeRepo) DeleteByID(ctx context.Context, id int64) error {
	return m.err
}

func (m *mockDHCPRangeRepo) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockDHCPRangeRepo) FindByNetworkID(ctx context.Context, networkID int64) ([]domain.DHCPRange, error) {
	return []domain.DHCPRange{}, errors.New("not implemented")
}

func TestAPI_GetNetworkByName_Success(t *testing.T) {
	mockRepo := &mockNetworkRepo{
		networks: []domain.Network{
			{ID: 1, Name: "test-network", Bridge: "br0", Subnet: "192.168.1.0/24"},
		},
	}
	api := &API{networkRepo: mockRepo}

	network, err := api.GetNetworkByName("test-network")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if network.Name != "test-network" {
		t.Errorf("Expected network name 'test-network', got '%s'", network.Name)
	}
}

func TestAPI_GetNetworkByName_NotFound(t *testing.T) {
	mockRepo := &mockNetworkRepo{networks: []domain.Network{}}
	api := &API{networkRepo: mockRepo}

	_, err := api.GetNetworkByName("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent network")
	}
}

func TestAPI_GetNetworkByName_Error(t *testing.T) {
	mockRepo := &mockNetworkRepo{err: errors.New("repository error")}
	api := &API{networkRepo: mockRepo}

	_, err := api.GetNetworkByName("test-network")
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}

func TestAPI_DeleteDHCPRange_Success(t *testing.T) {
	mockRepo := &mockDHCPRangeRepo{}
	api := &API{dhcpRangeRepo: mockRepo}

	err := api.DeleteDHCPRange(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAPI_DeleteDHCPRange_Error(t *testing.T) {
	mockRepo := &mockDHCPRangeRepo{err: errors.New("deletion error")}
	api := &API{dhcpRangeRepo: mockRepo}

	err := api.DeleteDHCPRange(1)
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}
