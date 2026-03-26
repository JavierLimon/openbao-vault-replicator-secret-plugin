# Metadata Sync

## Overview

This document describes what metadata is synchronized from HashiCorp Vault to OpenBao.

## KVv2 Metadata Structure

Each secret in KVv2 has two types of metadata:

### Version Metadata (Automatic)

| Field | Description | Controllable |
|-------|-------------|--------------|
| `created_time` | When version was created | ❌ No |
| `updated_time` | Last modification time | ❌ No |
| `version` | Version number (1, 2, 3...) | ❌ No |
| `deletion_time` | When secret was deleted | ❌ No |
| `destroyed` | Whether version is destroyed | ❌ No |

### Custom Metadata (User-Defined)

| Field | Description | Controllable |
|-------|-------------|--------------|
| `custom_metadata` | Arbitrary key-value pairs | ✅ Yes |

## What Gets Synced

### Synced

- Secret **data** (key-value pairs)
- **Custom metadata** - All custom metadata from source Vault secrets is preserved

### NOT Synced

- Version numbers are reset in destination
- `created_time` is reset to sync time
- `updated_time` is reset to sync time
- Source deletion history is lost

## How Custom Metadata Works

### Source (Vault)

```json
// Secret: kv2/data/org-1/database/password
{
  "data": {
    "password": "secret123"
  },
  "custom_metadata": {
    "owner": "team-dba",
    "env": "production",
    "ttl": "90d"
  }
}
```

### Destination (OpenBao)

```json
// Secret: kv2/data/replicator/org-1/database/password
{
  "data": {
    "password": "secret123"
  },
  "custom_metadata": {
    "owner": "team-dba",
    "env": "production",
    "ttl": "90d"
  }
}
```

All custom metadata key-value pairs are copied exactly as they exist in the source.

## Recommendations for Clients

1. **Use custom_metadata for tracking** - Store sync-relevant info in source custom_metadata
2. **Don't rely on version numbers** - Version will differ between Vault and OpenBao
3. **Check custom_metadata in applications** - Validate required fields exist before using secrets
