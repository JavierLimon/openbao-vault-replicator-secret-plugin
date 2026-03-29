# Test Limitations

This document describes functions that cannot be fully unit tested due to external dependencies or architectural constraints.

## Current Coverage: ~55% (after recent improvements)

## Cannot Unit Test (Require Integration Tests)

These functions require a running Vault/OpenBao instance or cannot be mocked in unit tests:

### Vault Client / Retry Functions (retry.go)
- `ListOrganizationsWithRetry` - Requires Vault API client
- `ListSecretsInOrgWithRetry` - Requires Vault API client
- `ReadSecretWithRetry` - Requires Vault API client
- `listOrganizationsInternal` - Requires Vault API client
- `listSecretsInOrgInternal` - Requires Vault API client
- `listSecretsRecursive` - Requires Vault API client
- `readSecretInternal` - Requires Vault API client
- `ReadSecretWithMetadata` - Requires Vault API client
- `readSecretWithMetadataInternal` - Requires Vault API client
- `ListOrganizationsPaged` - Requires Vault API client
- `ListSecretsInOrgPaged` - Requires Vault API client

### OpenBao Client (openbao_client.go)
- `getOpenBaoClient` - Partially tested
- `writeToLocalKVWithMetadata` - Partially tested
- `listSecretsInDestination` - Requires OpenBao API client (NEW)
- `deleteSecretFromDestination` - Requires OpenBao API client (NEW)

## Already Tested

These functions have unit tests:

| Function | Status |
|----------|--------|
| pathRolesRead | ✅ 100% - Tested ErrUnsupportedOperation |
| pathRolesList | ✅ 100% - Tested ErrUnsupportedOperation |
| pathConfigRead | ✅ Tested with storage errors |
| pathConfigWrite | ✅ Tested with validation failures |
| pathConfigDelete | ✅ Tested with storage error |
| pathSyncHistoryList | ✅ Tested empty and with entries |
| pathSyncHistoryRead | ✅ Tested invalid timestamp |
| pathSyncHistoryTimestampRead | ✅ Tested invalid timestamp |
| saveSyncStatus | ✅ Covered |
| saveSyncHistory | ✅ Covered |
| validateOrgName | ✅ Tested with various inputs (NEW) |

## Can Improve (Easy Remaining)

| Function | Current Coverage | Can Test |
|----------|------------------|----------|
| writeAuditEvent | ~67% | Storage Put error |
| pathConfigRead | ~50% | Storage Get error |
| LoginToVault | ~21% | Auth failure cases |
| writeToLocalKVWithMetadata | ~21% | Config error path |
| listSecretsInDestination | 0% | Config error, list error (NEW) |
| deleteSecretFromDestination | 0% | Config error, delete error (NEW) |

## Recent Changes (Encryption Removed)

The encryption layer was removed - OpenBao storage handles encryption at rest:

| Removed Function | Reason |
|-----------------|--------|
| encryptConfig | Removed - config stored directly |
| decryptConfig | Removed - config stored directly |
| writeEncryptedConfig | Removed - replaced with direct storage |
| Encrypter struct | Removed |
| SecureConfig struct | Removed |

## Recommendation

For full coverage, add integration tests with:
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) for Vault/OpenBao
- HTTP recording libraries like [go-vcr](https://github.com/dnaeon/go-vcr)
- Mock interface implementations for KV clients
