package replicator

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathHealthReadExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	resp, err := replicatorBackend.pathHealthRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "ok", resp.Data["status"])
	assert.Equal(t, "vault-replicator", resp.Data["plugin"])
}

func TestPathMetricsReadExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	resp, err := replicatorBackend.pathMetricsRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, int64(0), resp.Data["total_requests"])
	assert.Equal(t, int64(0), resp.Data["total_errors"])
	assert.Equal(t, int64(0), resp.Data["sync_total"])
	assert.Equal(t, int64(0), resp.Data["sync_completed"])
	assert.Equal(t, int64(0), resp.Data["sync_failed"])
	assert.Equal(t, int64(0), resp.Data["secrets_replicated"])
}

func TestVersionExtended(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "1.0.0", GetVersion())
	assert.Equal(t, "", GetCommit())
	assert.Equal(t, "", GetDate())
	assert.Equal(t, "", GetBuildType())
	assert.Equal(t, "", GetBranch())

	info := GetVersionInfo()
	assert.Equal(t, "1.0.0", info.Version)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.Platform)

	versionStr := GetVersionString()
	assert.Contains(t, versionStr, "1.0.0")

	shortStr := GetShortVersionString()
	assert.Equal(t, "1.0.0", shortStr)

	assert.True(t, IsVersionAtLeast("1.0.0"))
	assert.True(t, IsVersionAtLeast("0.9.0"))
	assert.False(t, IsVersionAtLeast("1.0.1"))
	assert.False(t, IsVersionAtLeast("2.0.0"))
}

func TestCompareVersionsExtended(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, compareVersions("1.0.0", "1.0.0"))
	assert.Equal(t, 0, compareVersions("1", "1"))
	assert.Equal(t, 1, compareVersions("1.0.1", "1.0.0"))
	assert.Equal(t, 1, compareVersions("2.0.0", "1.0.0"))
	assert.Equal(t, -1, compareVersions("1.0.0", "1.0.1"))
	assert.Equal(t, -1, compareVersions("1.0.0", "2.0.0"))
}

func TestAuditLoggerExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)
	logger := NewAuditLogger(replicatorBackend)

	err = logger.LogSyncStarted(ctx, storage, []string{"org-1", "org-2"}, "req-123")
	require.NoError(t, err)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestAuditLoggerEmptyOrgs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)
	logger := NewAuditLogger(replicatorBackend)

	err = logger.LogSyncStarted(ctx, storage, []string{}, "req-123")
	require.NoError(t, err)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestAuditLoggerCompleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)
	logger := NewAuditLogger(replicatorBackend)

	err = logger.LogSyncCompleted(ctx, storage, "req-123", 5, 100, 5000)
	require.NoError(t, err)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestAuditLoggerFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)
	logger := NewAuditLogger(replicatorBackend)

	err = logger.LogSyncFailed(ctx, storage, "req-123", fmt.Errorf("connection failed"))
	require.NoError(t, err)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestAuditLoggerListEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)
	logger := NewAuditLogger(replicatorBackend)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestPathAuditLogsListExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	logger := NewAuditLogger(replicatorBackend)
	err = logger.LogSyncStarted(ctx, storage, []string{"org-1"}, "req-1")
	require.NoError(t, err)

	resp, err := replicatorBackend.pathAuditLogsList(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	keys, ok := resp.Data["keys"].([]string)
	require.True(t, ok)
	assert.Len(t, keys, 1)
}

func TestPathAuditLogReadExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	logger := NewAuditLogger(replicatorBackend)
	err = logger.LogSyncStarted(ctx, storage, []string{"org-1"}, "req-123")
	require.NoError(t, err)

	keys, err := logger.ListAuditLogs(ctx, storage)
	require.NoError(t, err)
	require.NotEmpty(t, keys)

	timestamp := strings.TrimPrefix(keys[0], auditLogPrefix)

	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"timestamp": timestamp,
		},
		Schema: map[string]*framework.FieldSchema{
			"timestamp": {Type: framework.TypeString},
		},
	}

	resp, err := replicatorBackend.pathAuditLogRead(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, 1, resp.Data["org_count"])
}

func TestValidateURLExtended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{"valid https", "https://vault.example.com:8200", false},
		{"valid http", "http://vault.example.com:8200", false},
		{"valid localhost", "http://localhost:8200", false},
		{"missing scheme", "vault.example.com:8200", true},
		{"invalid scheme", "ftp://vault.example.com", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url, "vault_address")
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMountPathExtended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"valid kv2", "kv2", false},
		{"valid secret", "secret", false},
		{"valid with slash", "kv2/", false},
		{"valid custom", "my-secrets", false},
		{"valid empty", "", false},
		{"with space", "kv 2", true},
		{"starting space", " kv2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMountPath(tt.path)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsRetryableErrorExtended(t *testing.T) {
	t.Parallel()

	assert.True(t, isRetryableError(fmt.Errorf("connection refused")))
	assert.True(t, isRetryableError(fmt.Errorf("dial tcp: connection refused")))
	assert.True(t, isRetryableError(fmt.Errorf("i/o timeout")))

	assert.False(t, isRetryableError(fmt.Errorf("access denied")))
	assert.False(t, isRetryableError(fmt.Errorf("not found")))
	assert.False(t, isRetryableError(nil))
}

