package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// MachineRepository defines domain-specific operations for machines
type MachineRepository interface {
	Repository[domain.Machine, int64]
	FindByName(ctx context.Context, name string) (domain.Machine, error)
	FindByIPv4(ctx context.Context, ipv4 string) (domain.Machine, error)
}

// machineRepositoryImpl implements MachineRepository
type machineRepositoryImpl struct {
	db *sql.DB
}

// NewMachineRepository creates a new machine repository
func NewMachineRepository(db *sql.DB) MachineRepository {
	return &machineRepositoryImpl{
		db: db,
	}
}

// Save creates or updates a machine
func (r *machineRepositoryImpl) Save(ctx context.Context, machine domain.Machine) (domain.Machine, error) {
	if machine.ID == 0 {
		// Create new machine
		return r.createMachine(machine)
	} else {
		// Update existing machine
		return r.updateMachine(machine)
	}
}

// createMachine inserts a new machine into the database
func (r *machineRepositoryImpl) createMachine(m domain.Machine) (domain.Machine, error) {
	if m.Name == "" {
		return domain.Machine{}, fmt.Errorf("machine name is required")
	}
	if m.Hostname == "" {
		return domain.Machine{}, fmt.Errorf("machine hostname is required")
	}
	if m.IPv4 == "" {
		return domain.Machine{}, fmt.Errorf("machine IPv4 is required")
	}
	res, err := r.db.Exec("INSERT INTO machines (name, hostname, ipv4) VALUES (?, ?, ?)", m.Name, m.Hostname, m.IPv4)
	if err != nil {
		return domain.Machine{}, fmt.Errorf("failed to create machine: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Machine{}, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	m.ID = id
	return m, nil
}

// updateMachine updates an existing machine's details by ID
func (r *machineRepositoryImpl) updateMachine(m domain.Machine) (domain.Machine, error) {
	if m.ID == 0 {
		return domain.Machine{}, fmt.Errorf("machine ID is required")
	}
	if m.Name == "" {
		return domain.Machine{}, fmt.Errorf("machine name is required")
	}
	if m.Hostname == "" {
		return domain.Machine{}, fmt.Errorf("machine hostname is required")
	}
	if m.IPv4 == "" {
		return domain.Machine{}, fmt.Errorf("machine IPv4 is required")
	}
	_, err := r.db.Exec("UPDATE machines SET name = ?, hostname = ?, ipv4 = ? WHERE id = ?", m.Name, m.Hostname, m.IPv4, m.ID)
	if err != nil {
		return domain.Machine{}, fmt.Errorf("failed to update machine: %w", err)
	}
	// Return the updated machine
	return m, nil
}

// FindByID retrieves a machine by its ID
func (r *machineRepositoryImpl) FindByID(ctx context.Context, id int64) (domain.Machine, error) {
	var m domain.Machine
	err := r.db.QueryRow("SELECT id, name, hostname, ipv4 FROM machines WHERE id = ?", id).Scan(&m.ID, &m.Name, &m.Hostname, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Machine{}, fmt.Errorf("machine with ID %d: %w", id, ErrNotFound)
		}
		return domain.Machine{}, fmt.Errorf("failed to find machine: %w", err)
	}
	return m, nil
}

// FindAll retrieves all machines
func (r *machineRepositoryImpl) FindAll(ctx context.Context) ([]domain.Machine, error) {
	rows, err := r.db.Query("SELECT id, name, hostname, ipv4 FROM machines")
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the operation
		}
	}()

	var machines []domain.Machine
	for rows.Next() {
		var m domain.Machine
		if err := rows.Scan(&m.ID, &m.Name, &m.Hostname, &m.IPv4); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		machines = append(machines, m)
	}
	return machines, nil
}

// DeleteByID removes a machine by its ID
func (r *machineRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.Exec("DELETE FROM machines WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}
	return nil
}

// ExistsByID checks if a machine exists by its ID
func (r *machineRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM machines WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check machine existence: %w", err)
	}
	return count > 0, nil
}

// FindByName retrieves a machine by its name
func (r *machineRepositoryImpl) FindByName(ctx context.Context, name string) (domain.Machine, error) {
	var m domain.Machine
	err := r.db.QueryRow("SELECT id, name, hostname, ipv4 FROM machines WHERE name = ?", name).Scan(&m.ID, &m.Name, &m.Hostname, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Machine{}, fmt.Errorf("machine with name %s: %w", name, ErrNotFound)
		}
		return domain.Machine{}, fmt.Errorf("failed to find machine by name: %w", err)
	}
	return m, nil
}

// FindByIPv4 retrieves a machine by its IPv4 address
func (r *machineRepositoryImpl) FindByIPv4(ctx context.Context, ipv4 string) (domain.Machine, error) {
	var m domain.Machine
	err := r.db.QueryRow("SELECT id, name, hostname, ipv4 FROM machines WHERE ipv4 = ?", ipv4).Scan(&m.ID, &m.Name, &m.Hostname, &m.IPv4)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Machine{}, fmt.Errorf("machine with IPv4 %s: %w", ipv4, ErrNotFound)
		}
		return domain.Machine{}, fmt.Errorf("failed to find machine by IPv4: %w", err)
	}
	return m, nil
}
