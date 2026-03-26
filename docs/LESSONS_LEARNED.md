# OpenBao Plugin Development - Lessons Learned

## Overview

This document captures what we learned building the `openbao-vault-replicator-secret-plugin` and what we overlooked or didn't know at the beginning.

---

## What We Learned: Building OpenBao Plugins

### 1. Core Plugin Architecture

**Key Pattern - The Factory Function:**
```go
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
    b := &Backend{
        storage: conf.StorageView,
        logger:  conf.Logger,
    }
    
    b.Backend = &framework.Backend{
        Help:        "Plugin description",
        PathsSpecial: &logical.Paths{
            Unauthenticated: []string{"health"},  // Public endpoints
        },
        Paths: []*framework.Path{
            b.pathConfig(),
            b.pathSync(),
            // ... more paths
        },
        BackendType: logical.TypeLogical,
    }
    
    if err := b.Setup(ctx, conf); err != nil {
        return nil, err
    }
    return b, nil
}
```

**Key Points:**
- Embed `*framework.Backend` in your Backend struct
- Implement `Factory()` function - this is what OpenBao calls
- Use `b.Setup(ctx, conf)` to initialize
- Register paths in the `Paths` array

### 2. Path Patterns (API Endpoints)

**Standard CRUD Pattern:**
```go
func (b *Backend) pathConfig() *framework.Path {
    return &framework.Path{
        Pattern: "config",
        Operations: map[logical.Operation]framework.OperationHandler{
            logical.ReadOperation: &framework.PathOperation{
                Summary:     "Read config",
                Callback:    b.pathConfigRead,
            },
            logical.CreateOperation: &framework.PathOperation{
                Summary:     "Create config",
                Callback:    b.pathConfigWrite,
            },
            // Update, Delete...
        },
        Fields: map[string]*framework.FieldSchema{
            "vault_address": {
                Type:        framework.TypeString,
                Description: "Vault server URL",
            },
            // More fields...
        },
    }
}
```

**Key Points:**
- `Pattern` is the URL path (e.g., "config", "sync/secrets")
- `Operations` maps HTTP methods to handlers
- `Fields` defines input validation
- Use `data.Get("field_name")` to retrieve values

### 3. Storage Pattern

**Storing Configuration:**
```go
func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
    config := &Configuration{
        VaultAddress: data.Get("vault_address").(string),
        // ... more fields
    }
    
    entry, err := logical.StorageEntryJSON("config", config)
    if err != nil {
        return nil, err
    }
    
    if err := req.Storage.Put(ctx, entry); err != nil {
        return nil, err
    }
    return nil, nil
}
```

**Reading Configuration:**
```go
func (b *Backend) readConfig(ctx context.Context, storage logical.Storage) (*Configuration, error) {
    entry, err := storage.Get(ctx, "config")
    if err != nil {
        return nil, err
    }
    if entry == nil {
        return nil, nil
    }
    
    var config Configuration
    if err := entry.DecodeJSON(&config); err != nil {
        return nil, err
    }
    return &config, nil
}
```

**Key Points:**
- Use `logical.StorageEntryJSON(key, object)` to serialize
- Use `storage.Put(ctx, entry)` to save
- Use `storage.Get(ctx, key)` to retrieve
- Use `entry.DecodeJSON(&object)` to deserialize

### 4. Health and Metrics Endpoints

**Health Endpoint (must be unauthenticated):**
```go
func (b *Backend) pathHealth() *framework.Path {
    return &framework.Path{
        Pattern: "health",
        Operations: map[logical.Operation]framework.OperationHandler{
            logical.ReadOperation: &framework.PathOperation{
                Summary:     "Health check",
                Callback:    b.pathHealthRead,
                Unauthenticated: true,
            },
        },
    }
}
```

**Key Points:**
- Add to `PathsSpecial.Unauthenticated` for public access
- Track metrics with atomic operations for thread safety

---

## What We Overlooked / Missed

### 1. Not Checking Remote Branches Early

**What happened:** We didn't check GitHub for new AI-created branches until much later. The orchestrator was creating branches (`opencode-20260325_*`) but we weren't monitoring or merging them.

**What we should have done:**
```bash
# Check branches at start of every session
git ls-remote --heads git@github.com:JavierLimon/openbao-vault-replicator-secret-plugin

# Or check locally after clone
git fetch --all && git branch -r
```

**Lesson:** The orchestrator creates branches automatically. We need to poll for them and merge.

### 2. SSH Authentication Confusion

**What happened:** Initially tried SSH to remote without password, failed. The correct way is `sshpass` with password.

**What we learned:**
```bash
# Correct command (from ORCHESTRATION_GUIDE.md)
sshpass -p '7Jse3i1hpk' ssh -o StrictHostKeyChecking=no javier@10.0.0.37 "command"

# Check running tasks
sshpass -p '7Jse3i1hpk' ssh javier@10.0.0.37 "ps aux | grep opencode | grep -v grep"
```

