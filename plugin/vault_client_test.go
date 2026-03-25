package replicator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVaultClient struct {
	loginFunc       func(ctx context.Context, roleID, secretID string) (string, error)
	listSecretsFunc func(ctx context.Context, path string) ([]string, error)
	readSecretFunc  func(ctx context.Context, path string) (map[string]interface{}, error)
}

func (m *mockVaultClient) Login(ctx context.Context, roleID, secretID string) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, roleID, secretID)
	}
	return "mock-token", nil
}

func (m *mockVaultClient) ListSecrets(ctx context.Context, path string) ([]string, error) {
	if m.listSecretsFunc != nil {
		return m.listSecretsFunc(ctx, path)
	}
	return []string{}, nil
}

func (m *mockVaultClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if m.readSecretFunc != nil {
		return m.readSecretFunc(ctx, path)
	}
	return nil, nil
}

func (m *mockVaultClient) Close() error {
	return nil
}

func TestNewVaultClient(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")
	require.NotNil(t, client)
	assert.IsType(t, &vaultClient{}, client)
}

func TestVaultClient_Login(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	token, err := client.Login(context.Background(), "test-role-id", "test-secret-id")
	require.NoError(t, err)
	assert.Equal(t, "mock-token", token)
}

func TestVaultClient_Login_EmptyRoleID(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	token, err := client.Login(context.Background(), "", "test-secret-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role_id is required")
	assert.Empty(t, token)
}

func TestVaultClient_Login_EmptySecretID(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	token, err := client.Login(context.Background(), "test-role-id", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret_id is required")
	assert.Empty(t, token)
}

func TestVaultClient_ListSecrets(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	secrets, err := client.ListSecrets(context.Background(), "data/orgs")
	require.NoError(t, err)
	assert.NotNil(t, secrets)
}

func TestVaultClient_ListSecrets_EmptyPath(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	secrets, err := client.ListSecrets(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
	assert.Nil(t, secrets)
}

func TestVaultClient_ReadSecret(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	secret, err := client.ReadSecret(context.Background(), "data/orgs/test-org")
	require.NoError(t, err)
	assert.Nil(t, secret)
}

func TestVaultClient_ReadSecret_EmptyPath(t *testing.T) {
	t.Parallel()

	client := NewVaultClient("https://vault.example.com", "kv2")

	secret, err := client.ReadSecret(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
	assert.Nil(t, secret)
}

func TestMockVaultClient_Login(t *testing.T) {
	t.Parallel()

	mock := &mockVaultClient{
		loginFunc: func(ctx context.Context, roleID, secretID string) (string, error) {
			if roleID == "fail-role" {
				return "", errors.New("login failed")
			}
			return "custom-token", nil
		},
	}

	token, err := mock.Login(context.Background(), "test-role", "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "custom-token", token)

	token, err = mock.Login(context.Background(), "fail-role", "test-secret")
	require.Error(t, err)
	assert.Empty(t, token)
}

func TestMockVaultClient_ListSecrets(t *testing.T) {
	t.Parallel()

	expectedSecrets := []string{"org-1", "org-2", "org-3"}
	mock := &mockVaultClient{
		listSecretsFunc: func(ctx context.Context, path string) ([]string, error) {
			return expectedSecrets, nil
		},
	}

	secrets, err := mock.ListSecrets(context.Background(), "data/orgs")
	require.NoError(t, err)
	assert.Equal(t, expectedSecrets, secrets)
}

func TestMockVaultClient_ReadSecret(t *testing.T) {
	t.Parallel()

	expectedSecret := map[string]interface{}{
		"data": map[string]interface{}{
			"username": "admin",
			"password": "secret",
		},
	}
	mock := &mockVaultClient{
		readSecretFunc: func(ctx context.Context, path string) (map[string]interface{}, error) {
			return expectedSecret, nil
		},
	}

	secret, err := mock.ReadSecret(context.Background(), "data/orgs/test-org")
	require.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestMockVaultClient_ReadSecret_Error(t *testing.T) {
	t.Parallel()

	mock := &mockVaultClient{
		readSecretFunc: func(ctx context.Context, path string) (map[string]interface{}, error) {
			return nil, errors.New("secret not found")
		},
	}

	secret, err := mock.ReadSecret(context.Background(), "data/orgs/nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
	assert.Nil(t, secret)
}

func TestSyncStatus_Structure(t *testing.T) {
	t.Parallel()

	status := SyncStatus{
		LastSync: "2024-01-01T00:00:00Z",
		Status:   "success",
		Orgs: map[string]string{
			"org-1": "synced",
			"org-2": "failed",
		},
		Error: "",
	}

	assert.Equal(t, "2024-01-01T00:00:00Z", status.LastSync)
	assert.Equal(t, "success", status.Status)
	assert.Equal(t, "synced", status.Orgs["org-1"])
	assert.Equal(t, "failed", status.Orgs["org-2"])
	assert.Empty(t, status.Error)
}

func TestSyncOperation_Structure(t *testing.T) {
	t.Parallel()

	op := SyncOperation{
		ID:        "op-123",
		Timestamp: "2024-01-01T00:00:00Z",
		Orgs:      []string{"org-1", "org-2"},
		Status:    "completed",
		Stats: map[string]int{
			"succeeded": 10,
			"failed":    2,
		},
		Error: "",
	}

	assert.Equal(t, "op-123", op.ID)
	assert.Equal(t, "2024-01-01T00:00:00Z", op.Timestamp)
	assert.Len(t, op.Orgs, 2)
	assert.Equal(t, "completed", op.Status)
	assert.Equal(t, 10, op.Stats["succeeded"])
	assert.Equal(t, 2, op.Stats["failed"])
}

func TestSyncRequest_Structure(t *testing.T) {
	t.Parallel()

	req := SyncRequest{
		Organizations: []string{"org-1", "org-2"},
		DryRun:        true,
	}

	assert.Len(t, req.Organizations, 2)
	assert.True(t, req.DryRun)
}

func TestSyncResponse_Structure(t *testing.T) {
	t.Parallel()

	resp := SyncResponse{
		Status:  "success",
		Message: "Sync completed",
		Stats: map[string]int{
			"succeeded": 5,
		},
		Errors: []string{},
	}

	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "Sync completed", resp.Message)
	assert.Equal(t, 5, resp.Stats["succeeded"])
	assert.Empty(t, resp.Errors)
}

func TestSyncHistory_Structure(t *testing.T) {
	t.Parallel()

	history := SyncHistory{
		Operations: []SyncOperation{
			{
				ID:     "op-1",
				Status: "completed",
			},
			{
				ID:     "op-2",
				Status: "failed",
			},
		},
	}

	assert.Len(t, history.Operations, 2)
	assert.Equal(t, "completed", history.Operations[0].Status)
	assert.Equal(t, "failed", history.Operations[1].Status)
}
