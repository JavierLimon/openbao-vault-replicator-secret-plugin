# Test Coverage Status

Current test coverage for the Vault Replicator plugin.

## Coverage Summary

| Package | Coverage | Target | Status |
|---------|----------|--------|--------|
| plugin | ~55% | 80% | ⚠️ IN PROGRESS |

## Test Files

| File | Coverage | Notes |
|------|----------|-------|
| backend_test.go | ~80% | Factory, config CRUD, sync status |
| extended_test.go | ~60% | Extended tests, validation, helpers |
| vault_client_test.go | ~25% | Vault client tests |
| openbao_client.go | ~25% | OpenBao client tests |

## Files With Tests

| File | Functions Tested |
|------|-----------------|
| backend.go | Factory |
| path_config.go | pathConfigRead, pathConfigWrite, pathConfigDelete, readConfig |
| path_sync.go | pathSyncSecrets, pathSyncStatusRead, pathSyncHistory*, saveSyncStatus, saveSyncHistory |
| path_health.go | pathHealthRead |
| path_metrics.go | pathMetricsRead |
| path_roles.go | pathRolesRead, pathRolesList |
| openbao_client.go | getOpenBaoClient, writeToLocalKVWithMetadata |
| audit.go | AuditLogger methods |
| retry.go | LoginToVault, validateOrgName, RetryWithBackoff |
| version.go | All version functions |

## New Functions Added

| Function | File | Coverage |
|----------|------|----------|
| validateOrgName | retry.go | ~100% |
| listSecretsInDestination | openbao_client.go | 0% (needs tests) |
| deleteSecretFromDestination | openbao_client.go | 0% (needs tests) |

## Functions Removed (Encryption Layer)

| Function | File | Reason |
|----------|------|--------|
| encryptConfig | encryption.go | Removed - OpenBao handles encryption |
| decryptConfig | encryption.go | Removed - OpenBao handles encryption |
| writeEncryptedConfig | encryption.go | Removed - OpenBao handles encryption |
| Encrypter struct | encryption.go | Removed |
| SecureConfig struct | encryption.go | Removed |

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Next Steps to Reach 80%

1. **Add tests for new functions**:
   - listSecretsInDestination
   - deleteSecretFromDestination

2. **Improve coverage on existing functions**:
   - pathSyncSecrets - validation errors
   - LoginToVault - auth failure cases
   - writeToLocalKVWithMetadata - error paths

3. **Add integration tests** for Vault/OpenBao client functions (requires mocking)

See [TEST_COVERAGE_LIMITATIONS.md](./TEST_COVERAGE_LIMITATIONS.md) for full list.
