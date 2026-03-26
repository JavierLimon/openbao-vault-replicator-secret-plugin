package replicator

import (
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVaultClient(t *testing.T) {
	t.Parallel()

	client, err := NewVaultClient("https://vault.example.com", "kv2", "role-id", "secret-id")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.IsType(t, &api.Client{}, client)
}

func TestLoginToVault_InvalidClient(t *testing.T) {
	t.Parallel()

	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, client)

	_, err = LoginToVault(client, "", "")
	require.Error(t, err)
}
