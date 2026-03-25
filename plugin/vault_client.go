package replicator

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	address  string
	mount    string
	roleID   string
	secretID string
	token    string
	client   *api.Client
}

func NewVaultClient(addr, mount, roleID, secretID string) (*VaultClient, error) {
	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultClient{
		address:  addr,
		mount:    mount,
		roleID:   roleID,
		secretID: secretID,
		client:   client,
	}, nil
}

func (c *VaultClient) Login() error {
	resp, err := c.client.Logical().Write("auth/approle/login", map[string]interface{}{
		"role_id":   c.roleID,
		"secret_id": c.secretID,
	})
	if err != nil {
		return fmt.Errorf("approle login failed: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("approle login returned nil response")
	}

	tokenRaw, ok := resp.Data["token"]
	if !ok {
		return fmt.Errorf("token not found in approle login response")
	}

	token, ok := tokenRaw.(string)
	if !ok {
		return fmt.Errorf("token is not a string")
	}

	c.token = token
	c.client.SetToken(c.token)
	return nil
}

func (c *VaultClient) ListOrganizations() ([]string, error) {
	path := fmt.Sprintf("%s/metadata/", c.mount)
	resp, err := c.client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	if resp == nil {
		return nil, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return nil, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("keys is not a []interface{}")
	}

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func (c *VaultClient) ListSecretsInOrganization(org string) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s", c.mount, org)
	resp, err := c.client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in organization %s: %w", org, err)
	}
	if resp == nil {
		return nil, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return nil, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("keys is not a []interface{}")
	}

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func (c *VaultClient) ReadSecret(org, secret string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/data/%s/%s", c.mount, org, secret)
	resp, err := c.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s/%s: %w", org, secret, err)
	}
	if resp == nil {
		return nil, nil
	}

	dataRaw, ok := resp.Data["data"]
	if !ok {
		return nil, nil
	}

	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]interface{}")
	}

	return data, nil
}

func (c *VaultClient) Token() string {
	return c.token
}

func (c *VaultClient) Address() string {
	return c.address
}

func (c *VaultClient) Mount() string {
	return c.mount
}
