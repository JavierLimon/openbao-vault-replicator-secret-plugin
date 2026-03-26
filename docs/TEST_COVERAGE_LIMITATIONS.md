# Test Coverage Limitations

This document describes the limitations in achieving high test coverage for the Vault Replicator plugin.

## Current Status

- **Current Coverage**: ~44%
- **Target Coverage**: 80%
- **Gap**: 36%

## Known Limitations

### 1. External API Dependencies

The plugin integrates with two external systems that are difficult to mock:

| Component | Issue | Impact |
|-----------|-------|--------|
| Vault Client | Uses `hashicorp/vault/api` - requires running Vault | ~20% coverage gap |
| OpenBao Client | Uses OpenBao SDK - requires running OpenBao | ~10% coverage gap |

**Why this matters:**
- `vault_client.go` makes real HTTP calls to Vault
- `openbao_client.go` uses OpenBao SDK client
- Cannot mock at the HTTP layer without interfaces

**Solution:**
```go
// Create interfaces for testability
type VaultClienter interface {
    Login() error
    ListOrganizations() ([]string, error)
    ListSecretsInOrganization(org string) ([]string, error)
    ReadSecret(org, secret string) (map[string]interface{}, error)
}

type OpenBaoWriter interface {
    Write(path string, data map[string]interface{}) error
    Read(path string) (map[string]interface{}, error)
}
```

### 2. Framework FieldData Initialization

The OpenBao SDK's `framework.FieldData` requires proper initialization:

```go
// This works in tests
data := &framework.FieldData{
    Raw: map[string]interface{}{
        "field_name": "value",
    },
    Schema: fieldSchema,
}

// But many edge cases are hard to trigger
data.Get("nonexistent") // Returns nil, not error
```

### 3. Storage Operations

Some storage paths require specific state:
- Sync status storage needs prior sync
- Audit log storage needs audit events
- History storage needs completed syncs

### 4. Replication Logic

The core `pathSyncSecrets` function:
- Calls Vault API to list orgs
- Calls Vault API to list secrets per org
- Calls OpenBao API to write each secret
- Complex error handling paths

This requires integration testing with both Vault and OpenBao running.

---

## Unreachable Code Paths

### Functions with Low Coverage

| Function | File | Coverage | Reason |
|----------|------|----------|--------|
| pathSyncSecrets | path_sync.go | 0% | Requires full replication |
| createVaultClient | path_sync.go | 83% | Partial, needs auth |
| loginToVault | path_sync.go | 40% | Needs running Vault |
| listOrganizations | path_sync.go | 35% | Needs running Vault |
| listSecretsInOrg | path_sync.go | 30% | Needs running Vault |
| readSecret | path_sync.go | 37% | Needs running Vault |

### Storage Error Paths

These are hard to trigger without mocking storage:
- Storage read failures
- Storage write failures
- JSON encoding/decoding errors

---

## Recommendations for 80% Coverage

### 1. Add Interfaces (High Impact)

Create interfaces for external clients, enabling easy mocking:

```go
// plugin/interfaces.go
type VaultClientInterface interface {
    Login() error
    ListOrganizations() ([]string, error)
    ListSecretsInOrganization(org string) ([]string, error)
    ReadSecret(org, secret string) (map[string]interface{}, error)
}

type OpenBaoClientInterface interface {
    WriteSecret(mount, path string, data map[string]interface{}) error
    ReadSecret(mount, path string) (map[string]interface{}, error)
}
```

### 2. Use hashicorp/vault/sdk helper

The Vault SDK provides `teststorage` for mocking:

```go
import "github.com/hashicorp/vault/sdk/helper/teststorage"
```

### 3. Integration Tests

Create a separate test binary that runs against real instances:

```bash
# Run integration tests
RUN_INTEGRATION=true go test -tags=integration ./plugin/
```

### 4. Contract Testing

Test the API contracts without full implementation:

```go
// Test that paths are registered correctly
func TestPathRegistration(t *testing.T) {
    b := &Backend{}
    paths := b.paths()
    
    expectedPaths := []string{"config", "sync/secrets", "health", "metrics"}
    for _, p := range expectedPaths {
        // Verify path exists
    }
}
```

---

## Files That Would Help

1. `plugin/interfaces.go` - Define interfaces for mocking
2. `plugin/mocks/` - Mock implementations
3. `plugin/replicator_integration_test.go` - Integration tests
4. `test/` - Docker-compose for local testing

---

## Conclusion

The 44% coverage is primarily due to external dependencies on Vault and OpenBao APIs. With interfaces and proper mocking infrastructure, coverage could potentially reach 80% for the business logic, but integration tests would still be needed for full E2E coverage.