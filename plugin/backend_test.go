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
			"vault_address":       "https://vault.example.com",
			"vault_mount":         "kv2",
			"approle_role_id":     "test-role-id",
			"approle_secret_id":   "test-secret-id",
			"destination_token":   "test-token",
			"destination_mount":   "kv2",
			"org_skip_list":       []string{},
			"allow_deletion_sync": false,
		},
		Schema: pathConfigSchema(),
	}

	resp, err := replicatorBackend.pathConfigWrite(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	assert.Nil(t, resp)
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

	_, err = replicatorBackend.pathSyncSecrets(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.Error(t, err)
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
	}

	assert.Equal(t, "https://vault.example.com", config.VaultAddress)
	assert.Equal(t, "kv2", config.VaultMount)
	assert.Equal(t, "role-id", config.AppRoleRoleID)
	assert.Equal(t, "secret-id", config.AppRoleSecretID)
	assert.Equal(t, "dest-token", config.DestinationToken)
	assert.Equal(t, "kv2", config.DestinationMount)
}

func TestConfigStoragePath(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "config", configStoragePath)
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
		"org_skip_list": {
			Type:        framework.TypeStringSlice,
			Description: "Organizations to skip",
		},
		"allow_deletion_sync": {
			Type:        framework.TypeBool,
			Description: "Allow deletion sync",
		},
		"org_deletion_overrides": {
			Type:        framework.TypeMap,
			Description: "Per-org deletion overrides",
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
