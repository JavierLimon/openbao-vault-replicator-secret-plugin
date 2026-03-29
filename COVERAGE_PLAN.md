# Unit Test Coverage Plan

## Current Coverage: ~55% (after recent improvements)

## Functions by Difficulty (Easy → Hard)

### EASY - No External Dependencies (Can Test Now)

| Function | File | Current Coverage | Status |
|----------|------|------------------|--------|
| validateOrgName | retry.go | 100% | ✅ Tested |
| writeAuditEvent | audit.go | ~67% | Can improve with storage error test |
| pathConfigRead | path_config.go | ~50% | Can improve with read error test |
| writeToLocalKVWithMetadata | openbao_client.go | ~21% | Can improve with config error test |
| pathSyncSecrets | path_sync.go | ~7% | Can improve with validation tests |
| LoginToVault | retry.go | ~21% | Can improve with auth failure test |

### MEDIUM - Requires Simple Mocking

| Function | File | Current Coverage | Status |
|----------|------|------------------|--------|
| getOpenBaoClient | openbao_client.go | ~71% | Already tested |
| Factory | backend.go | ~80% | Mostly covered |
| saveSyncHistory | path_sync.go | ~71% | Can improve with storage error |
| readSyncStatus | path_sync.go | ~78% | Can improve with not found case |
| saveSyncStatus | path_sync.go | ~80% | Can improve with storage error |
| listSecretsInDestination | openbao_client.go | 0% | Can test config errors (NEW) |
| deleteSecretFromDestination | openbao_client.go | 0% | Can test config errors (NEW) |

### HARD - Require Vault/OpenBao API Mocking (Integration Tests)

| Function | File | Current Coverage | Status |
|----------|------|------------------|--------|
| ListOrganizationsWithRetry | retry.go | 0.0% | Needs Vault mock |
| ListSecretsInOrgWithRetry | retry.go | 0.0% | Needs Vault mock |
| ReadSecretWithRetry | retry.go | 0.0% | Needs Vault mock |
| listOrganizationsInternal | retry.go | 0.0% | Needs Vault mock |
| listSecretsInOrgInternal | retry.go | 0.0% | Needs Vault mock |
| listSecretsRecursive | retry.go | 0.0% | Needs Vault mock |
| readSecretInternal | retry.go | 0.0% | Needs Vault mock |
| ReadSecretWithMetadata | retry.go | 0.0% | Needs Vault mock |
| readSecretWithMetadataInternal | retry.go | 0.0% | Needs Vault mock |
| ListOrganizationsPaged | retry.go | 0.0% | Needs Vault mock |
| ListSecretsInOrgPaged | retry.go | 0.0% | Needs Vault mock |
| listSecretsInDestination | openbao_client.go | 0.0% | Needs OpenBao mock |
| deleteSecretFromDestination | openbao_client.go | 0.0% | Needs OpenBao mock |

---

## IMPROVED - Already Completed

These functions were improved in recent sessions:

| Function | File | Before | After |
|----------|------|--------|-------|
| pathRolesRead | path_roles.go | 0.0% | 100% |
| pathRolesList | path_roles.go | 0.0% | 100% |
| readConfig | path_config.go | 28.6% | ~93% |
| pathConfigWrite | path_config.go | 73.0% | ~76% |
| pathConfigDelete | path_config.go | 66.7% | 100% |
| pathSyncHistoryList | path_sync.go | 66.7% | ~78% |
| pathSyncHistoryRead | path_sync.go | 66.7% | ~71% |
| pathSyncHistoryTimestampRead | path_sync.go | 68.8% | ~75% |
| validateOrgName | retry.go | N/A | 100% (NEW) |
| listSecretsInDestination | openbao_client.go | N/A | 0% (NEW) |
| deleteSecretFromDestination | openbao_client.go | N/A | 0% (NEW) |

---

## REMOVED - Encryption Layer Removed

The encryption layer was removed - OpenBao storage handles encryption at rest:

| Function | File | Reason |
|----------|------|--------|
| encryptConfig | encryption.go | Removed |
| decryptConfig | encryption.go | Removed |
| writeEncryptedConfig | encryption.go | Removed |
| Encrypter | encryption.go | Removed |
| SecureConfig | encryption.go | Removed |
| getOrCreateKey | encryption.go | Removed |
| Encrypt | encryption.go | Removed |
| Decrypt | encryption.go | Removed |

---

## Recommended Order (Easy → Hard)

### Phase 1: Easy Wins (No Mocking Needed)
1. ✅ validateOrgName - 100% covered
2. writeAuditEvent - Add storage error test
3. pathConfigRead - Add read error test
4. pathSyncSecrets - Add validation tests
5. LoginToVault - Add auth failure test cases
6. writeToLocalKVWithMetadata - Add config error test

### Phase 2: Medium Effort
7. listSecretsInDestination - Test config errors (NEW)
8. deleteSecretFromDestination - Test config errors (NEW)
9. saveSyncHistory - Test storage errors
10. readSyncStatus - Test not found case
11. saveSyncStatus - Test storage errors

### Phase 3: Hard (Integration Tests Required)
12. All Vault API functions - Need integration tests with testcontainers
13. All OpenBao client functions - Need integration tests
14. Full sync path testing - Needs Vault + OpenBao mocks

---

## Summary

- **Easy remaining**: 5 functions can be improved
- **Medium remaining**: 5 functions can be improved
- **Hard (Integration)**: 13 functions require integration tests
- **New functions added**: 3 (validateOrgName, listSecretsInDestination, deleteSecretFromDestination)
- **Functions removed**: 8 (encryption layer)

## Recommendation

Continue with Phase 1 Easy tests. For Phase 3, use testcontainers or HTTP mocking libraries like go-vcr.