### 3. Test Coverage Limitations

**What happened:** The AI agent wrote tests, but coverage only reached ~44% instead of 80% target.

**Why:**
- Vault client requires a running Vault instance or complex mocking
- The OpenBao SDK's `framework.FieldData` requires proper initialization
- Network calls can't be easily mocked without interfaces

**What we learned:**
- For KV client testing, need to mock `logical.Storage`
- For external API testing (Vault), use interfaces for testability
- Consider using `testbackend` from hashicorp/vault for testing

### 4. Documentation Needed After Code

**What happened:** We created docs AFTER code was done, but some were created manually (WORKFLOW.md, TROUBLESHOOTING.md, EXAMPLES.md) while others came from AI agents.

**What we learned:**
- Have AI create docs as part of task completion
- Reference existing plugins (transform) for doc structure
- Docs should include: API reference, workflow diagrams, troubleshooting, examples

### 5. Prompt Format Matters

**What happened:** Early prompts were vague ("Add feature X"). Later we learned to use structured prompts with MUST DO / MUST NOT DO sections.

**What works:**
```
## GOAL
Add the /metrics endpoint to the plugin

## CONTEXT
- Working in: /home/javier/vault-replicator-plugin/plugin
- Reference: projects/openbao-transform-secret-plugin/plugin/path_metrics.go

## MUST DO
1. Add GET /metrics path to Backend
2. Response must include: total_requests, sync_total, etc.
3. Update backend.go to register the path

## MUST NOT DO
- Don't duplicate health endpoint functionality

## VERIFY
go build ./... must pass

After: create branch metrics, commit, push, report branch.
```

### 6. Merge Conflict Strategy

**What happened:** When merging branches from AI, there were conflicts in:
- Makefile
- dist/replicator (binary)
- go.mod

**Resolution:** Use `git checkout --ours` to keep our version, then commit.

```bash
git merge origin/opencode-branch
# If conflict:
git checkout --ours .
git add -A
git commit -m "Merge (keep ours)"
```

---

## Orchestrator Workflow Summary

### Starting a New Task
```bash
# Launch on remote
sshpass -p '7Jse3i1hpk' ssh -o StrictHostKeyChecking=no javier@10.0.0.37 \
  "cd ~ && nohup bash opencode_loop.sh --repo git@github.com:JavierLimon/openbao-vault-replicator-secret-plugin -- 'YOUR TASK' 1 > /tmp/task.log 2>&1 &"
```

### Checking Progress
```bash
# Check running tasks
sshpass -p '7Jse3i1hpk' ssh javier@10.0.0.37 "ps aux | grep opencode | grep -v grep | wc -l"

# Check logs
sshpass -p '7Jse3i1hpk' ssh javier@10.0.0.37 "tail -50 /tmp/task.log"
```

### Merging Branches
```bash
# Locally
git fetch origin
git checkout main
git merge origin/opencode-branch
git push origin main

# Delete old branches
git push origin --delete opencode-branch
```

---

## Files Created

### Plugin Core
- `plugin/backend.go` - Backend factory
- `plugin/path_config.go` - Config CRUD
- `plugin/path_sync.go` - Sync endpoints + replication logic
- `plugin/path_roles.go` - Role management
- `plugin/path_health.go` - Health endpoint
- `plugin/path_metrics.go` - Metrics endpoint
- `plugin/vault_client.go` - Vault AppRole client
- `plugin/openbao_client.go` - OpenBao KV client
- `plugin/audit.go` - Audit logging
- `plugin/version.go` - Version info

### Build & Config
- `Makefile` - Build targets
- `.golangci.yml` - Linter config
- `cmd/vault-replicator/main.go` - Entry point
- `go.mod`, `go.sum` - Dependencies

### Documentation
- `docs/API.md` - API reference
- `docs/WORKFLOW.md` - Workflow diagrams
- `docs/TROUBLESHOOTING.md` - Troubleshooting
- `docs/EXAMPLES.md` - Usage examples

### Testing
- `plugin/replicator_test.go` - Unit tests (~44% coverage)

---

## Key Takeaways

1. **Always poll for branches** - The orchestrator creates them automatically
2. **Use structured prompts** - Include GOAL, CONTEXT, MUST DO, MUST NOT DO, VERIFY
3. **Document as you go** - Don't leave docs for the end
4. **Test early** - Aim for tests from the start, not after
5. **Merge with --ours** - AI branches often conflict on binaries/config
6. **SSH uses sshpass** - Not direct SSH key auth
7. **Check orchestrator status** - Look at master-status.json for progress

---

## Future Improvements

1. **Auto-merge script** - Automatically fetch and merge new opencode branches
2. **Better test mocking** - Create interfaces for external clients to enable easier testing
3. **Documentation templates** - Create templates for AI to fill in
4. **Branch monitoring** - Set up alerts for new branches
5. **Coverage tracking** - Integrate coverage reporting into CI