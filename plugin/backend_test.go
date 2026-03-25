package replicator

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackend_Factory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	conf := &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	}

	backend, err := Factory(ctx, conf)
	require.NoError(t, err)
	require.NotNil(t, backend)
	assert.NotNil(t, backend)
}

func TestPathConfig_Read(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	resp, err := replicatorBackend.pathConfigRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	assert.Nil(t, resp, "response should be nil when config does not exist")

	config := &Configuration{
		VaultAddress:     "https://vault.example.com",
		VaultMount:       "kv2",
		AppRoleRoleID:    "test-role-id",
		AppRoleSecretID:  "test-secret-id",
		DestinationToken: "test-token",
		DestinationMount: "kv2",
		OrganizationPath: "data/",
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	require.NoError(t, err)
	err = storage.Put(ctx, entry)
	require.NoError(t, err)

	resp, err = replicatorBackend.pathConfigRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "https://vault.example.com", resp.Data["vault_address"])
	assert.Equal(t, "kv2", resp.Data["vault_mount"])
	assert.Equal(t, "test-role-id", resp.Data["approle_role_id"])
	assert.Equal(t, "", resp.Data["approle_secret_id"], "secret should be masked")
	assert.Equal(t, "[MASKED]", resp.Data["destination_token"], "token should be masked")
}

func TestPathConfig_Write(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"vault_address":     "https://vault.example.com",
			"vault_mount":       "kv2",
			"approle_role_id":   "test-role-id",
			"approle_secret_id": "test-secret-id",
			"destination_token": "test-token",
			"destination_mount": "kv2",
			"organization_path": "data/",
		},
		Schema: pathConfigSchema(),
	}

	resp, err := replicatorBackend.pathConfigWrite(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	assert.Nil(t, resp)

	entry, err := storage.Get(ctx, configStoragePath)
	require.NoError(t, err)
	require.NotNil(t, entry)

	var config Configuration
	err = entry.DecodeJSON(&config)
	require.NoError(t, err)

	assert.Equal(t, "https://vault.example.com", config.VaultAddress)
	assert.Equal(t, "kv2", config.VaultMount)
	assert.Equal(t, "test-role-id", config.AppRoleRoleID)
	assert.Equal(t, "test-secret-id", config.AppRoleSecretID)
	assert.Equal(t, "test-token", config.DestinationToken)
	assert.Equal(t, "kv2", config.DestinationMount)
	assert.Equal(t, "data/", config.OrganizationPath)
}

func TestPathConfig_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	config := &Configuration{
		VaultAddress:     "https://vault.example.com",
		VaultMount:       "kv2",
		AppRoleRoleID:    "test-role-id",
		DestinationMount: "kv2",
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	require.NoError(t, err)
	err = storage.Put(ctx, entry)
	require.NoError(t, err)

	resp, err := replicatorBackend.pathConfigDelete(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	assert.Nil(t, resp)

	entry, err = storage.Get(ctx, configStoragePath)
	require.NoError(t, err)
	assert.Nil(t, entry, "config should be deleted")
}

func TestReadConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	config, err := replicatorBackend.readConfig(ctx, storage)
	require.NoError(t, err)
	assert.Nil(t, config, "config should be nil when storage is empty")

	storeConfig := &Configuration{
		VaultAddress:    "https://vault.example.com",
		VaultMount:      "kv2",
		AppRoleRoleID:   "role-id",
		AppRoleSecretID: "secret-id",
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, storeConfig)
	require.NoError(t, err)
	err = storage.Put(ctx, entry)
	require.NoError(t, err)

	config, err = replicatorBackend.readConfig(ctx, storage)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "https://vault.example.com", config.VaultAddress)
}

func TestPathSync_Secrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"organizations": []string{"org-1", "org-2"},
			"dry_run":       true,
		},
		Schema: pathSyncSchema(),
	}

	resp, err := replicatorBackend.pathSyncSecrets(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "not implemented yet", resp.Data["status"])
}

func TestPathSync_Definition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	path := replicatorBackend.pathSync()
	assert.NotNil(t, path)
	assert.Equal(t, "sync/secrets", path.Pattern)
	assert.Contains(t, path.Callbacks, logical.CreateOperation)
}

func TestPathConfig_Definition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend := backend.(*Backend)

	path := replicatorBackend.pathConfig()
	assert.NotNil(t, path)
	assert.Equal(t, "config", path.Pattern)
	assert.Contains(t, path.Callbacks, logical.ReadOperation)
	assert.Contains(t, path.Callbacks, logical.CreateOperation)
	assert.Contains(t, path.Callbacks, logical.UpdateOperation)
	assert.Contains(t, path.Callbacks, logical.DeleteOperation)
	assert.NotNil(t, path.ExistenceCheck, "ExistenceCheck should be set")
}

func TestConfiguration_Fields(t *testing.T) {
	t.Parallel()

	config := Configuration{
		VaultAddress:     "https://vault.example.com",
		VaultMount:       "kv2",
		AppRoleRoleID:    "role-id",
		AppRoleSecretID:  "secret-id",
		DestinationToken: "dest-token",
		DestinationMount: "kv2",
		OrganizationPath: "data/",
	}

	assert.Equal(t, "https://vault.example.com", config.VaultAddress)
	assert.Equal(t, "kv2", config.VaultMount)
	assert.Equal(t, "role-id", config.AppRoleRoleID)
	assert.Equal(t, "secret-id", config.AppRoleSecretID)
	assert.Equal(t, "dest-token", config.DestinationToken)
	assert.Equal(t, "kv2", config.DestinationMount)
	assert.Equal(t, "data/", config.OrganizationPath)
}

func TestConfigStoragePath(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "config", configStoragePath)
}

func TestConfigExistenceCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	exists, err := configExistenceCheck(ctx, &logical.Request{
		Storage: storage,
	}, nil)
	require.NoError(t, err)
	assert.False(t, exists, "config should not exist when storage is empty")

	config := &Configuration{
		VaultAddress:     "https://vault.example.com",
		VaultMount:       "kv2",
		AppRoleRoleID:    "test-role-id",
		DestinationMount: "kv2",
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	require.NoError(t, err)
	err = storage.Put(ctx, entry)
	require.NoError(t, err)

	exists, err = configExistenceCheck(ctx, &logical.Request{
		Storage: storage,
	}, nil)
	require.NoError(t, err)
	assert.True(t, exists, "config should exist after being written")
}

func pathConfigSchema() map[string]*framework.FieldSchema {
	return map[string]*framework.FieldSchema{
		"vault_address": {
			Type:        framework.TypeString,
			Description: "Vault server URL",
		},
		"vault_mount": {
			Type:        framework.TypeString,
			Description: "Vault KVv2 mount path",
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
			Description: "OpenBao token",
		},
		"destination_mount": {
			Type:        framework.TypeString,
			Description: "OpenBao KVv2 mount",
		},
		"organization_path": {
			Type:        framework.TypeString,
			Description: "Path in Vault where orgs live",
		},
	}
}

func pathSyncSchema() map[string]*framework.FieldSchema {
	return map[string]*framework.FieldSchema{
		"organizations": {
			Type:        framework.TypeStringSlice,
			Description: "Specific organizations to sync",
		},
		"dry_run": {
			Type:        framework.TypeBool,
			Description: "Preview only",
		},
	}
}
