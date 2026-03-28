# OpenBao Vault Replicator Plugin - API Documentation

Secret engine plugin for replicating secrets from HashiCorp Vault (KVv2) to OpenBao (KVv2).

**Plugin Mount Path:** `replicator/`

---

## Table of Contents

- [Configuration](#configuration)
  - [Read Configuration](#get-config)
  - [Create/Update Configuration](#create--update-config)
  - [Delete Configuration](#delete-config)
- [Sync Operations](#sync-operations)
  - [Trigger Sync](#post-syncsecrets)
  - [Get Sync Status](#get-syncstatus)
  - [List Sync History](#get-synchistory)
  - [Get Sync History Entry](#get-synchistorytimestamp)

---

## Configuration

### Configuration Object

| Field | Type | Description |
|-------|------|-------------|
| `vault_address` | string | Vault server URL (e.g., `https://vault.example.com:8200`) |
| `vault_mount` | string | Vault KVv2 mount path (default: `kv2`) |
| `approle_role_id` | string | AppRole role_id for Vault authentication |
| `approle_secret_id` | string | AppRole secret_id for Vault authentication |
| `destination_token` | string | OpenBao token for writing replicated secrets |
| `destination_mount` | string | OpenBao KVv2 mount for storing secrets (default: `kv2`) |
| `org_skip_list` | array | Organizations to skip during sync (blacklist) |
| `allow_deletion_sync` | bool | Enable deletion sync (delete secrets in destination when deleted in source) |
| `org_deletion_overrides` | object | Per-organization deletion sync overrides (e.g., `{"org-1": false}`) |

**Note:** When reading configuration, `approle_secret_id` and `destination_token` are masked for security.

**Note:** `organization_path` is no longer required - KVv2 always uses `metadata/` for listing and `data/` for read/write.

---

## GET /config

Read the current plugin configuration.

#### Request

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

#### Response

```json
{
  "data": {
    "vault_address": "https://vault.example.com:8200",
    "vault_mount": "kv2",
    "approle_role_id": "my-role-id",
    "approle_secret_id": "",
    "destination_token": "[MASKED]",
    "destination_mount": "kv2",
    "org_skip_list": [],
    "allow_deletion_sync": false,
    "org_deletion_overrides": {}
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `vault_address` | string | Configured Vault server URL |
| `vault_mount` | string | Vault KVv2 mount path |
| `approle_role_id` | string | AppRole role_id (secret is masked) |
| `approle_secret_id` | string | Always empty string on read |
| `destination_token` | string | Always `[MASKED]` on read |
| `destination_mount` | string | OpenBao KVv2 destination mount |
| `org_skip_list` | array | Organizations to skip during sync |
| `allow_deletion_sync` | bool | Enable deletion sync |
| `org_deletion_overrides` | object | Per-org deletion overrides |

---

## POST /config

Create or update the plugin configuration.

#### Request Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `vault_address` | string | Yes | Vault server URL |
| `vault_mount` | string | No | Vault KVv2 mount (default: `kv2`) |
| `approle_role_id` | string | Yes | AppRole role_id |
| `approle_secret_id` | string | Yes | AppRole secret_id |
| `destination_token` | string | Yes | OpenBao token for writes |
| `destination_mount` | string | No | OpenBao KVv2 mount (default: `kv2`) |
| `org_skip_list` | array | No | Organizations to skip (e.g., `["org-3", "org-4"]`) |
| `allow_deletion_sync` | bool | No | Enable deletion sync (default: `false`) |
| `org_deletion_overrides` | object | No | Per-org deletion overrides (e.g., `{"org-1": false}`) |

#### Request

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "vault_address": "https://vault.example.com:8200",
    "vault_mount": "kv2",
    "approle_role_id": "my-role-id",
    "approle_secret_id": "my-secret-id",
    "destination_token": "openbao-token-here",
    "destination_mount": "kv2",
    "org_skip_list": ["test-org", "deprecated-org"],
    "allow_deletion_sync": false,
    "org_deletion_overrides": {"org-1": false}
  }'
```

**Example with deletion sync enabled:**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "vault_address": "https://vault.example.com:8200",
    "vault_mount": "kv2",
    "approle_role_id": "my-role-id",
    "approle_secret_id": "my-secret-id",
    "destination_token": "openbao-token-here",
    "destination_mount": "kv2",
    "allow_deletion_sync": true,
    "org_deletion_overrides": {
      "org-sensitive": false,
      "org-keep-forever": false
    }
  }'
```

#### Response

```json
{
  "data": null
}
```

Returns empty data on success.

---

## DELETE /config

Delete the plugin configuration.

#### Request

```bash
curl -X DELETE http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

#### Response

```json
{
  "data": null
}
```

Returns empty data on success.

---

## Sync Operations

### POST /sync/secrets

Trigger secret replication from HashiCorp Vault to OpenBao.

#### Request Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `organizations` | array | No | Specific organizations to sync. If empty, syncs all organizations |
| `dry_run` | bool | No | Preview only, do not write to destination |
| `deletion_sync` | bool | No | Enable deletion sync (delete secrets in destination that don't exist in source). Requires `allow_deletion_sync=true` in config for each org |

#### Request

**Sync all organizations:**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/sync/secrets \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Sync specific organizations:**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/sync/secrets \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "organizations": ["org-1", "org-2", "org-3"]
  }'
```

**Dry run (preview only):**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/sync/secrets \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "dry_run": true
  }'
```

**Sync with deletion sync (delete orphans):**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/sync/secrets \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "deletion_sync": true
  }'
```

**Safety:** Deletion sync will be skipped for an organization if Vault returned an error when listing secrets (e.g., network issues, timeouts). This prevents accidental deletion when Vault is unavailable.

#### Response

```json
{
  "data": {
    "started_at": "2026-03-25T10:00:00Z",
    "status": "completed",
    "organizations_synced": 1500,
    "secrets_synced": 5000,
    "deleted_secrets": 10,
    "failed": 0,
    "completed_at": "2026-03-25T10:05:00Z",
    "duration_seconds": 300
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `started_at` | string | RFC3339 timestamp when sync started |
| `status` | string | Overall status: `completed`, `running`, or `failed` |
| `organizations_synced` | int | Number of organizations successfully synced |
| `secrets_synced` | int | Number of secrets successfully synced |
| `deleted_secrets` | int | Number of secrets deleted in destination (when `allow_deletion_sync` is enabled) |
| `failed` | int | Number of organizations/secrets that failed |
| `completed_at` | string | RFC3339 timestamp when sync completed |
| `duration_seconds` | int | Total sync duration in seconds |

---

### GET /sync/status

Get the current sync status and progress.

#### Request

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/sync/status \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

#### Response (idle state)

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

#### Response (never synced)

```json
{
  "data": {
    "last_sync": null,
    "last_org": null,
    "total_organizations": 0,
    "synced_organizations": 0,
    "total_secrets": 0,
    "synced_secrets": 0,
    "status": "idle",
    "last_error": null
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `last_sync` | string | RFC3339 timestamp of last sync completion (null if never) |
| `last_org` | string | Last organization processed |
| `total_organizations` | int | Total organizations in sync operation |
| `synced_organizations` | int | Organizations successfully synced |
| `total_secrets` | int | Total secrets in sync operation |
| `synced_secrets` | int | Secrets successfully synced |
| `status` | string | Current status: `idle`, `running`, `completed`, or `failed` |
| `last_error` | string | Last error message (null if no errors) |

---

### GET /sync/history

List past sync operations or get the most recent sync entry.

#### Request

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/sync/history \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

#### Response (list mode)

```json
{
  "data": {
    "keys": [
      "2026-03-25T10:00:00Z",
      "2026-03-24T10:00:00Z",
      "2026-03-23T10:00:00Z"
    ]
  }
}
```

#### Response (single entry when only one exists)

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

| Field | Type | Description |
|-------|------|-------------|
| `keys` | array | List of RFC3339 timestamps for each sync operation |
| `timestamp` | string | RFC3339 timestamp of this sync |
| `status` | string | Sync status: `completed` or `failed` |
| `organizations_synced` | int | Organizations successfully synced |
| `secrets_synced` | int | Secrets successfully synced |
| `failed` | int | Count of failures |
| `duration_seconds` | int | Sync duration in seconds |

---

### GET /sync/history/:timestamp

Get a specific sync history entry by timestamp.

#### Request Parameters

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | RFC3339 timestamp of the sync history entry |

#### Request

```bash
curl -X GET "http://127.0.0.1:8200/v1/replicator/sync/history/2026-03-25T10:00:00Z" \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

#### Response

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

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | RFC3339 timestamp of the sync |
| `status` | string | Sync status: `completed` or `failed` |
| `organizations_synced` | int | Organizations successfully synced |
| `secrets_synced` | int | Secrets successfully synced |
| `failed` | int | Count of failures |
| `duration_seconds` | int | Sync duration in seconds |

---

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "errors": [
    "configuration not found"
  ]
}
```

### Common Error Messages

| Endpoint | Error | Cause |
|----------|-------|-------|
| `/sync/secrets` | `configuration not found` | Configuration not yet set |
| `/sync/secrets` | `failed to create Vault client: ...` | Invalid vault_address |
| `/sync/secrets` | `failed to login to Vault: ...` | Invalid AppRole credentials |
| `/sync/secrets` | `failed to list organizations: ...` | Invalid organization_path |
| `/sync/history/:timestamp` | `sync history entry not found` | Invalid or expired timestamp |
| All | Various | Storage or internal errors |

---

## Usage Examples

### Complete Setup and Sync Flow

**1. Configure the plugin:**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "vault_address": "https://vault.example.com:8200",
    "vault_mount": "kv2",
    "approle_role_id": "your-role-id",
    "approle_secret_id": "your-secret-id",
    "destination_token": "your-openbao-token",
    "destination_mount": "kv2",
    "org_skip_list": ["test-org"],
    "allow_deletion_sync": false
  }'
```

**2. Verify configuration:**

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/config \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

**3. Trigger a sync:**

```bash
curl -X POST http://127.0.0.1:8200/v1/replicator/sync/secrets \
  -H "X-Vault-Token: ${VAULT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**4. Check sync status:**

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/sync/status \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

**5. View sync history:**

```bash
curl -X GET http://127.0.0.1:8200/v1/replicator/sync/history \
  -H "X-Vault-Token: ${VAULT_TOKEN}"
```

---

## Storage Keys

The plugin stores data under these internal storage paths:

| Path | Description |
|------|-------------|
| `config` | Plugin configuration |
| `sync/status` | Current sync status |
| `sync/history` | List of all sync timestamps |
| `sync/history/:timestamp` | Individual sync history entry |

---

## ACL Policies

To use the plugin, you need to create appropriate policies in OpenBao.

### Full Access Policy

Grants complete access to all plugin endpoints:

```hcl
path "replicator/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

### Read-Only Policy

For monitoring and viewing configuration:

```hcl
path "replicator/config" {
  capabilities = ["read"]
}

path "replicator/sync/status" {
  capabilities = ["read"]
}

path "replicator/sync/history" {
  capabilities = ["list", "read"]
}

path "replicator/health" {
  capabilities = ["read"]
}

path "replicator/metrics" {
  capabilities = ["read"]
}
```

### Sync-Only Policy

Allows triggering syncs but not modifying configuration:

```hcl
path "replicator/config" {
  capabilities = ["read"]
}

path "replicator/sync/secrets" {
  capabilities = ["create", "update"]
}

path "replicator/sync/status" {
  capabilities = ["read"]
}

path "replicator/sync/history" {
  capabilities = ["list", "read"]
}
```

### Operator Policy

Full access plus ability to manage the plugin:

```hcl
# Plugin management
path "sys/plugins/catalog/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Plugin endpoints
path "replicator/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

### Vault Source Policy (AppRole)

The plugin requires an AppRole in Vault with this policy to read secrets:

```hcl
# List organizations
path "kv2/metadata/*" {
  capabilities = ["list"]
}

# Read secrets (all orgs)
path "kv2/data/*" {
  capabilities = ["read"]
}

# Read secret metadata (for custom_metadata sync)
path "kv2/metadata/*" {
  capabilities = ["read"]
}
```

**Note:** Replace `kv2` with your actual Vault KVv2 mount path.

### Minimum Required Vault Policy

For a specific subset of organizations:

```hcl
# List only specific orgs
path "kv2/metadata/org-1/*" {
  capabilities = ["list"]
}

path "kv2/metadata/org-2/*" {
  capabilities = ["list"]
}

# Read secrets from specific orgs
path "kv2/data/org-1/*" {
  capabilities = ["read"]
}

path "kv2/data/org-2/*" {
  capabilities = ["read"]
}

# Read metadata from specific orgs
path "kv2/metadata/org-1/*" {
  capabilities = ["read"]
}

path "kv2/metadata/org-2/*" {
  capabilities = ["read"]
}
```
