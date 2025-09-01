package repository

import (
	"context"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

// SSHKeyRepository extends the generic Repository with SSH key-specific operations
type SSHKeyRepository interface {
	Repository[datastore.SSHKey, int64]

	// Domain-specific operations
	FindByMachineID(ctx context.Context, machineID int64) ([]datastore.SSHKey, error)
	CreateForMachine(ctx context.Context, machineID int64, keyText string) (*datastore.SSHKey, error)
}

// sshKeyRepositoryImpl implements SSHKeyRepository
type sshKeyRepositoryImpl struct {
	*DatastoreRepository[datastore.SSHKey, int64]
}

// NewSSHKeyRepository creates a new SSH key repository
func NewSSHKeyRepository(ds *datastore.Datastore) SSHKeyRepository {
	base := NewDatastoreRepository[datastore.SSHKey, int64](ds)
	return &sshKeyRepositoryImpl{
		DatastoreRepository: base,
	}
}

// Save creates or updates an SSH key
func (r *sshKeyRepositoryImpl) Save(ctx context.Context, entity datastore.SSHKey) (datastore.SSHKey, error) {
	// For SSH keys, we always create new ones (no updates)
	key, err := r.CreateForMachine(ctx, entity.MachineID, entity.KeyText)
	if err != nil {
		return datastore.SSHKey{}, err
	}
	return *key, nil
}

// FindByID retrieves an SSH key by its ID
func (r *sshKeyRepositoryImpl) FindByID(ctx context.Context, id int64) (datastore.SSHKey, error) {
	key, err := r.ds.GetSSHKey(id)
	if err != nil {
		return datastore.SSHKey{}, fmt.Errorf("failed to find SSH key: %w", err)
	}
	if key == nil {
		return datastore.SSHKey{}, fmt.Errorf("SSH key with ID %d: %w", id, ErrNotFound)
	}
	return *key, nil
}

// FindAll retrieves all SSH keys
func (r *sshKeyRepositoryImpl) FindAll(ctx context.Context) ([]datastore.SSHKey, error) {
	keys, err := r.ds.ListAllSSHKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to list all SSH keys: %w", err)
	}
	return keys, nil
}

// DeleteByID deletes an SSH key by its ID
func (r *sshKeyRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	err := r.ds.DeleteSSHKey(id)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}
	return nil
}

// ExistsByID checks if an SSH key exists by its ID
func (r *sshKeyRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	key, err := r.ds.GetSSHKey(id)
	if err != nil {
		return false, fmt.Errorf("failed to check SSH key existence: %w", err)
	}
	return key != nil, nil
}

// FindByMachineID retrieves all SSH keys for a specific machine
func (r *sshKeyRepositoryImpl) FindByMachineID(ctx context.Context, machineID int64) ([]datastore.SSHKey, error) {
	keys, err := r.ds.ListSSHKeys(machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys for machine %d: %w", machineID, err)
	}
	return keys, nil
}

// CreateForMachine creates a new SSH key for a specific machine
func (r *sshKeyRepositoryImpl) CreateForMachine(ctx context.Context, machineID int64, keyText string) (*datastore.SSHKey, error) {
	id, err := r.ds.CreateSSHKey(machineID, keyText)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key for machine %d: %w", machineID, err)
	}

	// Fetch the created key to return the full entity
	key, err := r.ds.GetSSHKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created SSH key: %w", err)
	}

	return key, nil
}
