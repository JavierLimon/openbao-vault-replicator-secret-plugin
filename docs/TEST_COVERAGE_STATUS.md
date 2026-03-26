# Test Coverage Status

Current test coverage for the Vault Replicator plugin.

## Coverage Summary

| Package | Coverage | Target | Status |
|---------|----------|--------|--------|
| plugin | 0% | 80% | ❌ NO TESTS |

## Status

**⚠️ Tests were removed** - The repository previously had unit tests but they were removed due to build failures. The test files were deleted in commit `b6cc33e`.

### What Was Tested Before

| File | Coverage (when tests existed) |
|------|-------------------------------|
| backend.go | ~70% |
| path_config.go | ~77% |
| path_sync.go | ~50% |
| path_health.go | 100% |
| path_metrics.go | ~40% |
| vault_client.go | ~25% |
| openbao_client.go | ~30% |
| audit.go | ~80% |
| version.go | 100% |

---

## Running Tests

```bash
go test ./...
go test -cover ./...
```

---

## Next Steps

To restore and improve test coverage:

1. **Restore deleted tests** - Restore `plugin/backend_test.go` and fix the issues
2. **Create interfaces** - Make VaultClient and OpenBaoClient testable via interfaces
3. **Add integration tests** - Use Docker to test with actual Vault/OpenBao instances

See [TEST_COVERAGE_LIMITATIONS.md](./TEST_COVERAGE_LIMITATIONS.md) for details on what's needed.