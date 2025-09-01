package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// IPLeaseRepository defines domain-specific operations for IP address leases
type IPLeaseRepository interface {
	Repository[domain.IPAddressLease, int64]
	FindByMachineID(ctx context.Context, machineID int64) ([]domain.IPAddressLease, error)
	FindByNetworkID(ctx context.Context, networkID int64) ([]domain.IPAddressLease, error)
	FindByIPAddress(ctx context.Context, ipAddress string) (*domain.IPAddressLease, error)
	AllocateIPAddress(ctx context.Context, machineID, networkID int64) (*domain.IPAddressLease, error)
	DeallocateIPAddress(ctx context.Context, machineID, networkID int64) error
	IsIPAddressAvailable(ctx context.Context, networkID int64, ipAddress string) (bool, error)
	ExistsByID(ctx context.Context, id int64) (bool, error)
}

// ipLeaseRepositoryImpl implements IPLeaseRepository
type ipLeaseRepositoryImpl struct {
	db *sql.DB
}

// NewIPLeaseRepository creates a new IP lease repository
func NewIPLeaseRepository(db *sql.DB) IPLeaseRepository {
	return &ipLeaseRepositoryImpl{
		db: db,
	}
}

// Save creates or updates an IP address lease
func (r *ipLeaseRepositoryImpl) Save(ctx context.Context, lease domain.IPAddressLease) (domain.IPAddressLease, error) {
	if lease.ID == 0 {
		return r.createLease(lease)
	} else {
		return r.updateLease(lease)
	}
}

// createLease inserts a new IP address lease into the database
func (r *ipLeaseRepositoryImpl) createLease(lease domain.IPAddressLease) (domain.IPAddressLease, error) {
	if lease.MachineID == 0 {
		return domain.IPAddressLease{}, fmt.Errorf("machine ID is required")
	}
	if lease.NetworkID == 0 {
		return domain.IPAddressLease{}, fmt.Errorf("network ID is required")
	}
	if lease.IPAddress == "" {
		return domain.IPAddressLease{}, fmt.Errorf("IP address is required")
	}

	// Validate IP address format
	if net.ParseIP(lease.IPAddress) == nil {
		return domain.IPAddressLease{}, fmt.Errorf("invalid IP address format: %s", lease.IPAddress)
	}

	// Check if IP is already leased
	available, err := r.IsIPAddressAvailable(context.Background(), lease.NetworkID, lease.IPAddress)
	if err != nil {
		return domain.IPAddressLease{}, fmt.Errorf("failed to check IP availability: %w", err)
	}
	if !available {
		return domain.IPAddressLease{}, fmt.Errorf("IP address %s is already leased", lease.IPAddress)
	}

	query := `
		INSERT INTO ip_address_leases (machine_id, network_id, ip_address, lease_time)
		VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(context.Background(), query,
		lease.MachineID, lease.NetworkID, lease.IPAddress, lease.LeaseTime)
	if err != nil {
		return domain.IPAddressLease{}, fmt.Errorf("failed to create IP lease: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.IPAddressLease{}, fmt.Errorf("failed to get lease ID: %w", err)
	}

	lease.ID = id
	return lease, nil
}

// updateLease updates an existing IP address lease in the database
func (r *ipLeaseRepositoryImpl) updateLease(lease domain.IPAddressLease) (domain.IPAddressLease, error) {
	if lease.ID == 0 {
		return domain.IPAddressLease{}, fmt.Errorf("lease ID is required for update")
	}

	query := `
		UPDATE ip_address_leases
		SET machine_id = ?, network_id = ?, ip_address = ?, lease_time = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := r.db.ExecContext(context.Background(), query,
		lease.MachineID, lease.NetworkID, lease.IPAddress, lease.LeaseTime, lease.ID)
	if err != nil {
		return domain.IPAddressLease{}, fmt.Errorf("failed to update IP lease: %w", err)
	}

	return lease, nil
}

// FindByID finds an IP address lease by ID
func (r *ipLeaseRepositoryImpl) FindByID(ctx context.Context, id int64) (domain.IPAddressLease, error) {
	query := `
		SELECT id, machine_id, network_id, ip_address, lease_time, created_at, updated_at
		FROM ip_address_leases
		WHERE id = ?`

	var lease domain.IPAddressLease
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lease.ID, &lease.MachineID, &lease.NetworkID, &lease.IPAddress,
		&lease.LeaseTime, &lease.CreatedAt, &lease.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.IPAddressLease{}, ErrNotFound
		}
		return domain.IPAddressLease{}, fmt.Errorf("failed to find IP lease: %w", err)
	}

	return lease, nil
}

