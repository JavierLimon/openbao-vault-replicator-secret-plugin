package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	configStoragePath = "config"
)

// Configuration holds the plugin configuration
type Configuration struct {
	VaultAddress       string 
	VaultMount         string 
	AppRoleRoleID      string 
	AppRoleSecretID    string 
	DestinationToken  string 
	DestinationMount  string 
	OrganizationPath  string 
}

// pathConfig returns the config path
func (b *Backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern:      "config",
		Operations: map[logical.Operation]*framework.OperationHandler{
			logical.ReadOperation: &framework.OperationHandler{
				Summary:     "Read configuration",
				Description: "Returns the plugin configuration (token is masked)",
			},
			logical.CreateOperation: &framework.OperationHandler{
				Summary:     "Create/Update configuration",
				Description: "Creates or updates the plugin configuration",
			},
			logical.UpdateOperation: &framework.OperationHandler{
				Summary:     "Update configuration",
				Description: "Updates the plugin configuration",
			},
			logical.DeleteOperation: &framework.OperationHandler{
				Summary:     "Delete configuration",
				Description: "Deletes the plugin configuration",
			},
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
		ExistenceCheck: b.HandleExistenceCheck,
		HelpSynopsis:    "Configuration endpoint for Vault Replicator plugin",
		HelpDescription: "Configure the connection to Vault and OpenBao for secret replication",
	}
}

// pathConfigRead handles reading the configuration
func (b *Backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.readConfig(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	// Mask sensitive fields
	config.AppRoleSecretID = ""
	config.DestinationToken = "[MASKED]"

	return &logical.Response{
		Data: map[string]interface{}{
			"vault_address":       config.VaultAddress,
			"vault_mount":         config.VaultMount,
			"approle_role_id":    config.AppRoleRoleID,
			"approle_secret_id":  config.AppRoleSecretID,
			"destination_token":  config.DestinationToken,
			"destination_mount":  config.DestinationMount,
			"organization_path":  config.OrganizationPath,
		},
	}, nil
}

// pathConfigWrite handles writing the configuration
func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config := &Configuration{
		VaultAddress:      data.Get("vault_address").(string),
		VaultMount:        data.Get("vault_mount").(string),
		AppRoleRoleID:     data.Get("approle_role_id").(string),
		AppRoleSecretID:   data.Get("approle_secret_id").(string),
		DestinationToken:  data.Get("destination_token").(string),
		DestinationMount:  data.Get("destination_mount").(string),
		OrganizationPath:  data.Get("organization_path").(string),
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

// pathConfigDelete handles deleting the configuration
func (b *Backend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configStoragePath); err != nil {
		return nil, err
	}
	return nil, nil
}

// readConfig reads the configuration from storage
func (b *Backend) readConfig(ctx context.Context) (*Configuration, error) {
	entry, err := req.Storage.Get(ctx, configStoragePath)
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
