# Testing Tasks - openbao-vault-replicator-secret-plugin

Unit tests with 80%+ coverage target.

## Status
- Total: 1
- Completed: 0
- Pending: 1

---

## T-013: Unit Tests

**Priority**: HIGH | **Status**: pending

Implement unit tests with 80%+ coverage.

### Sub-tasks
- [ ] T-013.1: Backend unit tests
- [ ] T-013.2: Vault client tests (with mock)
- [ ] T-013.3: Config path tests
- [ ] T-013.4: Sync logic tests

### Test Coverage Targets

| Package | Target |
|---------|--------|
| plugin | 80% |
| models | 90% |

### Test Implementation Workflow

1. Bring a list of unit tests from easy to hard to implement
2. Work implementing Easy and Medium tests first
3. Bring a list of unit tests not covered 100%
4. Final push to achieve 80-98% coverage

### Test Categories

- Backend tests: Factory, paths, existence check
- Config tests: CRUD operations, validation
- Vault client tests: Login, list, read (mock Vault)
- Sync tests: Trigger sync, status, history

### Test Pattern

```go
package replicator

import (
    "testing"
)

func TestBackend_Factory(t *testing.T) {
    // Test Factory creates backend correctly
}

func TestPathConfig_Read(t *testing.T) {
    // Test config read
}

func TestVaultClient_Login(t *testing.T) {
    // Test AppRole login with mock
}
```

---

## Test Execution

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test -v ./plugin/

# Run with race detection
go test -race ./...
```

---

## Acceptance Criteria

- Unit tests achieve 80%+ coverage on plugin package
- All CRUD operations have tests
- Vault client methods have tests (mocked)
- Sync logic has tests