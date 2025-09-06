package api

import (
	"context"
	"errors"
	"testing"

	"github.com/jbweber/homelab/nook/internal/domain"
)

type mockSSHKeyRepo struct {
	sshKeys []domain.SSHKey
	err     error
}

func (m *mockSSHKeyRepo) Save(ctx context.Context, key domain.SSHKey) (domain.SSHKey, error) {
	return domain.SSHKey{}, errors.New("not implemented")
}

func (m *mockSSHKeyRepo) FindByID(ctx context.Context, id int64) (domain.SSHKey, error) {
	return domain.SSHKey{}, errors.New("not implemented")
}

func (m *mockSSHKeyRepo) FindAll(ctx context.Context) ([]domain.SSHKey, error) {
	return m.sshKeys, m.err
}

func (m *mockSSHKeyRepo) DeleteByID(ctx context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	// Find and remove the key
	for i, key := range m.sshKeys {
		if key.ID == id {
			m.sshKeys = append(m.sshKeys[:i], m.sshKeys[i+1:]...)
			return nil
		}
	}
	return nil // Key not found, but don't error
}

func (m *mockSSHKeyRepo) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockSSHKeyRepo) FindByMachineID(ctx context.Context, machineID int64) ([]domain.SSHKey, error) {
	return []domain.SSHKey{}, errors.New("not implemented")
}

func (m *mockSSHKeyRepo) CreateForMachine(ctx context.Context, machineID int64, keyText string) (*domain.SSHKey, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := domain.SSHKey{
		ID:        int64(len(m.sshKeys) + 1),
		MachineID: machineID,
		KeyText:   keyText,
	}
	m.sshKeys = append(m.sshKeys, key)
	return &key, nil
}

func TestAPI_ListAllSSHKeys_Success(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{
		sshKeys: []domain.SSHKey{
			{ID: 1, MachineID: 1, KeyText: "ssh-rsa AAAAB3NzaC1yc2E..."},
			{ID: 2, MachineID: 2, KeyText: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."},
		},
	}

	api := &API{sshKeyRepo: mockRepo}

	keys, err := api.ListAllSSHKeys()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	if keys[0].ID != 1 || keys[0].MachineID != 1 {
		t.Errorf("Unexpected first key: %+v", keys[0])
	}
}

func TestAPI_ListAllSSHKeys_Empty(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{sshKeys: []domain.SSHKey{}}
	api := &API{sshKeyRepo: mockRepo}

	keys, err := api.ListAllSSHKeys()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}

func TestAPI_ListAllSSHKeys_Error(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{err: errors.New("repository error")}
	api := &API{sshKeyRepo: mockRepo}

	keys, err := api.ListAllSSHKeys()
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	if keys != nil {
		t.Errorf("Expected nil keys on error, got %v", keys)
	}
}

func TestAPI_CreateSSHKey_Success(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{sshKeys: []domain.SSHKey{}}
	api := &API{sshKeyRepo: mockRepo}

	key, err := api.CreateSSHKey(1, "ssh-rsa AAAAB3NzaC1yc2E...")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key == nil {
		t.Fatal("Expected key, got nil")
	}

	if key.MachineID != 1 {
		t.Errorf("Expected MachineID 1, got %d", key.MachineID)
	}

	if key.KeyText != "ssh-rsa AAAAB3NzaC1yc2E..." {
		t.Errorf("Expected key text to match, got %s", key.KeyText)
	}

	if key.ID == 0 {
		t.Error("Expected ID to be assigned")
	}
}

func TestAPI_CreateSSHKey_Error(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{err: errors.New("repository error")}
	api := &API{sshKeyRepo: mockRepo}

	key, err := api.CreateSSHKey(1, "ssh-rsa AAAAB3NzaC1yc2E...")
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	if key != nil {
		t.Errorf("Expected nil key on error, got %v", key)
	}
}

func TestAPI_DeleteSSHKey_Success(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{
		sshKeys: []domain.SSHKey{
			{ID: 1, MachineID: 1, KeyText: "ssh-rsa AAAAB3NzaC1yc2E..."},
		},
	}
	api := &API{sshKeyRepo: mockRepo}

	err := api.DeleteSSHKey(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify key was removed
	if len(mockRepo.sshKeys) != 0 {
		t.Errorf("Expected key to be removed, but %d keys remain", len(mockRepo.sshKeys))
	}
}

func TestAPI_DeleteSSHKey_NotFound(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{sshKeys: []domain.SSHKey{}}
	api := &API{sshKeyRepo: mockRepo}

	// Deleting non-existent key should not error
	err := api.DeleteSSHKey(999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent key, got %v", err)
	}
}

func TestAPI_DeleteSSHKey_Error(t *testing.T) {
	mockRepo := &mockSSHKeyRepo{err: errors.New("repository error")}
	api := &API{sshKeyRepo: mockRepo}

	err := api.DeleteSSHKey(1)
	if err == nil {
		t.Fatal("Expected error, got none")
	}
}
