package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	configStoragePath = "config"
)

type Configuration struct {
	VaultAddress     string `json:"vault_address"`
	VaultMount       string `json:"vault_mount"`
	AppRoleRoleID    string `json:"approle_role_id"`
	AppRoleSecretID  string `json:"approle_secret_id"`
	DestinationToken string `json:"destination_token"`
	DestinationMount string `json:"destination_mount"`
	OrganizationPath string `json:"organization_path"`
}

func (b *Backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigRead,
			logical.CreateOperation: b.pathConfigWrite,
			logical.UpdateOperation: b.pathConfigWrite,
			logical.DeleteOperation: b.pathConfigDelete,
		},
		Fields: map[string]*framework.FieldSchema{
			"vault_address": {
				Type:        framework.TypeString,
				Description: "Vault server URL (e.g., https://vault.example.com:8200)",
			},
			"vault_mount": {
				Type:        framework.TypeString,
				Description: "Vault KVv2 mount path (default: kv2)",
			},
			"approle_role_id": {
				Type:        framework.TypeString,
				Description: "AppRole role_id",
			},
			"approle_secret_id": {
				Type:        framework.TypeString,
				Description: "AppRole secret_id",
			},
			"destination_token": {
				Type:        framework.TypeString,
				Description: "OpenBao token to write secrets",
			},
			"destination_mount": {
				Type:        framework.TypeString,
				Description: "OpenBao KVv2 mount (default: kv2)",
			},
			"organization_path": {
				Type:        framework.TypeString,
				Description: "Path in Vault where orgs live (e.g., data/)",
			},
		},
		ExistenceCheck:  configExistenceCheck,
		HelpSynopsis:    "Configuration endpoint for Vault Replicator plugin",
		HelpDescription: "Configure the connection to Vault and OpenBao for secret replication",
	}
}

func (b *Backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.readConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	config.AppRoleSecretID = ""
	config.DestinationToken = "[MASKED]"

	return &logical.Response{
		Data: map[string]interface{}{
			"vault_address":     config.VaultAddress,
			"vault_mount":       config.VaultMount,
			"approle_role_id":   config.AppRoleRoleID,
			"approle_secret_id": config.AppRoleSecretID,
			"destination_token": config.DestinationToken,
			"destination_mount": config.DestinationMount,
			"organization_path": config.OrganizationPath,
		},
	}, nil
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config := &Configuration{
		VaultAddress:     data.Get("vault_address").(string),
		VaultMount:       data.Get("vault_mount").(string),
		AppRoleRoleID:    data.Get("approle_role_id").(string),
		AppRoleSecretID:  data.Get("approle_secret_id").(string),
		DestinationToken: data.Get("destination_token").(string),
		DestinationMount: data.Get("destination_mount").(string),
		OrganizationPath: data.Get("organization_path").(string),
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *Backend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configStoragePath); err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *Backend) readConfig(ctx context.Context, storage logical.Storage) (*Configuration, error) {
	entry, err := storage.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var config Configuration
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
