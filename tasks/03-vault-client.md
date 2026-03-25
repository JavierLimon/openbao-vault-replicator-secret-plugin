# Vault Client Tasks - openbao-vault-replicator-secret-plugin

Vault client with AppRole authentication and secret listing.

## Status
- Total: 2
- Completed: 0
- Pending: 2

---

## T-006: Vault Client - AppRole Auth

**Priority**: HIGH | **Status**: pending

Create Vault client with AppRole authentication.

### Sub-tasks
- [ ] T-006.1: Define VaultClient struct in plugin/vault_client.go
- [ ] T-006.2: Implement NewVaultClient constructor
- [ ] T-006.3: Implement Login() method for AppRole authentication
- [ ] T-006.4: Handle token refresh and errors

### Dependencies
- T-004 (Configuration)

### Implementation

```go
type VaultClient struct {
    address string
    mount   string
    roleID  string
    secretID string
    token   string
    client  *api.Client
}

func NewVaultClient(addr, mount, roleID, secretID string) (*VaultClient, error) {
    config := &api.Config{
        Address: addr,
    }
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    return &VaultClient{
        address: addr,
        mount:   mount,
        roleID:  roleID,
        secretID: secretID,
        client:  client,
    }, nil
}

func (c *VaultClient) Login() error {
    resp, err := c.client.Logical().Write("auth/approle/login", map[string]interface{}{
        "role_id":   c.roleID,
        "secret_id": c.secretID,
    })
    if err != nil {
        return err
    }
    
    if token, ok := resp.Data["token"]; ok {
        c.token = token.(string)
        c.client.SetToken(c.token)
    }
    return nil
}
```

### References
- HashiCorp Vault SDK: github.com/hashicorp/vault/api
- projects/openbao-plugin-cf/client/api_client.go

---

## T-007: Vault Client - List Secrets

**Priority**: HIGH | **Status**: pending

List secrets from Vault KVv2 mount.

### Sub-tasks
- [ ] T-007.1: Implement ListOrganizations() - list all orgs under KV mount
- [ ] T-007.2: Implement ListSecretsInOrganization(org) - list secrets in an org
- [ ] T-007.3: Implement ReadSecret(org, secret) - read a specific secret
- [ ] T-007.4: Handle pagination for large result sets

### Implementation

```go
// ListOrganizations lists all organizations under the KV mount
func (c *VaultClient) ListOrganizations() ([]string, error) {
    path := fmt.Sprintf("%s/metadata/", c.mount)
    resp, err := c.client.Logical().List(path)
    if err != nil {
        return nil, err
    }
    if resp == nil {
        return nil, nil
    }
    return resp.Data["keys"].([]string), nil
}

// ListSecretsInOrganization lists secrets in an org
func (c *VaultClient) ListSecretsInOrganization(org string) ([]string, error) {
    path := fmt.Sprintf("%s/metadata/%s", c.mount, org)
    resp, err := c.client.Logical().List(path)
    if err != nil {
        return nil, err
    }
    if resp == nil {
        return nil, nil
    }
    return resp.Data["keys"].([]string), nil
}

// ReadSecret reads a specific secret
func (c *VaultClient) ReadSecret(org, secret string) (map[string]interface{}, error) {
    path := fmt.Sprintf("%s/data/%s/%s", c.mount, org, secret)
    resp, err := c.client.Logical().Read(path)
    if err != nil {
        return nil, err
    }
    if resp == nil {
        return nil, nil
    }
    return resp.Data["data"].(map[string]interface{}), nil
}
```

---

## Acceptance Criteria

- NewVaultClient creates client
- Login performs AppRole authentication
- Token stored for subsequent requests
- ListOrganizations returns org list
- ListSecretsInOrganization returns secrets per org
- ReadSecret returns secret data