package api

import (
	"context"
)

// ListAllSSHKeys implements SSHKeysStore interface
func (a *API) ListAllSSHKeys() ([]SSHKey, error) {
	keys, err := a.sshKeyRepo.FindAll(context.Background())
	if err != nil {
		return nil, err
	}
	// Convert domain.SSHKey to api.SSHKey
	var result []SSHKey
	for _, k := range keys {
		result = append(result, SSHKey{
			ID:        k.ID,
			MachineID: k.MachineID,
			KeyText:   k.KeyText,
		})
	}
	return result, nil
}

// CreateSSHKey implements SSHKeysStore interface
func (a *API) CreateSSHKey(machineID int64, keyText string) (*SSHKey, error) {
	key, err := a.sshKeyRepo.CreateForMachine(context.Background(), machineID, keyText)
	if err != nil {
		return nil, err
	}
	// Convert domain.SSHKey to api.SSHKey
	return &SSHKey{
		ID:        key.ID,
		MachineID: key.MachineID,
		KeyText:   key.KeyText,
	}, nil
}

// DeleteSSHKey implements SSHKeysStore interface
func (a *API) DeleteSSHKey(id int64) error {
	return a.sshKeyRepo.DeleteByID(context.Background(), id)
}