// FindAll finds all IP address leases
func (r *ipLeaseRepositoryImpl) FindAll(ctx context.Context) ([]domain.IPAddressLease, error) {
	query := `
		SELECT id, machine_id, network_id, ip_address, lease_time, created_at, updated_at
		FROM ip_address_leases
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find IP leases: %w", err)
	}
	defer rows.Close()

	var leases []domain.IPAddressLease
	for rows.Next() {
		var lease domain.IPAddressLease
		err := rows.Scan(
			&lease.ID, &lease.MachineID, &lease.NetworkID, &lease.IPAddress,
			&lease.LeaseTime, &lease.CreatedAt, &lease.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP lease: %w", err)
		}
		leases = append(leases, lease)
	}

	return leases, nil
}

// DeleteByID deletes an IP address lease by ID
func (r *ipLeaseRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM ip_address_leases WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete IP lease: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// ExistsByID checks if an IP lease exists by ID
func (r *ipLeaseRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	query := `SELECT COUNT(*) FROM ip_address_leases WHERE id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check IP lease existence: %w", err)
	}

	return count > 0, nil
}

// FindByMachineID finds all IP leases for a specific machine
func (r *ipLeaseRepositoryImpl) FindByMachineID(ctx context.Context, machineID int64) ([]domain.IPAddressLease, error) {
	query := `
		SELECT id, machine_id, network_id, ip_address, lease_time, created_at, updated_at
		FROM ip_address_leases
		WHERE machine_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find IP leases for machine: %w", err)
	}
	defer rows.Close()

	var leases []domain.IPAddressLease
	for rows.Next() {
		var lease domain.IPAddressLease
		err := rows.Scan(
			&lease.ID, &lease.MachineID, &lease.NetworkID, &lease.IPAddress,
			&lease.LeaseTime, &lease.CreatedAt, &lease.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP lease: %w", err)
		}
		leases = append(leases, lease)
	}

	return leases, nil
}

// FindByNetworkID finds all IP leases for a specific network
func (r *ipLeaseRepositoryImpl) FindByNetworkID(ctx context.Context, networkID int64) ([]domain.IPAddressLease, error) {
	query := `
		SELECT id, machine_id, network_id, ip_address, lease_time, created_at, updated_at
		FROM ip_address_leases
		WHERE network_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to find IP leases for network: %w", err)
	}
	defer rows.Close()

	var leases []domain.IPAddressLease
	for rows.Next() {
		var lease domain.IPAddressLease
		err := rows.Scan(
			&lease.ID, &lease.MachineID, &lease.NetworkID, &lease.IPAddress,
			&lease.LeaseTime, &lease.CreatedAt, &lease.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP lease: %w", err)
		}
		leases = append(leases, lease)
	}

	return leases, nil
}

// FindByIPAddress finds an IP lease by IP address
func (r *ipLeaseRepositoryImpl) FindByIPAddress(ctx context.Context, ipAddress string) (*domain.IPAddressLease, error) {
	query := `
		SELECT id, machine_id, network_id, ip_address, lease_time, created_at, updated_at
		FROM ip_address_leases
		WHERE ip_address = ?`

	var lease domain.IPAddressLease
	err := r.db.QueryRowContext(ctx, query, ipAddress).Scan(
		&lease.ID, &lease.MachineID, &lease.NetworkID, &lease.IPAddress,
		&lease.LeaseTime, &lease.CreatedAt, &lease.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find IP lease by address: %w", err)
	}

	return &lease, nil
}

// AllocateIPAddress finds and allocates an available IP address for the given machine and network
func (r *ipLeaseRepositoryImpl) AllocateIPAddress(ctx context.Context, machineID, networkID int64) (*domain.IPAddressLease, error) {
	if machineID == 0 {
		return nil, fmt.Errorf("machine ID is required")
	}
	// Get all DHCP ranges for this network
	dhcpRanges, err := r.getDHCPRangesForNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCP ranges: %w", err)
	}

	if len(dhcpRanges) == 0 {
		return nil, fmt.Errorf("no DHCP ranges configured for network %d", networkID)
	}

	// Try to find an available IP in each range
	for _, dhcpRange := range dhcpRanges {
		ip, err := r.findAvailableIPInRange(ctx, networkID, dhcpRange.StartIP, dhcpRange.EndIP)
		if err != nil {
			continue // Try next range
		}
		if ip != "" {
			// Found available IP, create lease
			lease := domain.IPAddressLease{
				MachineID: machineID,
				NetworkID: networkID,
				IPAddress: ip,
				LeaseTime: dhcpRange.LeaseTime,
			}
			createdLease, err := r.createLease(lease)
			if err != nil {
				return nil, err
			}
			return &createdLease, nil
		}
	}

	return nil, fmt.Errorf("no available IP addresses in network %d", networkID)
}

