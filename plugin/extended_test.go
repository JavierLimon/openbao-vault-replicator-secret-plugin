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

func TestShortCommit(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "abc1234", shortCommit("abc1234"))
	assert.Equal(t, "abc", shortCommit("abc"))
	assert.Equal(t, "", shortCommit(""))
	assert.Equal(t, "a", shortCommit("a"))
	assert.Equal(t, "1234567", shortCommit("1234567890"))
}

func TestParseVersionPart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"valid number", "123", 123},
		{"empty string", "", 0},
		{"invalid number", "abc", 0},
		{"negative number", "-5", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionPart(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
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
		wantError bool
		config    *Configuration
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

	secret := &SecretData{}

	assert.Empty(t, secret.CustomMetadata)
	assert.Equal(t, 0, secret.Version)
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

func TestMetricIncrementors(t *testing.T) {
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

	replicatorBackend.incrementTotalRequests()
	replicatorBackend.incrementTotalErrors()
	replicatorBackend.incrementSyncTotal()
	replicatorBackend.incrementSyncCompleted()
	replicatorBackend.incrementSyncFailed()
	replicatorBackend.incrementSecretsReplicated(10)

	resp, err := replicatorBackend.pathMetricsRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, int64(1), resp.Data["total_requests"])
	assert.Equal(t, int64(1), resp.Data["total_errors"])
	assert.Equal(t, int64(1), resp.Data["sync_total"])
	assert.Equal(t, int64(1), resp.Data["sync_completed"])
	assert.Equal(t, int64(1), resp.Data["sync_failed"])
	assert.Equal(t, int64(10), resp.Data["secrets_replicated"])
}

// Test Configuration helper functions

func TestShouldSyncOrg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Configuration
		org      string
		expected bool
	}{
		{
			name:     "nil config returns true",
			config:   nil,
			org:      "org-1",
			expected: true,
		},
		{
			name:     "empty skip list returns true",
			config:   &Configuration{},
			org:      "org-1",
			expected: true,
		},
		{
			name:     "org in skip list returns false",
			config:   &Configuration{OrgSkipList: []string{"org-1", "org-2"}},
			org:      "org-1",
			expected: false,
		},
		{
			name:     "org not in skip list returns true",
			config:   &Configuration{OrgSkipList: []string{"org-1", "org-2"}},
			org:      "org-3",
			expected: true,
		},
		{
			name:     "single org skip list - match",
			config:   &Configuration{OrgSkipList: []string{"skipped-org"}},
			org:      "skipped-org",
			expected: false,
		},
		{
			name:     "single org skip list - no match",
			config:   &Configuration{OrgSkipList: []string{"skipped-org"}},
			org:      "other-org",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldSyncOrg(tt.org)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldAllowDeletionSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Configuration
		org      string
		expected bool
	}{
		{
			name:     "nil config returns false",
			config:   nil,
			org:      "org-1",
			expected: false,
		},
		{
			name:     "global disabled no override",
			config:   &Configuration{AllowDeletionSync: false},
			org:      "org-1",
			expected: false,
		},
		{
			name:     "global enabled no override",
			config:   &Configuration{AllowDeletionSync: true},
			org:      "org-1",
			expected: true,
		},
		{
			name:     "org override true",
			config:   &Configuration{AllowDeletionSync: false, OrgDeletionOverrides: map[string]bool{"org-1": true}},
			org:      "org-1",
			expected: true,
		},
		{
			name:     "org override false",
			config:   &Configuration{AllowDeletionSync: true, OrgDeletionOverrides: map[string]bool{"org-1": false}},
			org:      "org-1",
			expected: false,
		},
		{
			name:     "org override not set falls back to global true",
			config:   &Configuration{AllowDeletionSync: true, OrgDeletionOverrides: map[string]bool{"other-org": false}},
			org:      "org-1",
			expected: true,
		},
		{
			name:     "org override not set falls back to global false",
			config:   &Configuration{AllowDeletionSync: false, OrgDeletionOverrides: map[string]bool{"other-org": true}},
			org:      "org-1",
			expected: false,
		},
		{
			name:     "empty overrides falls back to global",
			config:   &Configuration{AllowDeletionSync: true, OrgDeletionOverrides: map[string]bool{}},
			org:      "org-1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldAllowDeletionSync(tt.org)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test retry configuration helpers

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	config := DefaultRetryConfig()
	require.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, config.InitialInterval)
	assert.Equal(t, 10*time.Second, config.MaxInterval)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestDefaultListOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultListOptions()
	require.NotNil(t, opts)
	assert.Equal(t, 100, opts.PageSize)
}

func TestListOptionsWithPageSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		inputPageSize int
		expectedSize  int
	}{
		{
			name:          "valid size",
			inputPageSize: 50,
			expectedSize:  50,
		},
		{
			name:          "zero size uses default",
			inputPageSize: 0,
			expectedSize:  100,
		},
		{
			name:          "negative size uses default",
			inputPageSize: -1,
			expectedSize:  100,
		},
		{
			name:          "exceeds max uses max",
			inputPageSize: 10000,
			expectedSize:  500,
		},
		{
			name:          "max size exact",
			inputPageSize: 500,
			expectedSize:  500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ListOptions{}
			result := opts.WithPageSize(tt.inputPageSize)
			assert.Equal(t, tt.expectedSize, result.PageSize)
		})
	}
}

