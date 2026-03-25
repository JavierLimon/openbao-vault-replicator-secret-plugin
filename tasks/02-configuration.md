# Configuration Tasks - openbao-vault-replicator-secret-plugin

Configuration path CRUD for Vault connection details.

## Status
- Total: 1
- Completed: 0
- Pending: 1

---

## T-004: Configuration Path

**Priority**: HIGH | **Status**: pending

Implement configuration CRUD path for storing Vault connection details.

### Sub-tasks
- [ ] T-004.1: Define Configuration model in models/configuration.go
- [ ] T-004.2: Implement pathConfig with read/write/delete operations
- [ ] T-004.3: Add config validation and defaults
- [ ] T-004.4: Mask sensitive fields (token, secret_id) on read

### Dependencies
- T-003 (Backend)

### Config Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| vault_address | string | yes | Vault server URL (e.g., https://vault.example.com:8200) |
| vault_mount | string | yes | Vault KVv2 mount path (default: kv2) |
| approle_role_id | string | yes | AppRole role_id |
| approle_secret_id | string | yes | AppRole secret_id |
| destination_token | string | yes | OpenBao token to write secrets |
| destination_mount | string | yes | OpenBao KVv2 mount (default: kv2) |
| organization_path | string | yes | Path in Vault where orgs live (e.g., data/) |

### References
- File: projects/openbao-plugin-cf/plugin/path_config.go
- File: projects/openbao-plugin-cf/models/configuration.go

### Implementation Pattern

Configuration model:
```go
type Configuration struct {
    VaultAddress      string `json:"vault_address"`
    VaultMount        string `json:"vault_mount"`
    AppRoleRoleID     string `json:"approle_role_id"`
    AppRoleSecretID   string `json:"approle_secret_id"`
    DestinationToken  string `json:"destination_token"`
    DestinationMount string `json:"destination_mount"`
    OrganizationPath string `json:"organization_path"`
}
```

Path handler pattern:
```go
func (b *Backend) pathConfig() *framework.Path {
    return &framework.Path{
        Pattern: "config",
        Operations: map[logical.Operation]*framework.OperationHandler{
            logical.ReadOperation:   &framework.OperationHandler{...},
            logical.CreateOperation: &framework.OperationHandler{...},
            logical.DeleteOperation: &framework.OperationHandler{...},
        },
        Fields: map[string]*framework.FieldSchema{...},
    }
}
```

---

## Acceptance Criteria

- POST /config creates/updates config
- GET /config reads config (token masked)
- DELETE /config removes config
- Config stored in OpenBao storage