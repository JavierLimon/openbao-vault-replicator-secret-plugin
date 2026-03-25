# Sync Logic Tasks - openbao-vault-replicator-secret-plugin

Sync endpoints, local KV write, and progress tracking.

## Status
- Total: 5
- Completed: 0
- Pending: 5

---

## T-008: Sync Secrets Endpoint

**Priority**: HIGH | **Status**: pending

Implement POST /sync/secrets to trigger replication.

### Sub-tasks
- [ ] T-008.1: Implement pathSync pattern in plugin/path_sync.go
- [ ] T-008.2: Load config from storage
- [ ] T-008.3: Create Vault client with AppRole
- [ ] T-008.4: Login to Vault
- [ ] T-008.5: For each org: list secrets, read, write to local KV
- [ ] T-008.6: Store sync status in storage

### Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| organizations | array | Specific orgs to sync (optional, syncs all if empty) |
| dry_run | bool | Preview only, do not write (optional) |

### Process Flow

1. Load config from storage
2. Create Vault client with AppRole
3. Login to Vault
4. For each organization:
   a. List secrets in org
   b. For each secret: read secret data
   c. Write to local KVv2 at kv2/org/secret
5. Store sync status in storage

### Response

```json
{
  "data": {
    "started_at": "2026-03-25T10:00:00Z",
    "status": "completed",
    "organizations_synced": 1500,
    "secrets_synced": 5000,
    "failed": 0,
    "completed_at": "2026-03-25T10:05:00Z"
  }
}
```

---

## T-009: Sync Status Endpoint

**Priority**: HIGH | **Status**: pending

Implement GET /sync/status to show current sync status.

### Sub-tasks
- [ ] T-009.1: Implement GET /sync/status path
- [ ] T-009.2: Read status from storage
- [ ] T-009.3: Return current sync state

### Response

```json
{
  "data": {
    "last_sync": "2026-03-25T10:05:00Z",
    "last_org": "organization-1500",
    "total_organizations": 1500,
    "synced_organizations": 1500,
    "total_secrets": 5000,
    "synced_secrets": 5000,
    "status": "idle",
    "last_error": null
  }
}
```

---

## T-010: Sync History Endpoint

**Priority**: HIGH | **Status**: pending

Implement GET /sync/history to list past sync operations.

### Sub-tasks
- [ ] T-010.1: Implement LIST /sync/history
- [ ] T-010.2: Implement GET /sync/history/:timestamp
- [ ] T-010.3: Store history entries in storage

### Response

LIST /sync/history:
```json
{
  "data": {
    "keys": ["2026-03-25T10:00:00Z", "2026-03-24T10:00:00Z"]
  }
}
```

GET /sync/history/:timestamp:
```json
{
  "data": {
    "timestamp": "2026-03-25T10:00:00Z",
    "status": "completed",
    "organizations_synced": 1500,
    "secrets_synced": 5000,
    "failed": 0,
    "duration_seconds": 300
  }
}
```

---

## T-011: Write to Local KVv2

**Priority**: HIGH | **Status**: pending

Implement writing replicated secrets to local OpenBao KVv2.

### Sub-tasks
- [ ] T-011.1: Create OpenBao client using token from config
- [ ] T-011.2: Write secrets to kv2/ mount
- [ ] T-011.3: Handle write errors

### Implementation

```go
func (b *Backend) writeToLocalKV(org, secret string, data map[string]interface{}) error {
    client := b.getOpenBaoClient()
    
    path := fmt.Sprintf("%s/data/%s/%s", config.DestinationMount, org, secret)
    _, err := client.Logical().Write(path, map[string]interface{}{
        "data": data,
    })
    return err
}
```

### Acceptance Criteria

- Secrets written to local KVv2
- Proper path structure maintained (kv2/org/secret)
- Error handling for write failures

---

## T-012: Progress Tracking in Storage

**Priority**: HIGH | **Status**: pending

Store sync progress in OpenBao storage.

### Storage Keys

| Key | Description |
|-----|-------------|
| config | Plugin configuration |
| sync/status | Current sync status |
| sync/history | List of past syncs |
| sync/history/:timestamp | Individual sync record |

### SyncStatus Model

```go
type SyncStatus struct {
    LastSync        time.Time `json:"last_sync"`
    LastOrg         string    `json:"last_org"`
    TotalOrgs       int       `json:"total_organizations"`
    SyncedOrgs      int       `json:"synced_organizations"`
    TotalSecrets    int       `json:"total_secrets"`
    SyncedSecrets   int       `json:"synced_secrets"`
    Status          string    `json:"status"` // idle, running, completed, failed
    LastError       string    `json:"last_error,omitempty"`
}
```

### Acceptance Criteria

- Status updated during sync
- History stored after sync
- Status persists across restarts