// Test sync history read endpoints

func TestPathSyncHistoryRead(t *testing.T) {
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

	// First save a history entry
	status := &SyncStatus{
		StartedAt:           time.Now().UTC(),
		Status:              "completed",
		OrganizationsSynced: 2,
		SyncedSecrets:       20,
		CompletedAt:         time.Now().UTC(),
	}

	err = replicatorBackend.saveSyncHistory(ctx, storage, status)
	require.NoError(t, err)

	// List to get the key
	respList, err := replicatorBackend.pathSyncHistoryList(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, respList)

	keys, ok := respList.Data["keys"].([]string)
	require.True(t, ok)
	require.NotEmpty(t, keys)

	// Now read the specific history entry
	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"timestamp": keys[0],
		},
		Schema: map[string]*framework.FieldSchema{
			"timestamp": {Type: framework.TypeString},
		},
	}

	resp, err := replicatorBackend.pathSyncHistoryRead(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "completed", resp.Data["status"])
}

func TestPathSyncHistoryTimestampRead(t *testing.T) {
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

	// First save a history entry
	status := &SyncStatus{
		StartedAt:           time.Now().UTC(),
		Status:              "completed",
		OrganizationsSynced: 3,
		SyncedSecrets:       30,
		CompletedAt:         time.Now().UTC(),
	}

	err = replicatorBackend.saveSyncHistory(ctx, storage, status)
	require.NoError(t, err)

	// List to get the key
	respList, err := replicatorBackend.pathSyncHistoryList(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, respList)

	keys, ok := respList.Data["keys"].([]string)
	require.True(t, ok)
	require.NotEmpty(t, keys)

	// Now read using timestamp endpoint
	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"timestamp": keys[0],
		},
		Schema: map[string]*framework.FieldSchema{
			"timestamp": {Type: framework.TypeString},
		},
	}

	resp, err := replicatorBackend.pathSyncHistoryTimestampRead(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "completed", resp.Data["status"])
	assert.Equal(t, 3, resp.Data["organizations_synced"])
}

// Test encryption Decrypt function

func TestEncrypterDecrypt(t *testing.T) {
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

	encrypter := NewEncrypter(replicatorBackend)

	// First encrypt some data
	originalText := "Hello, World!"
	encrypted, err := encrypter.Encrypt(originalText)
	require.NoError(t, err)

	// Now decrypt it
	decrypted, err := encrypter.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, originalText, decrypted)
}

func TestEncrypterDecryptEmpty(t *testing.T) {
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

	encrypter := NewEncrypter(replicatorBackend)

	// Encrypt empty string
	encrypted, err := encrypter.Encrypt("")
	require.NoError(t, err)

	decrypted, err := encrypter.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestEncrypterDecryptInvalid(t *testing.T) {
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

	encrypter := NewEncrypter(replicatorBackend)

	// Test invalid base64
	_, err = encrypter.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)

	// Test too short ciphertext
	_, err = encrypter.Decrypt("YWJj") // "abc" in base64
	assert.Error(t, err)
}

// Test decryptConfig function

func TestDecryptConfig(t *testing.T) {
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

	// Create encrypter and encrypt config
	encrypter := NewEncrypter(replicatorBackend)

	// First encrypt the sensitive fields
	roleID, err := encrypter.Encrypt("my-role-id")
	require.NoError(t, err)
	secretID, err := encrypter.Encrypt("my-secret-id")
	require.NoError(t, err)
	token, err := encrypter.Encrypt("my-token")
	require.NoError(t, err)

	// Create secure config
	secureConfig := &SecureConfig{
		VaultAddress:     "https://vault.example.com",
		VaultMount:       "kv2",
		AppRoleRoleID:    roleID,
		AppRoleSecretID:  secretID,
		DestinationToken: token,
		DestinationMount: "kv2",
	}

	// Now decrypt
	config, err := replicatorBackend.decryptConfig(ctx, secureConfig)
	require.NoError(t, err)

	assert.Equal(t, "https://vault.example.com", config.VaultAddress)
	assert.Equal(t, "kv2", config.VaultMount)
	assert.Equal(t, "my-role-id", config.AppRoleRoleID)
	assert.Equal(t, "my-secret-id", config.AppRoleSecretID)
	assert.Equal(t, "my-token", config.DestinationToken)
	assert.Equal(t, "kv2", config.DestinationMount)
}

