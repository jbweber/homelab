package repository

import (
	"context"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

// MachineRepository defines domain-specific operations for machines
type MachineRepository interface {
	Repository[datastore.Machine, int64]
	FindByName(ctx context.Context, name string) (datastore.Machine, error)
	FindByIPv4(ctx context.Context, ipv4 string) (datastore.Machine, error)
}

// machineRepositoryImpl implements MachineRepository
type machineRepositoryImpl struct {
	*DatastoreRepository[datastore.Machine, int64]
}

// NewMachineRepository creates a new machine repository
func NewMachineRepository(ds *datastore.Datastore) MachineRepository {
	base := NewDatastoreRepository[datastore.Machine, int64](ds)
	return &machineRepositoryImpl{
		DatastoreRepository: base,
	}
}

// Save creates or updates a machine
func (r *machineRepositoryImpl) Save(ctx context.Context, machine datastore.Machine) (datastore.Machine, error) {
	if machine.ID == 0 {
		// Create new machine
		created, err := r.ds.CreateMachine(machine)
		if err != nil {
			return datastore.Machine{}, fmt.Errorf("failed to create machine: %w", err)
		}
		return created, nil
	} else {
		// Update existing machine
		updated, err := r.ds.UpdateMachine(machine)
		if err != nil {
			return datastore.Machine{}, fmt.Errorf("failed to update machine: %w", err)
		}
		return updated, nil
	}
}

// FindByID retrieves a machine by its ID
func (r *machineRepositoryImpl) FindByID(ctx context.Context, id int64) (datastore.Machine, error) {
	machine, err := r.ds.GetMachine(id)
	if err != nil {
		return datastore.Machine{}, fmt.Errorf("failed to find machine: %w", err)
	}
	if machine == nil {
		return datastore.Machine{}, fmt.Errorf("machine with ID %d: %w", id, ErrNotFound)
	}
	return *machine, nil
}

// FindAll retrieves all machines
func (r *machineRepositoryImpl) FindAll(ctx context.Context) ([]datastore.Machine, error) {
	machines, err := r.ds.ListMachines()
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	return machines, nil
}

// DeleteByID removes a machine by its ID
func (r *machineRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	err := r.ds.DeleteMachine(id)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}
	return nil
}

// ExistsByID checks if a machine exists by its ID
func (r *machineRepositoryImpl) ExistsByID(ctx context.Context, id int64) (bool, error) {
	machine, err := r.ds.GetMachine(id)
	if err != nil {
		return false, fmt.Errorf("failed to check machine existence: %w", err)
	}
	return machine != nil, nil
}

// FindByName retrieves a machine by its name
func (r *machineRepositoryImpl) FindByName(ctx context.Context, name string) (datastore.Machine, error) {
	machine, err := r.ds.GetMachineByName(name)
	if err != nil {
		return datastore.Machine{}, fmt.Errorf("failed to find machine by name: %w", err)
	}
	if machine == nil {
		return datastore.Machine{}, fmt.Errorf("machine with name %s: %w", name, ErrNotFound)
	}
	return *machine, nil
}

// FindByIPv4 retrieves a machine by its IPv4 address
func (r *machineRepositoryImpl) FindByIPv4(ctx context.Context, ipv4 string) (datastore.Machine, error) {
	machine, err := r.ds.GetMachineByIPv4(ipv4)
	if err != nil {
		return datastore.Machine{}, fmt.Errorf("failed to find machine by IPv4: %w", err)
	}
	if machine == nil {
		return datastore.Machine{}, fmt.Errorf("machine with IPv4 %s: %w", ipv4, ErrNotFound)
	}
	return *machine, nil
}
