package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/jbweber/homelab/nook/internal/domain"
	"github.com/jbweber/homelab/nook/internal/repository"
)

// ListMachines implements MachinesStore interface
func (a *API) ListMachines() ([]Machine, error) {
	machines, err := a.machineRepo.FindAll(context.Background())
	if err != nil {
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	var result []Machine
	for _, m := range machines {
		result = append(result, Machine{
			ID:        m.ID,
			Name:      m.Name,
			Hostname:  m.Hostname,
			IPv4:      m.IPv4,
			NetworkID: m.NetworkID,
		})
	}
	return result, nil
}

// CreateMachine implements MachinesStore interface
func (a *API) CreateMachine(m Machine) (Machine, error) {
	// Convert api.Machine to domain.Machine
	domainMachine := domain.Machine{
		ID:        m.ID,
		Name:      m.Name,
		Hostname:  m.Hostname,
		IPv4:      m.IPv4,
		NetworkID: m.NetworkID,
	}
	saved, err := a.machineRepo.Save(context.Background(), domainMachine)
	if err != nil {
		return Machine{}, err
	}

	// If network_id is provided but no IPv4, allocate IP after machine creation
	if m.NetworkID != nil && m.IPv4 == "" {
		lease, err := a.ipLeaseRepo.AllocateIPAddress(context.Background(), saved.ID, *m.NetworkID)
		if err != nil {
			// If IP allocation fails, delete the machine and return error
			if deleteErr := a.machineRepo.DeleteByID(context.Background(), saved.ID); deleteErr != nil {
				fmt.Printf("Warning: failed to delete machine after IP allocation failure: %v\n", deleteErr)
			}
			return Machine{}, fmt.Errorf("failed to allocate IP address: %w", err)
		}
		// Update the machine with the allocated IP
		saved.IPv4 = lease.IPAddress
		updated, err := a.machineRepo.Save(context.Background(), saved)
		if err != nil {
			// If update fails, deallocate the IP
			if deallocErr := a.ipLeaseRepo.DeallocateIPAddress(context.Background(), saved.ID, *m.NetworkID); deallocErr != nil {
				fmt.Printf("Warning: failed to deallocate IP after machine update failure: %v\n", deallocErr)
			}
			return Machine{}, err
		}
		saved = updated
	}

	// Convert back to api.Machine
	return Machine{
		ID:        saved.ID,
		Name:      saved.Name,
		Hostname:  saved.Hostname,
		IPv4:      saved.IPv4,
		NetworkID: saved.NetworkID,
	}, nil
}

// GetMachine implements MachinesStore interface
func (a *API) GetMachine(id int64) (*Machine, error) {
	machine, err := a.machineRepo.FindByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      machine.IPv4,
		NetworkID: machine.NetworkID,
	}, nil
}

// DeleteMachine implements MachinesStore interface
func (a *API) DeleteMachine(id int64) error {
	// First, get the machine to check if it has a network-allocated IP
	machine, err := a.machineRepo.FindByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // Machine doesn't exist, consider it deleted
		}
		return err
	}

	// If the machine has a network_id and IPv4, deallocate the IP
	if machine.NetworkID != nil && machine.IPv4 != "" {
		if deallocErr := a.ipLeaseRepo.DeallocateIPAddress(context.Background(), machine.ID, *machine.NetworkID); deallocErr != nil {
			// Log the error but don't fail the deletion
			fmt.Printf("Warning: failed to deallocate IP for machine %d: %v\n", machine.ID, deallocErr)
		}
	}

	// Delete the machine
	return a.machineRepo.DeleteByID(context.Background(), id)
}

// GetMachineByName implements MachinesStore interface
func (a *API) GetMachineByName(name string) (*Machine, error) {
	machine, err := a.machineRepo.FindByName(context.Background(), name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      machine.IPv4,
		NetworkID: machine.NetworkID,
	}, nil
}

// AllocateIPAddress implements MachinesStore interface
func (a *API) AllocateIPAddress(machineID, networkID int64) (string, error) {
	lease, err := a.ipLeaseRepo.AllocateIPAddress(context.Background(), machineID, networkID)
	if err != nil {
		return "", err
	}
	return lease.IPAddress, nil
}

// DeallocateIPAddress implements MachinesStore interface
func (a *API) DeallocateIPAddress(machineID, networkID int64) error {
	return a.ipLeaseRepo.DeallocateIPAddress(context.Background(), machineID, networkID)
}

// GetMachineByIPv4 implements MetaDataStore interface
func (a *API) GetMachineByIPv4(ipv4 string) (*Machine, error) {
	machine, err := a.machineRepo.FindByIPv4(context.Background(), ipv4)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// Convert domain.Machine to api.Machine
	return &Machine{
		ID:        machine.ID,
		Name:      machine.Name,
		Hostname:  machine.Hostname,
		IPv4:      machine.IPv4,
		NetworkID: machine.NetworkID,
	}, nil
}
