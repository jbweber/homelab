package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// NetworkRepository defines domain-specific operations for networks
type NetworkRepository interface {
	Repository[domain.Network, int64]
	FindByName(ctx context.Context, name string) (domain.Network, error)
	FindByBridge(ctx context.Context, bridge string) (domain.Network, error)
	GetDHCPRanges(ctx context.Context, networkID int64) ([]domain.DHCPRange, error)
}

// networkRepositoryImpl implements NetworkRepository
type networkRepositoryImpl struct {
	db *sql.DB
}

// NewNetworkRepository creates a new network repository
func NewNetworkRepository(db *sql.DB) NetworkRepository {
	return &networkRepositoryImpl{
		db: db,
	}
}

// Save creates or updates a network
func (r *networkRepositoryImpl) Save(ctx context.Context, network domain.Network) (domain.Network, error) {
	if network.ID == 0 {
		// Create new network
		return r.createNetwork(network)
	} else {
		// Update existing network
		return r.updateNetwork(network)
	}
}

// createNetwork inserts a new network into the database
func (r *networkRepositoryImpl) createNetwork(n domain.Network) (domain.Network, error) {
	if n.Name == "" {
		return domain.Network{}, fmt.Errorf("network name is required")
	}
	if n.Bridge == "" {
		return domain.Network{}, fmt.Errorf("network bridge is required")
	}
	if n.Subnet == "" {
		return domain.Network{}, fmt.Errorf("network subnet is required")
	}

	// Check for duplicate name
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM networks WHERE name = ?", n.Name).Scan(&count)
	if err != nil {
		return domain.Network{}, fmt.Errorf("failed to check for duplicate network name: %w", err)
	}
	if count > 0 {
		return domain.Network{}, fmt.Errorf("network with name '%s' already exists", n.Name)
	}

	result, err := r.db.Exec(`
		INSERT INTO networks (name, bridge, subnet, gateway, dns_servers, description)
		VALUES (?, ?, ?, ?, ?, ?)`,
		n.Name, n.Bridge, n.Subnet, n.Gateway, n.DNSServers, n.Description)
	if err != nil {
		return domain.Network{}, fmt.Errorf("failed to create network: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Network{}, fmt.Errorf("failed to get network ID: %w", err)
	}

	n.ID = id
	return n, nil
}

// updateNetwork updates an existing network in the database
func (r *networkRepositoryImpl) updateNetwork(n domain.Network) (domain.Network, error) {
	if n.Name == "" {
		return domain.Network{}, fmt.Errorf("network name is required")
	}
	if n.Bridge == "" {
		return domain.Network{}, fmt.Errorf("network bridge is required")
	}
	if n.Subnet == "" {
		return domain.Network{}, fmt.Errorf("network subnet is required")
	}

	// Check for duplicate name (excluding current network)
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM networks WHERE name = ? AND id != ?", n.Name, n.ID).Scan(&count)
	if err != nil {
		return domain.Network{}, fmt.Errorf("failed to check for duplicate network name: %w", err)
	}
	if count > 0 {
		return domain.Network{}, fmt.Errorf("network with name '%s' already exists", n.Name)
	}

	_, err = r.db.Exec(`
		UPDATE networks
		SET name = ?, bridge = ?, subnet = ?, gateway = ?, dns_servers = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		n.Name, n.Bridge, n.Subnet, n.Gateway, n.DNSServers, n.Description, n.ID)
	if err != nil {
		return domain.Network{}, fmt.Errorf("failed to update network: %w", err)
	}

	return n, nil
}

// FindByID finds a network by ID
func (r *networkRepositoryImpl) FindByID(ctx context.Context, id int64) (domain.Network, error) {
	var network domain.Network
	err := r.db.QueryRow(`
		SELECT id, name, bridge, subnet, gateway, dns_servers, description
		FROM networks WHERE id = ?`, id).Scan(
		&network.ID, &network.Name, &network.Bridge, &network.Subnet,
		&network.Gateway, &network.DNSServers, &network.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Network{}, fmt.Errorf("network with ID %d not found", id)
		}
		return domain.Network{}, fmt.Errorf("failed to find network: %w", err)
	}
	return network, nil
}

// FindByName finds a network by name
func (r *networkRepositoryImpl) FindByName(ctx context.Context, name string) (domain.Network, error) {
	var network domain.Network
	err := r.db.QueryRow(`
		SELECT id, name, bridge, subnet, gateway, dns_servers, description
		FROM networks WHERE name = ?`, name).Scan(
		&network.ID, &network.Name, &network.Bridge, &network.Subnet,
		&network.Gateway, &network.DNSServers, &network.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Network{}, fmt.Errorf("network with name '%s' not found", name)
		}
		return domain.Network{}, fmt.Errorf("failed to find network: %w", err)
	}
	return network, nil
}

// FindByBridge finds a network by bridge interface
func (r *networkRepositoryImpl) FindByBridge(ctx context.Context, bridge string) (domain.Network, error) {
	var network domain.Network
	err := r.db.QueryRow(`
		SELECT id, name, bridge, subnet, gateway, dns_servers, description
		FROM networks WHERE bridge = ?`, bridge).Scan(
		&network.ID, &network.Name, &network.Bridge, &network.Subnet,
		&network.Gateway, &network.DNSServers, &network.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Network{}, fmt.Errorf("network with bridge '%s' not found", bridge)
		}
		return domain.Network{}, fmt.Errorf("failed to find network: %w", err)
	}
	return network, nil
}

// FindAll finds all networks
func (r *networkRepositoryImpl) FindAll(ctx context.Context) ([]domain.Network, error) {
	rows, err := r.db.Query(`
		SELECT id, name, bridge, subnet, gateway, dns_servers, description
		FROM networks ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to find networks: %w", err)
	}
	defer rows.Close()

	var networks []domain.Network
	for rows.Next() {
		var network domain.Network
		err := rows.Scan(
			&network.ID, &network.Name, &network.Bridge, &network.Subnet,
			&network.Gateway, &network.DNSServers, &network.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan network: %w", err)
		}
		networks = append(networks, network)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating networks: %w", err)
	}

	return networks, nil
}

// DeleteByID deletes a network by ID
func (r *networkRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.Exec("DELETE FROM networks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("network with ID %d not found", id)
	}

	return nil
}

// GetDHCPRanges gets all DHCP ranges for a network
func (r *networkRepositoryImpl) GetDHCPRanges(ctx context.Context, networkID int64) ([]domain.DHCPRange, error) {
	rows, err := r.db.Query(`
		SELECT id, network_id, start_ip, end_ip, lease_time
		FROM dhcp_ranges WHERE network_id = ? ORDER BY start_ip`, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCP ranges: %w", err)
	}
	defer rows.Close()

	var ranges []domain.DHCPRange
	for rows.Next() {
		var dhcpRange domain.DHCPRange
		err := rows.Scan(
			&dhcpRange.ID, &dhcpRange.NetworkID, &dhcpRange.StartIP,
			&dhcpRange.EndIP, &dhcpRange.LeaseTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DHCP range: %w", err)
		}
		ranges = append(ranges, dhcpRange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating DHCP ranges: %w", err)
	}

	return ranges, nil
}

// ExistsByID checks if a network exists by ID
func (r *networkRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM networks WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check network existence: %w", err)
	}
	return count > 0, nil
}
