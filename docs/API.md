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
| `organization_path` | string | Path in Vault where organizations live (e.g., `data/`) |

**Note:** When reading configuration, `approle_secret_id` and `destination_token` are masked for security.

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
    "organization_path": "data/"
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
| `organization_path` | string | Organization base path in Vault |

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
| `organization_path` | string | Yes | Base path for organizations |

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
    "organization_path": "data/"
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

#### Response

```json
{
  "data": {
    "started_at": "2026-03-25T10:00:00Z",
    "status": "completed",
    "organizations_synced": 1500,
    "secrets_synced": 5000,
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
    "organization_path": "data/"
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
