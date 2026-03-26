# Test Coverage Limitations

This document describes the limitations in achieving high test coverage for the Vault Replicator plugin.

## Current Status

- **Current Coverage**: 15.9%
- **Target Coverage**: 80%
- **Gap**: 64.1%

## What's Tested

Current tests in `plugin/backend_test.go` and `plugin/vault_client_test.go`:
- ✅ Backend Factory
- ✅ Configuration CRUD (partial)
- ✅ Sync status
- ✅ Version functions
- ✅ Configuration helpers (ShouldSyncOrg, ShouldAllowDeletionSync)

## Known Limitations

### 1. External API Dependencies

The plugin integrates with two external systems that are difficult to mock:

| Component | Issue | Impact |
|-----------|-------|--------|
| Vault Client | Uses `hashicorp/vault/api` - requires running Vault | ~20% coverage gap |
| OpenBao Client | Uses OpenBao SDK - requires running OpenBao | ~10% coverage gap |

### 2. Sync Logic (pathSyncSecrets)

The core sync function:
- Calls Vault API to list orgs
- Calls Vault API to list secrets per org  
- Calls OpenBao API to write each secret
- Requires mocking both Vault and OpenBao clients

---

## Recommended Path to 80%

See [TEST_COVERAGE_PLAN.md](./TEST_COVERAGE_PLAN.md) for prioritized list from easy to hard.

### Quick Wins (Easy)

1. **Health endpoint** - 3% coverage, no dependencies
2. **Metrics endpoint** - 5% coverage  
3. **Version functions** - 4% coverage
4. **Config helpers** - 3% coverage

### Medium Effort

5. **Audit logger** - 10% coverage (needs storage mock)
6. **Retry logic** - 15% coverage (some needs Vault mock)
7. **Secret data** - 5% coverage

### Hard (Needs Interfaces)

8. **Sync logic** - 20% coverage
9. **Backend factory** - 5% coverage  
10. **OpenBao client** - 8% coverage

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