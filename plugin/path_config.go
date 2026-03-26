package replicator

import (
	"context"
	"fmt"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	configStoragePath = "config"
)

type Configuration struct {
	VaultAddress         string          `json:"vault_address"`
	VaultMount           string          `json:"vault_mount"`
	AppRoleRoleID        string          `json:"approle_role_id"`
	AppRoleSecretID      string          `json:"approle_secret_id"`
	DestinationToken     string          `json:"destination_token"`
	DestinationMount     string          `json:"destination_mount"`
	OrgSkipList          []string        `json:"org_skip_list"`
	AllowDeletionSync    bool            `json:"allow_deletion_sync"`
	OrgDeletionOverrides map[string]bool `json:"org_deletion_overrides"`
}

// ShouldSyncOrg returns true if the organization should be synced
func (c *Configuration) ShouldSyncOrg(org string) bool {
	if c == nil {
		return true
	}
	for _, skipOrg := range c.OrgSkipList {
		if skipOrg == org {
			return false
		}
	}
	return true
}

// ShouldAllowDeletionSync returns true if deletion sync is allowed for the organization
func (c *Configuration) ShouldAllowDeletionSync(org string) bool {
	if c == nil {
		return false
	}
	// Check for org-specific override
	if override, exists := c.OrgDeletionOverrides[org]; exists {
		return override
	}
	// Fall back to global setting
	return c.AllowDeletionSync
}

func (b *Backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Read configuration",
				Description: "Returns the plugin configuration (token is masked)",
				Callback:    b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Summary:     "Create/Update configuration",
				Description: "Creates or updates the plugin configuration",
				Callback:    b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Summary:     "Update configuration",
				Description: "Updates the plugin configuration",
				Callback:    b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Summary:     "Delete configuration",
				Description: "Deletes the plugin configuration",
				Callback:    b.pathConfigDelete,
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
			"org_skip_list": {
				Type:        framework.TypeStringSlice,
				Description: "Organizations to skip during sync (blacklist)",
			},
			"allow_deletion_sync": {
				Type:        framework.TypeBool,
				Description: "Allow deletion sync (delete secrets in destination when deleted in source)",
			},
			"org_deletion_overrides": {
				Type:        framework.TypeMap,
				Description: "Per-organization deletion sync overrides (e.g., {\"org-1\": false})",
			},
		},
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
			"vault_address":          config.VaultAddress,
			"vault_mount":            config.VaultMount,
			"approle_role_id":        config.AppRoleRoleID,
			"approle_secret_id":      config.AppRoleSecretID,
			"destination_token":      config.DestinationToken,
			"destination_mount":      config.DestinationMount,
			"org_skip_list":          config.OrgSkipList,
			"allow_deletion_sync":    config.AllowDeletionSync,
			"org_deletion_overrides": config.OrgDeletionOverrides,
		},
	}, nil
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vaultAddress, ok := data.Get("vault_address").(string)
	if !ok {
		return logical.ErrorResponse("vault_address must be a string"), logical.ErrInvalidRequest
	}
	vaultMount, ok := data.Get("vault_mount").(string)
	if !ok {
		return logical.ErrorResponse("vault_mount must be a string"), logical.ErrInvalidRequest
	}
	approleRoleID, ok := data.Get("approle_role_id").(string)
	if !ok {
		return logical.ErrorResponse("approle_role_id must be a string"), logical.ErrInvalidRequest
	}
	approleSecretID, ok := data.Get("approle_secret_id").(string)
	if !ok {
		return logical.ErrorResponse("approle_secret_id must be a string"), logical.ErrInvalidRequest
	}
	destinationToken, ok := data.Get("destination_token").(string)
	if !ok {
		return logical.ErrorResponse("destination_token must be a string"), logical.ErrInvalidRequest
	}
	destinationMount, ok := data.Get("destination_mount").(string)
	if !ok {
		return logical.ErrorResponse("destination_mount must be a string"), logical.ErrInvalidRequest
	}

	orgSkipList, ok := data.Get("org_skip_list").([]string)
	if !ok {
		orgSkipList = []string{}
	}

	allowDeletionSync, ok := data.Get("allow_deletion_sync").(bool)
	if !ok {
		allowDeletionSync = false
	}

	orgDeletionOverridesRaw, ok := data.Get("org_deletion_overrides").(map[string]interface{})
	if !ok {
		orgDeletionOverridesRaw = map[string]interface{}{}
	}
	orgDeletionOverrides := make(map[string]bool)
	for k, v := range orgDeletionOverridesRaw {
		if boolVal, ok := v.(bool); ok {
			orgDeletionOverrides[k] = boolVal
		}
	}

	config := &Configuration{
		VaultAddress:         vaultAddress,
		VaultMount:           vaultMount,
		AppRoleRoleID:        approleRoleID,
		AppRoleSecretID:      approleSecretID,
		DestinationToken:     destinationToken,
		DestinationMount:     destinationMount,
		OrgSkipList:          orgSkipList,
		AllowDeletionSync:    allowDeletionSync,
		OrgDeletionOverrides: orgDeletionOverrides,
	}

	if err := ValidateConfig(config); err != nil {
		return logical.ErrorResponse("invalid configuration: " + err.Error()), logical.ErrInvalidRequest
	}

	if err := b.writeEncryptedConfig(ctx, req.Storage, config); err != nil {
		return nil, fmt.Errorf("failed to store configuration: %w", err)
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

	var secureConfig SecureConfig
	if err := entry.DecodeJSON(&secureConfig); err != nil {
		return nil, err
	}

	if secureConfig.AppRoleRoleID == "" && secureConfig.AppRoleSecretID == "" && secureConfig.DestinationToken == "" {
		var config Configuration
		if err := entry.DecodeJSON(&config); err != nil {
			return nil, err
		}
		return &config, nil
	}

	return b.decryptConfig(ctx, &secureConfig)
}
