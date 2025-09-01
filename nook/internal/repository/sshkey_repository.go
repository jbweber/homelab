package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/domain"
)

// SSHKeyRepository extends the generic Repository with SSH key-specific operations
type SSHKeyRepository interface {
	Repository[domain.SSHKey, int64]

	// Domain-specific operations
	FindByMachineID(ctx context.Context, machineID int64) ([]domain.SSHKey, error)
	CreateForMachine(ctx context.Context, machineID int64, keyText string) (*domain.SSHKey, error)
}

// sshKeyRepositoryImpl implements SSHKeyRepository
type sshKeyRepositoryImpl struct {
	db *sql.DB
}

// NewSSHKeyRepository creates a new SSH key repository
func NewSSHKeyRepository(db *sql.DB) SSHKeyRepository {
	return &sshKeyRepositoryImpl{
		db: db,
	}
}

// Save creates or updates an SSH key
func (r *sshKeyRepositoryImpl) Save(ctx context.Context, entity domain.SSHKey) (domain.SSHKey, error) {
	// For SSH keys, we always create new ones (no updates)
	key, err := r.CreateForMachine(ctx, entity.MachineID, entity.KeyText)
	if err != nil {
		return domain.SSHKey{}, err
	}
	return *key, nil
}

// FindByID retrieves an SSH key by its ID
func (r *sshKeyRepositoryImpl) FindByID(ctx context.Context, id int64) (domain.SSHKey, error) {
	var k domain.SSHKey
	err := r.db.QueryRow("SELECT id, machine_id, key_text FROM ssh_keys WHERE id = ?", id).Scan(&k.ID, &k.MachineID, &k.KeyText)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.SSHKey{}, fmt.Errorf("SSH key with ID %d: %w", id, ErrNotFound)
		}
		return domain.SSHKey{}, fmt.Errorf("failed to find SSH key: %w", err)
	}
	return k, nil
}

// FindAll retrieves all SSH keys
func (r *sshKeyRepositoryImpl) FindAll(ctx context.Context) ([]domain.SSHKey, error) {
	rows, err := r.db.Query("SELECT id, machine_id, key_text FROM ssh_keys ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to list all SSH keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.SSHKey
	for rows.Next() {
		var k domain.SSHKey
		if err := rows.Scan(&k.ID, &k.MachineID, &k.KeyText); err != nil {
			return nil, fmt.Errorf("failed to scan SSH key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// DeleteByID deletes an SSH key by its ID
func (r *sshKeyRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.Exec("DELETE FROM ssh_keys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}
	return nil
}

// ExistsByID checks if an SSH key exists by its ID
func (r *sshKeyRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM ssh_keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check SSH key existence: %w", err)
	}
	return count > 0, nil
}

// FindByMachineID retrieves all SSH keys for a specific machine
func (r *sshKeyRepositoryImpl) FindByMachineID(ctx context.Context, machineID int64) ([]domain.SSHKey, error) {
	rows, err := r.db.Query("SELECT id, machine_id, key_text FROM ssh_keys WHERE machine_id = ? ORDER BY id ASC", machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys for machine %d: %w", machineID, err)
	}
	defer rows.Close()

	var keys []domain.SSHKey
	for rows.Next() {
		var k domain.SSHKey
		if err := rows.Scan(&k.ID, &k.MachineID, &k.KeyText); err != nil {
			return nil, fmt.Errorf("failed to scan SSH key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// CreateForMachine creates a new SSH key for a specific machine
func (r *sshKeyRepositoryImpl) CreateForMachine(ctx context.Context, machineID int64, keyText string) (*domain.SSHKey, error) {
	res, err := r.db.Exec("INSERT INTO ssh_keys (machine_id, key_text) VALUES (?, ?)", machineID, keyText)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key for machine %d: %w", machineID, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// Fetch the created key to return the full entity
	var k domain.SSHKey
	err = r.db.QueryRow("SELECT id, machine_id, key_text FROM ssh_keys WHERE id = ?", id).Scan(&k.ID, &k.MachineID, &k.KeyText)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created SSH key: %w", err)
	}

	return &k, nil
}
