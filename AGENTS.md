# OpenBAO Vault Replicator - AI Agent Tasks

> **Quick Index** - See detailed tasks in separate files below

## Status Legend
| Status | Meaning |
|--------|---------|
| pending | Not started |
| in_progress | Currently working |
| completed | Done and verified |
| blocked | Waiting on dependency |

---

## Quick Status Summary

| Category | Total | Completed | Pending | In Progress |
|----------|-------|-----------|---------|-------------|
| Foundation | 3 | 0 | 3 | 0 |
| Config | 1 | 0 | 1 | 0 |
| Vault Client | 2 | 0 | 2 | 0 |
| Sync | 5 | 0 | 5 | 0 |
| Tests | 1 | 0 | 1 | 0 |
| **TOTAL** | **12** | **0** | **12** | **0** |

---

## Task Categories

| # | Category | File | Status |
|---|----------|------|--------|
| 01 | Foundation | [tasks/01-foundation.md](./tasks/01-foundation.md) | pending |
| 02 | Configuration | [tasks/02-configuration.md](./tasks/02-configuration.md) | pending |
| 03 | Vault Client | [tasks/03-vault-client.md](./tasks/03-vault-client.md) | pending |
| 04 | Sync Logic | [tasks/04-sync.md](./tasks/04-sync.md) | pending |
| 05 | Testing | [tasks/05-testing.md](./tasks/05-testing.md) | pending |

---

## Project Overview

- **Project Name**: openbao-vault-replicator-secret-plugin
- **Type**: Secret Engine Plugin for OpenBao
- **Core Functionality**: Replicate secrets from HashiCorp Vault (KVv2) to OpenBao (KVv2)
- **Module Path**: github.com/JavierLimon/openbao-vault-replicator-secret-plugin
- **Mount Path**: replicator/

---

## Architecture

HashiCorp Vault (Source)          OpenBao (Destination)
 kv2/ (shared mount)                secret/replicator/
   org-1/                           config
   org-2/                          sync/secrets
   ... (1500+ orgs)               sync/status
                                   sync/history
                                     
 AppRole: role_id/secret_id        secret/kv2/
                                        replicated secrets

---

## Auth Methods
- Source (Vault): AppRole - role_id + secret_id
- Destination (OpenBao): Token (stored in config)

---

## Quick Reference

### Core Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| /config | CRUD | Plugin configuration |
| /sync/secrets | POST | Trigger secret replication |
| /sync/status | GET | Show sync status |
| /sync/history | GET | List past operations |

### Build and Test
```
make build
make test
make test-cover
make lint
```

---

## Claim System

To claim a task:
1. Update status to in_progress in task file
2. Create branch for the task
3. Implement the solution
4. Commit, push, create PR
5. Update status to completed

---

## Production Quality Requirements

This should be in production. This should not have bugs or security issues. No shortcuts. If you find an issue or a bug, fix it, commit and push it, then continue with your work.

- All security vulnerabilities must be fixed before merging
- All bugs must be fixed before moving forward
- No TODO comments in production code
- Comprehensive error handling required

---

For detailed task lists, see individual task files above.