// DeallocateIPAddress removes the IP lease for a machine on a specific network
func (r *ipLeaseRepositoryImpl) DeallocateIPAddress(ctx context.Context, machineID, networkID int64) error {
	query := `DELETE FROM ip_address_leases WHERE machine_id = ? AND network_id = ?`

	result, err := r.db.ExecContext(ctx, query, machineID, networkID)
	if err != nil {
		return fmt.Errorf("failed to deallocate IP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// IsIPAddressAvailable checks if an IP address is available for leasing
func (r *ipLeaseRepositoryImpl) IsIPAddressAvailable(ctx context.Context, networkID int64, ipAddress string) (bool, error) {
	// Check if IP is already leased
	leaseQuery := `
		SELECT COUNT(*) FROM ip_address_leases
		WHERE network_id = ? AND ip_address = ?`

	var leaseCount int
	err := r.db.QueryRowContext(ctx, leaseQuery, networkID, ipAddress).Scan(&leaseCount)
	if err != nil {
		return false, fmt.Errorf("failed to check IP lease availability: %w", err)
	}

	// Check if IP is already assigned to a machine
	machineQuery := `
		SELECT COUNT(*) FROM machines
		WHERE ipv4 = ?`

	var machineCount int
	err = r.db.QueryRowContext(ctx, machineQuery, ipAddress).Scan(&machineCount)
	if err != nil {
		return false, fmt.Errorf("failed to check machine IP availability: %w", err)
	}

	return leaseCount == 0 && machineCount == 0, nil
}

// Helper methods

func (r *ipLeaseRepositoryImpl) getDHCPRangesForNetwork(ctx context.Context, networkID int64) ([]domain.DHCPRange, error) {
	query := `
		SELECT id, network_id, start_ip, end_ip, lease_time
		FROM dhcp_ranges
		WHERE network_id = ?
		ORDER BY start_ip`

	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCP ranges: %w", err)
	}
	defer rows.Close()

	var ranges []domain.DHCPRange
	for rows.Next() {
		var dhcpRange domain.DHCPRange
		err := rows.Scan(&dhcpRange.ID, &dhcpRange.NetworkID, &dhcpRange.StartIP, &dhcpRange.EndIP, &dhcpRange.LeaseTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DHCP range: %w", err)
		}
		ranges = append(ranges, dhcpRange)
	}

	return ranges, nil
}

func (r *ipLeaseRepositoryImpl) findAvailableIPInRange(ctx context.Context, networkID int64, startIP, endIP string) (string, error) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)
	if start == nil || end == nil {
		return "", fmt.Errorf("invalid IP range: %s - %s", startIP, endIP)
	}

	// Convert IPs to integers for iteration
	startInt := ipToInt(start)
	endInt := ipToInt(end)

	// Get all leased IPs in this range for this network
	leasedIPs, err := r.getLeasedIPsInRange(ctx, networkID, startInt, endInt)
	if err != nil {
		return "", err
	}

	// Find first available IP
	for ipInt := startInt; ipInt <= endInt; ipInt++ {
		ip := intToIP(ipInt).String()
		if !containsIP(leasedIPs, ip) {
			return ip, nil
		}
	}

	return "", nil // No available IPs in this range
}

func (r *ipLeaseRepositoryImpl) getLeasedIPsInRange(ctx context.Context, networkID int64, startInt, endInt uint32) ([]string, error) {
	// Get IPs from leases
	leaseQuery := `
		SELECT ip_address FROM ip_address_leases
		WHERE network_id = ?`

	leaseRows, err := r.db.QueryContext(ctx, leaseQuery, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get leased IPs: %w", err)
	}
	defer leaseRows.Close()

	var leasedIPs []string
	for leaseRows.Next() {
		var ip string
		if err := leaseRows.Scan(&ip); err != nil {
			return nil, fmt.Errorf("failed to scan leased IP: %w", err)
		}

		ipInt := ipToInt(net.ParseIP(ip))
		if ipInt >= startInt && ipInt <= endInt {
			leasedIPs = append(leasedIPs, ip)
		}
	}

	// Also get IPs from machines that have network_id set (dynamically assigned IPs)
	machineQuery := `
		SELECT ipv4 FROM machines
		WHERE ipv4 != ''`

	machineRows, err := r.db.QueryContext(ctx, machineQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine IPs: %w", err)
	}
	defer machineRows.Close()

	for machineRows.Next() {
		var ip string
		if err := machineRows.Scan(&ip); err != nil {
			return nil, fmt.Errorf("failed to scan machine IP: %w", err)
		}

		ipInt := ipToInt(net.ParseIP(ip))
		if ipInt >= startInt && ipInt <= endInt {
			leasedIPs = append(leasedIPs, ip)
		}
	}

	return leasedIPs, nil
}

// IP conversion utilities
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

func intToIP(ipInt uint32) net.IP {
	return net.IPv4(byte(ipInt>>24), byte(ipInt>>16), byte(ipInt>>8), byte(ipInt))
}

func containsIP(ips []string, target string) bool {
	for _, ip := range ips {
		if ip == target {
			return true
		}
	}
	return false
}
