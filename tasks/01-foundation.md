# Foundation Tasks - openbao-vault-replicator-secret-plugin

Tasks required to set up the basic plugin structure.

## Status
- Total: 3
- Completed: 0
- Pending: 3

---

## T-001: Initialize Project Structure

**Priority**: HIGH | **Status**: pending

Create the Go module with proper dependencies.

### Sub-tasks
- [ ] T-001.1: Initialize go.mod with module name github.com/JavierLimon/openbao-vault-replicator-secret-plugin
- [ ] T-001.2: Add dependencies: github.com/openbao/openbao/sdk/v2, github.com/hashicorp/vault/api, github.com/stretchr/testify
- [ ] T-001.3: Create go.sum via go mod tidy
- [ ] T-001.4: Set up project directory structure (cmd/, plugin/, models/, docs/, .github/workflows/)

### Dependencies
- None

### References
- OpenBao Plugin Development: https://openbao.org/docs/plugins/plugin-development/
- Reference: projects/openbao-plugin-cf

---

## T-002: Create Plugin Entry Point

**Priority**: HIGH | **Status**: pending

Create main.go entry point in cmd/.

### Sub-tasks
- [ ] T-002.1: Create cmd/vault-replicator/main.go
- [ ] T-002.2: Implement plugin.ServeMultiplex with BackendFactoryFunc
- [ ] T-002.3: Add TLS provider configuration for plugin multiplexing

### Dependencies
- T-001 (Initialize Project)

### References
- File: projects/openbao-plugin-cf/cmd/openbao-plugin-cf/main.go

### Code Template
```go
package main

import (
    "fmt"
    "os"

    "github.com/openbao/openbao/api/v2"
    "github.com/openbao/openbao/sdk/v2/plugin"
    "github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin"
)

func GetVersion() string { return "0.1.0" }

func main() {
    if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
        fmt.Printf("OpenBAO Vault Replicator\n")
        fmt.Printf("Version: %s\n", GetVersion())
        os.Exit(0)
    }

    apiClientMeta := &api.PluginAPIClientMeta{}
    flags := apiClientMeta.FlagSet()
    flags.Parse(os.Args[1:])

    tlsConfig := apiClientMeta.GetTLSConfig()
    tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

    err := plugin.ServeMultiplex(&plugin.ServeOpts{
        BackendFactoryFunc: replicator.Factory,
        TLSProviderFunc:    tlsProviderFunc,
    })
    if err != nil {
        os.Exit(1)
    }
}
```

---

## T-003: Implement Backend

**Priority**: HIGH | **Status**: pending

Create Backend struct and Factory function.

### Sub-tasks
- [ ] T-003.1: Define Backend struct in plugin/backend.go
- [ ] T-003.2: Implement Factory function
- [ ] T-003.3: Register paths with OpenBao framework
- [ ] T-003.4: Add pathConfig, pathRoles, pathSync path handlers

### Dependencies
- T-002 (Plugin Entry Point)

### References
- File: projects/openbao-plugin-cf/plugin/backend.go

### Backend Struct Template
```go
type Backend struct {
    *framework.Backend
    storage logical.Storage
    
    // Thread safety
    mu sync.RWMutex
    logger hclog.Logger
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
    b := &Backend{
        storage: conf.StorageView,
        logger:  conf.Logger,
    }
    
    b.Backend = framework.NewBackend(&framework.BackendConfig{
        Logger:     conf.Logger,
        StorageView: conf.StorageView,
        System:     conf.System,
    })
    
    b.Paths = b.paths()
    
    return b, nil
}

func (b *Backend) paths() []*framework.Path {
    return []*framework.Path{
        b.pathConfig(),
        b.pathRoles(),
        b.pathSync(),
    }
}
```

---

## Implementation Notes

### OpenBao SDK Patterns

The Factory function signature must match:
```go
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error)
```

Path registration with ExistenceCheck:
```go
func (b *Backend) paths() []*framework.Path {
    return []*framework.Path{
        b.pathConfig(),
    }
}

func (b *Backend) HandleExistenceCheck(ctx context.Context, req *logical.Request, 
    data *framework.FieldData) (bool, error) {
    return false, nil
}
```

---

## Reference Implementations

| Plugin | Maturity | What to Copy |
|--------|----------|--------------|
| transform | 100% | All features, patterns, tests |
| kmip | 100% | Security patterns, audit |
| plugin-cf | 50% | Basic structure |