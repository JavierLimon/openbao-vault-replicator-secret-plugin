# Test Coverage Limitations

This document describes the limitations in achieving high test coverage for the Vault Replicator plugin.

## Current Status

- **Current Coverage**: ~55%
- **Target Coverage**: 80%
- **Gap**: 25%

## What's Tested

Tests exist in:
- `plugin/backend_test.go` - Core backend tests
- `plugin/extended_test.go` - Extended integration-style tests
- `plugin/vault_client_test.go` - Vault client tests

### Coverage by File

| File | Coverage | Notes |
|------|----------|-------|
| backend.go | ~80% | Factory |
| path_config.go | ~85% | CRUD operations |
| path_sync.go | ~35% | Sync logic (partial) |
| path_health.go | 100% | Health endpoint |
| path_metrics.go | ~40% | Metrics |
| path_roles.go | 100% | Unsupported operations |
| openbao_client.go | ~25% | Client functions |
| audit.go | ~67% | Audit logging |
| retry.go | ~30% | Retry logic (partial) |
| version.go | 100% | Version functions |
| encryption.go | N/A | REMOVED |

## Known Limitations

### 1. External API Dependencies

The plugin integrates with two external systems that are difficult to mock:

| Component | Issue | Impact |
|-----------|-------|--------|
| Vault Client | Uses `hashicorp/vault/api` - requires running Vault | ~15% coverage gap |
| OpenBao Client | Uses OpenBao SDK - requires running OpenBao | ~10% coverage gap |

### 2. Sync Logic (pathSyncSecrets)

The core sync function:
- Calls Vault API to list orgs
- Calls Vault API to list secrets per org
- Calls OpenBao API to write each secret
- Calls OpenBao API to list/delete orphaned secrets
- Requires mocking both Vault and OpenBao clients

### 3. New Functions Added

| Function | File | Issue |
|----------|------|-------|
| listSecretsInDestination | openbao_client.go | Requires OpenBao mock |
| deleteSecretFromDestination | openbao_client.go | Requires OpenBao mock |
| validateOrgName | retry.go | ✅ Fully testable |

### 4. Removed Functions (Encryption)

The encryption layer was removed - OpenBao storage handles encryption at rest:

| Removed Function | File |
|-----------------|------|
| encryptConfig | encryption.go |
| decryptConfig | encryption.go |
| writeEncryptedConfig | encryption.go |
| Encrypter | encryption.go |
| SecureConfig | encryption.go |

---

## Recommended Path to 80%

### Quick Wins (Easy - No Dependencies)

1. **validateOrgName** - 100% ✅ Done
2. **pathSyncSecrets validation** - ~10% → ~25%
3. **LoginToVault error paths** - ~21% → ~40%
4. **writeToLocalKVWithMetadata errors** - ~21% → ~40%
5. **listSecretsInDestination** - 0% → ~30%
6. **deleteSecretFromDestination** - 0% → ~30%

### Medium Effort

7. **pathSyncSecrets full flow** - needs mocks for Vault/OpenBao clients
8. **saveSyncHistory storage errors**
9. **readSyncStatus not found case**

### Hard (Needs Interfaces or Integration Tests)

10. **Vault API functions** - All require Vault mock
11. **OpenBao client functions** - Require OpenBao mock

---

## Prerequisites for Hard Tests

Before implementing Hard tests, create interfaces:

```go
// plugin/interfaces.go
type VaultClientInterface {...}
type OpenBaoClientInterface {...}
```

Or use integration tests with Docker:
```bash
RUN_INTEGRATION=true go test -tags=integration ./...
```

---

## Files That Would Help

1. `plugin/interfaces.go` - Define interfaces for mocking
2. `plugin/mocks/` - Mock implementations
3. `plugin/replicator_integration_test.go` - Integration tests
4. `test/` - Docker-compose for local testing
