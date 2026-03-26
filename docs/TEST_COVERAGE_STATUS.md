# Test Coverage Status

Current test coverage for the Vault Replicator plugin.

## Coverage Summary

| Package | Coverage | Target | Status |
|---------|----------|--------|--------|
| plugin | ~44% | 80% | ⚠️ BELOW TARGET |

## Test Status

### Coverage by File

| File | Coverage | Notes |
|------|----------|-------|
| backend.go | ~70% | Core factory tested |
| path_config.go | ~77% | CRUD operations tested |
| path_sync.go | ~50% | Some operations need mocking |
| path_health.go | 100% | Fully tested |
| path_metrics.go | ~40% | Basic structure tested |
| vault_client.go | ~25% | Requires running Vault |
| openbao_client.go | ~30% | Requires OpenBao |
| audit.go | ~80% | Storage operations tested |
| version.go | 100% | Fully tested |

### Working Tests

- ✅ TestFactory (valid config)
- ✅ TestBackend_Help
- ✅ TestPathConfig_Read (exists/doesn't exist)
- ✅ TestPathSyncStatus_Read
- ✅ TestPathSyncHistory_List
- ✅ TestPathSyncHistory_Read
- ✅ TestPathAuditLogs_List
- ✅ TestVaultClient_* (expected failures without Vault)
- ✅ TestSyncStatus_JSON
- ✅ TestConfiguration_JSON
- ✅ TestVersion functions

### Skipped/Unable to Test

| Test | Reason |
|------|--------|
| VaultClient tests | Requires running Vault instance |
| OpenBaoClient tests | Requires running OpenBao instance |
| pathSyncSecrets | Complex replication logic needs integration test |

---

## Running Tests

### All Tests
```bash
go test ./...
go test -cover ./...
```

### Specific Package
```bash
go test -cover -run "TestPathConfig" ./plugin/
```

### With Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

---

## Improving Coverage

To reach 80% target:

1. **Create interfaces for clients** - Make VaultClient and OpenBaoClient testable via interfaces
2. **Use test storage** - Mock logical.Storage for path tests
3. **Add integration tests** - Run with actual Vault/OpenBao for E2E coverage

See [TEST_COVERAGE_LIMITATIONS.md](./TEST_COVERAGE_LIMITATIONS.md) for details.