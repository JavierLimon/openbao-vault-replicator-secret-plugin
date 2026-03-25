package replicator

import (
	"context"
	"fmt"

	"github.com/openbao/openbao/sdk/v2/logical"
)

type VaultClient interface {
	Login(ctx context.Context, roleID, secretID string) (string, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

type vaultClient struct {
	address string
	mount   string
}

func NewVaultClient(address, mount string) VaultClient {
	return &vaultClient{
		address: address,
		mount:   mount,
	}
}

func (c *vaultClient) Login(ctx context.Context, roleID, secretID string) (string, error) {
	if roleID == "" {
		return "", fmt.Errorf("role_id is required")
	}
	if secretID == "" {
		return "", fmt.Errorf("secret_id is required")
	}
	return "mock-token", nil
}

func (c *vaultClient) ListSecrets(ctx context.Context, path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	return []string{}, nil
}

func (c *vaultClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	return nil, nil
}

type VaultClientInterface interface {
	Login(ctx context.Context, roleID, secretID string) (string, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	Close() error
}

type SyncStatus struct {
	LastSync string            `json:"last_sync"`
	Status   string            `json:"status"`
	Orgs     map[string]string `json:"organizations,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type SyncHistory struct {
	Operations []SyncOperation `json:"operations"`
}

type SyncOperation struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Orgs      []string       `json:"organizations"`
	Status    string         `json:"status"`
	Stats     map[string]int `json:"stats,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type SyncRequest struct {
	Organizations []string `json:"organizations,omitempty"`
	DryRun        bool     `json:"dry_run"`
}

type SyncResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message,omitempty"`
	Stats   map[string]int `json:"stats,omitempty"`
	Errors  []string       `json:"errors,omitempty"`
}

type ReplicationStore interface {
	GetConfig(ctx context.Context, storage logical.Storage) (*Configuration, error)
	SaveConfig(ctx context.Context, storage logical.Storage, config *Configuration) error
	GetSyncStatus(ctx context.Context, storage logical.Storage) (*SyncStatus, error)
	SaveSyncStatus(ctx context.Context, storage logical.Storage, status *SyncStatus) error
	GetSyncHistory(ctx context.Context, storage logical.Storage) (*SyncHistory, error)
	AddSyncOperation(ctx context.Context, storage logical.Storage, op SyncOperation) error
}
