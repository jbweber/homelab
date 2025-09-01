package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// DHCPRangeRepository defines domain-specific operations for DHCP ranges
type DHCPRangeRepository interface {
	Repository[domain.DHCPRange, int64]
	FindByNetworkID(ctx context.Context, networkID int64) ([]domain.DHCPRange, error)
}

// dhcpRangeRepositoryImpl implements DHCPRangeRepository
type dhcpRangeRepositoryImpl struct {
	db *sql.DB
}

// NewDHCPRangeRepository creates a new DHCP range repository
func NewDHCPRangeRepository(db *sql.DB) DHCPRangeRepository {
	return &dhcpRangeRepositoryImpl{
		db: db,
	}
}

// Save creates or updates a DHCP range
func (r *dhcpRangeRepositoryImpl) Save(ctx context.Context, dhcpRange domain.DHCPRange) (domain.DHCPRange, error) {
	if dhcpRange.ID == 0 {
		// Create new DHCP range
		return r.createDHCPRange(dhcpRange)
	} else {
		// Update existing DHCP range
		return r.updateDHCPRange(dhcpRange)
	}
}

// createDHCPRange inserts a new DHCP range into the database
func (r *dhcpRangeRepositoryImpl) createDHCPRange(d domain.DHCPRange) (domain.DHCPRange, error) {
	if d.NetworkID == 0 {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range network ID is required")
	}
	if d.StartIP == "" {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range start IP is required")
	}
	if d.EndIP == "" {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range end IP is required")
	}

	result, err := r.db.Exec(`
		INSERT INTO dhcp_ranges (network_id, start_ip, end_ip, lease_time)
		VALUES (?, ?, ?, ?)`,
		d.NetworkID, d.StartIP, d.EndIP, d.LeaseTime)
	if err != nil {
		return domain.DHCPRange{}, fmt.Errorf("failed to create DHCP range: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.DHCPRange{}, fmt.Errorf("failed to get DHCP range ID: %w", err)
	}

	d.ID = id
	return d, nil
}

// updateDHCPRange updates an existing DHCP range in the database
func (r *dhcpRangeRepositoryImpl) updateDHCPRange(d domain.DHCPRange) (domain.DHCPRange, error) {
	if d.NetworkID == 0 {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range network ID is required")
	}
	if d.StartIP == "" {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range start IP is required")
	}
	if d.EndIP == "" {
		return domain.DHCPRange{}, fmt.Errorf("DHCP range end IP is required")
	}

	_, err := r.db.Exec(`
		UPDATE dhcp_ranges
		SET network_id = ?, start_ip = ?, end_ip = ?, lease_time = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		d.NetworkID, d.StartIP, d.EndIP, d.LeaseTime, d.ID)
	if err != nil {
		return domain.DHCPRange{}, fmt.Errorf("failed to update DHCP range: %w", err)
	}

	return d, nil
}

// FindByID finds a DHCP range by ID
func (r *dhcpRangeRepositoryImpl) FindByID(ctx context.Context, id int64) (domain.DHCPRange, error) {
	var dhcpRange domain.DHCPRange
	err := r.db.QueryRow(`
		SELECT id, network_id, start_ip, end_ip, lease_time
		FROM dhcp_ranges WHERE id = ?`, id).Scan(
		&dhcpRange.ID, &dhcpRange.NetworkID, &dhcpRange.StartIP,
		&dhcpRange.EndIP, &dhcpRange.LeaseTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.DHCPRange{}, fmt.Errorf("DHCP range with ID %d not found", id)
		}
		return domain.DHCPRange{}, fmt.Errorf("failed to find DHCP range: %w", err)
	}
	return dhcpRange, nil
}

// FindAll finds all DHCP ranges
func (r *dhcpRangeRepositoryImpl) FindAll(ctx context.Context) ([]domain.DHCPRange, error) {
	rows, err := r.db.Query(`
		SELECT id, network_id, start_ip, end_ip, lease_time
		FROM dhcp_ranges ORDER BY network_id, start_ip`)
	if err != nil {
		return nil, fmt.Errorf("failed to find DHCP ranges: %w", err)
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

// DeleteByID deletes a DHCP range by ID
func (r *dhcpRangeRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.Exec("DELETE FROM dhcp_ranges WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete DHCP range: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DHCP range with ID %d not found", id)
	}

	return nil
}

// ExistsByID checks if a DHCP range exists by ID
func (r *dhcpRangeRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM dhcp_ranges WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check DHCP range existence: %w", err)
	}
	return count > 0, nil
}

// FindByNetworkID finds all DHCP ranges for a specific network
func (r *dhcpRangeRepositoryImpl) FindByNetworkID(ctx context.Context, networkID int64) ([]domain.DHCPRange, error) {
	rows, err := r.db.Query(`
		SELECT id, network_id, start_ip, end_ip, lease_time
		FROM dhcp_ranges WHERE network_id = ? ORDER BY start_ip`, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to find DHCP ranges for network %d: %w", networkID, err)
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