func TestValidateConfigExtended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    *Configuration
		wantError bool
	}{
		{
			name: "valid config",
			config: &Configuration{
				VaultAddress:     "https://vault.example.com",
				VaultMount:       "kv2",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
				DestinationMount: "kv2",
			},
			wantError: false,
		},
		{
			name: "missing vault address",
			config: &Configuration{
				VaultMount:       "kv2",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
				DestinationMount: "kv2",
			},
			wantError: true,
		},
		{
			name: "missing vault mount",
			config: &Configuration{
				VaultAddress:     "https://vault.example.com",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
				DestinationMount: "kv2",
			},
			wantError: true,
		},
		{
			name: "missing destination token",
			config: &Configuration{
				VaultAddress:     "https://vault.example.com",
				VaultMount:       "kv2",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationMount: "kv2",
			},
			wantError: true,
		},
		{
			name: "missing destination mount",
			config: &Configuration{
				VaultAddress:     "https://vault.example.com",
				VaultMount:       "kv2",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
			},
			wantError: true,
		},
		{
			name: "invalid vault address",
			config: &Configuration{
				VaultAddress:     "not-a-url",
				VaultMount:       "kv2",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
				DestinationMount: "kv2",
			},
			wantError: true,
		},
		{
			name: "invalid vault mount",
			config: &Configuration{
				VaultAddress:     "https://vault.example.com",
				VaultMount:       "with space",
				AppRoleRoleID:    "role-id",
				AppRoleSecretID:  "secret-id",
				DestinationToken: "dest-token",
				DestinationMount: "kv2",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretDataExtended(t *testing.T) {
	t.Parallel()

	secret := &SecretData{
		Data: map[string]interface{}{
			"password": "secret123",
			"username": "admin",
		},
		CustomMetadata: map[string]interface{}{
			"owner": "team-dba",
			"env":   "production",
		},
		Version: 5,
	}

	assert.Equal(t, "secret123", secret.Data["password"])
	assert.Equal(t, "admin", secret.Data["username"])
	assert.Equal(t, "team-dba", secret.CustomMetadata["owner"])
	assert.Equal(t, "production", secret.CustomMetadata["env"])
	assert.Equal(t, 5, secret.Version)
}

func TestSecretDataEmptyMetadata(t *testing.T) {
	t.Parallel()

	secret := &SecretData{
		Data:    map[string]interface{}{"key": "value"},
		Version: 1,
	}

	assert.Empty(t, secret.CustomMetadata)
}

func TestSyncStatusStorageExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	status := &SyncStatus{
		StartedAt:           time.Now().UTC(),
		Status:              "completed",
		OrganizationsSynced: 5,
		SyncedSecrets:       100,
		Failed:              0,
		CompletedAt:         time.Now().UTC(),
		DurationSeconds:     30,
	}

	err = replicatorBackend.saveSyncStatus(ctx, storage, status)
	require.NoError(t, err)

	readStatus, err := replicatorBackend.readSyncStatus(ctx, storage)
	require.NoError(t, err)
	require.NotNil(t, readStatus)
	assert.Equal(t, "completed", readStatus.Status)
	assert.Equal(t, 5, readStatus.OrganizationsSynced)
	assert.Equal(t, 100, readStatus.SyncedSecrets)
}

func TestSyncStatusStorageEmptyExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	readStatus, err := replicatorBackend.readSyncStatus(ctx, storage)
	require.NoError(t, err)
	assert.Nil(t, readStatus)
}

func TestPathSyncStatusReadExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	status := &SyncStatus{
		StartedAt:           time.Now().UTC(),
		Status:              "completed",
		OrganizationsSynced: 3,
		SyncedSecrets:       50,
		Failed:              1,
		CompletedAt:         time.Now().UTC(),
		DurationSeconds:     15,
	}

	err = replicatorBackend.saveSyncStatus(ctx, storage, status)
	require.NoError(t, err)

	resp, err := replicatorBackend.pathSyncStatusRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "completed", resp.Data["status"])
}

func TestPathSyncHistoryListExtended(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)

	replicatorBackend, ok := backend.(*Backend)
	require.True(t, ok)

	status := &SyncStatus{
		StartedAt:           time.Now().UTC(),
		Status:              "completed",
		OrganizationsSynced: 2,
		SyncedSecrets:       20,
		CompletedAt:         time.Now().UTC(),
	}

	err = replicatorBackend.saveSyncHistory(ctx, storage, status)
	require.NoError(t, err)

	resp, err := replicatorBackend.pathSyncHistoryList(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	keys, ok := resp.Data["keys"].([]string)
	require.True(t, ok)
	assert.NotEmpty(t, keys)
}