func TestPathConfigWriteWithSkipList(t *testing.T) {
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

	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"vault_address":          "https://vault.example.com",
			"vault_mount":            "kv2",
			"approle_role_id":        "role-id",
			"approle_secret_id":      "secret-id",
			"destination_token":      "dest-token",
			"destination_mount":      "kv2",
			"org_skip_list":          []string{"org-1", "org-2"},
			"allow_deletion_sync":    true,
			"org_deletion_overrides": map[string]interface{}{"org-3": false},
		},
		Schema: map[string]*framework.FieldSchema{
			"vault_address":          {Type: framework.TypeString},
			"vault_mount":            {Type: framework.TypeString},
			"approle_role_id":        {Type: framework.TypeString},
			"approle_secret_id":      {Type: framework.TypeString},
			"destination_token":      {Type: framework.TypeString},
			"destination_mount":      {Type: framework.TypeString},
			"org_skip_list":          {Type: framework.TypeStringSlice},
			"allow_deletion_sync":    {Type: framework.TypeBool},
			"org_deletion_overrides": {Type: framework.TypeMap},
		},
	}

	resp, err := replicatorBackend.pathConfigWrite(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)
	assert.Nil(t, resp)
}

func TestPathConfigDelete(t *testing.T) {
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

	data := &framework.FieldData{
		Raw: map[string]interface{}{
			"vault_address":          "https://vault.example.com",
			"vault_mount":            "kv2",
			"approle_role_id":        "role-id",
			"approle_secret_id":      "secret-id",
			"destination_token":      "dest-token",
			"destination_mount":      "kv2",
			"org_skip_list":          "",
			"allow_deletion_sync":    false,
			"org_deletion_overrides": map[string]interface{}{},
		},
		Schema: map[string]*framework.FieldSchema{
			"vault_address":          {Type: framework.TypeString},
			"vault_mount":            {Type: framework.TypeString},
			"approle_role_id":        {Type: framework.TypeString},
			"approle_secret_id":      {Type: framework.TypeString},
			"destination_token":      {Type: framework.TypeString},
			"destination_mount":      {Type: framework.TypeString},
			"org_skip_list":          {Type: framework.TypeString},
			"allow_deletion_sync":    {Type: framework.TypeBool},
			"org_deletion_overrides": {Type: framework.TypeMap},
		},
	}

	_, err = replicatorBackend.pathConfigWrite(ctx, &logical.Request{
		Storage: storage,
	}, data)
	require.NoError(t, err)

	resp, err := replicatorBackend.pathConfigDelete(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	assert.Nil(t, resp)

	respRead, err := replicatorBackend.pathConfigRead(ctx, &logical.Request{
		Storage: storage,
	}, &framework.FieldData{})
	require.NoError(t, err)
	assert.Nil(t, respRead)
}

func TestRetryWithBackoffNonRetryableError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries:      3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      2.0,
	}

	callCount := 0
	err := RetryWithBackoff(ctx, config, func() error {
		callCount++
		return fmt.Errorf("access denied")
	})

	require.Error(t, err)
	assert.Equal(t, 1, callCount)
	assert.Contains(t, err.Error(), "access denied")
}

func TestRetryWithBackoffSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries:      3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      2.0,
	}

	callCount := 0
	err := RetryWithBackoff(ctx, config, func() error {
		callCount++
		if callCount < 2 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRetryWithBackoffContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	config := &RetryConfig{
		MaxRetries:      10,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     200 * time.Millisecond,
		Multiplier:      2.0,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := RetryWithBackoff(ctx, config, func() error {
		return fmt.Errorf("connection refused")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestRetryWithBackoffMaxRetries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries:      2,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      2.0,
	}

	err := RetryWithBackoff(ctx, config, func() error {
		return fmt.Errorf("connection refused")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries")
}

func TestLoginToVaultInvalidCredentials(t *testing.T) {
	t.Parallel()

	client, err := NewVaultClient("https://127.0.0.1:1", "nonexistent-mount", "invalid-role", "invalid-secret")
	require.NoError(t, err)
	require.NotNil(t, client)

	_, err = LoginToVault(client, "invalid-role", "invalid-secret")
	require.Error(t, err)
}

func TestBackendFactory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      hclog.NewNullLogger(),
	})
	require.NoError(t, err)
	require.NotNil(t, backend)

	be, ok := backend.(*Backend)
	require.True(t, ok)
	assert.NotNil(t, be.Logger)
}

func TestBackendFactoryNilLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := &logical.InmemStorage{}

	backend, err := Factory(ctx, &logical.BackendConfig{
		StorageView: storage,
		Logger:      nil,
	})
	require.NoError(t, err)
	require.NotNil(t, backend)
}
