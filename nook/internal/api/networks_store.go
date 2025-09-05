package api

import (
	"context"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// CreateNetwork implements NetworksStore interface
func (a *API) CreateNetwork(network domain.Network) (domain.Network, error) {
	return a.networkRepo.Save(context.Background(), network)
}

// GetNetwork implements NetworksStore interface
func (a *API) GetNetwork(id int64) (domain.Network, error) {
	return a.networkRepo.FindByID(context.Background(), id)
}

// GetNetworkByName implements NetworksStore interface
func (a *API) GetNetworkByName(name string) (domain.Network, error) {
	return a.networkRepo.FindByName(context.Background(), name)
}

// ListNetworks implements NetworksStore interface
func (a *API) ListNetworks() ([]domain.Network, error) {
	return a.networkRepo.FindAll(context.Background())
}

// UpdateNetwork implements NetworksStore interface
func (a *API) UpdateNetwork(network domain.Network) (domain.Network, error) {
	return a.networkRepo.Save(context.Background(), network)
}

// DeleteNetwork implements NetworksStore interface
func (a *API) DeleteNetwork(id int64) error {
	return a.networkRepo.DeleteByID(context.Background(), id)
}

// GetDHCPRanges implements NetworksStore interface
func (a *API) GetDHCPRanges(networkID int64) ([]domain.DHCPRange, error) {
	return a.networkRepo.GetDHCPRanges(context.Background(), networkID)
}

// CreateDHCPRange implements NetworksStore interface
func (a *API) CreateDHCPRange(dhcpRange domain.DHCPRange) (domain.DHCPRange, error) {
	return a.dhcpRangeRepo.Save(context.Background(), dhcpRange)
}

// DeleteDHCPRange implements NetworksStore interface
func (a *API) DeleteDHCPRange(id int64) error {
	return a.dhcpRangeRepo.DeleteByID(context.Background(), id)
}
