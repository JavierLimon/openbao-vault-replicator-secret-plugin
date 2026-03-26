# Vault Replicator Plugin - Examples

This document provides practical examples for using the Vault Replicator plugin.

## Table of Contents

- [Basic Setup](#basic-setup)
- [Configuration Examples](#configuration-examples)
- [Sync Operations](#sync-operations)
- [Advanced Usage](#advanced-usage)

---

## Basic Setup

### Step 1: Build the Plugin

```bash
# Clone the repository
git clone git@github.com:JavierLimon/openbao-vault-replicator-secret-plugin.git
cd openbao-vault-replicator-secret-plugin

# Build the plugin
make build

# Verify the binary exists
ls -la dist/replicator
```

### Step 2: Register and Enable Plugin

```bash
# Register the plugin with OpenBao
bao write sys/plugins/catalog/vault-replicator \
    sha_256=$(sha256sum dist/replicator | cut -d' ' -f1) \
    command="replicator" \
    plugin_type=secret

# Enable as secrets engine
bao secrets enable -path=replicator -plugin-name=vault-replicator plugin
```

### Step 3: Verify Installation

```bash
# Check health endpoint
bao read replicator/health

# Should return:
# {
#   "status": "ok",
#   "uptime": 60,
#   "total_requests": 5,
#   "total_errors": 0,
#   "version": "1.0.0"
# }
```

---

## Configuration Examples

### Example 1: Basic Configuration

```bash
# Configure the plugin with Vault connection details
bao write replicator/config \
    vault_address="https://vault.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="approle-role-id-value" \
    approle_secret_id="approle-secret-id-value" \
    destination_token="openbao-token-with-kv-write-access" \
    destination_mount="kv2"

# Verify configuration (shows redacted values)
bao read replicator/config

# Output:
# Key                     Value
# ---                     -----
# vault_address           https://vault.example.com:8200
# vault_mount             kv2
# destination_mount        kv2
# last_updated            2026-03-25T18:00:00Z
```

### Example 2: Configuration with Custom Organization Path

```bash
# Some Vault deployments use custom paths for organizations
bao write replicator/config \
    vault_address="https://vault.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="..." \
    approle_secret_id="..." \
    organization_path="data/organizations/" \
    destination_mount="kv2"
```

### Example 3: Using TLS with Custom CA

```bash
# For Vault with self-signed certificate
# First, add CA to system trust or configure OpenBao

bao write replicator/config \
    vault_address="https://vault.internal.company.com:8200" \
    vault_mount="kv2" \
    approle_role_id="..." \
    approle_secret_id="..." \
    destination_token="..." \
    destination_mount="kv2"
```

### Example 4: Updating Configuration

```bash
# Update specific fields (partial update)
bao write replicator/config \
    approle_secret_id="new-secret-id"

# Or update entire config
bao write replicator/config \
    vault_address="https://new-vault.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="..." \
    approle_secret_id="..." \
    destination_token="..." \
    destination_mount="kv2"

# Delete config (reset to defaults)
bao delete replicator/config
```

---

## Sync Operations

### Example 1: Full Sync

```bash
# Sync all organizations from Vault to OpenBao
bao write replicator/sync/secrets organizations=[]

# Output:
# Key              Value
# ---              -----
# status           completed
# organizations    50
# secrets_synced   1500
# duration         45s
```

### Example 2: Selective Sync

```bash
# Sync only specific organizations
bao write replicator/sync/secrets organizations="[production, staging]"

# Output:
# Key              Value
# ---              -----
# status           completed
# organizations    2
# secrets_synced   300
# duration         10s
```

### Example 3: Dry Run

```bash
# Preview what would be synced without making changes
bao write replicator/sync/secrets dry_run=true

# Output:
# Key                 Value
# ---                 -----
# status              preview
# organizations       50
# would_sync          1500
# new_secrets         100
# updated_secrets     1400
```

### Example 4: Check Sync Status

```bash
# Get current sync status
bao read replicator/sync/status

# Output:
# Key                  Value
# ---                  -----
# last_sync            2026-03-25T18:00:00Z
# last_status          completed
# organizations_synced    50
# secrets_synced       1500
# failed               0
```

### Example 5: View Sync History

```bash
# List past sync operations
bao list replicator/sync/history

# Output:
# Keys
# ----
# 2026-03-25T18:00:00Z
# 2026-03-24T18:00:00Z
# 2026-03-23T18:00:00Z

# Get details of a specific sync
bao read replicator/sync/history/2026-03-25T18:00:00Z

# Output:
# Key              Value
# ---              -----
# timestamp        2026-03-25T18:00:00Z
# status           completed
# organizations    50
# secrets_synced   1500
# duration         45s
# errors           []
```

---

## Advanced Usage

### Example 1: Role-Based Sync

```bash
# Create multiple roles for different Vault sources

# Role for production Vault
bao write replicator/roles/prod-vault \
    vault_address="https://vault.prod.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="prod-role-id" \
    approle_secret_id="prod-secret-id" \
    destination_mount="kv2"

# Role for staging Vault
bao write replicator/roles/staging-vault \
    vault_address="https://vault.staging.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="staging-role-id" \
    approle_secret_id="staging-secret-id" \
    destination_mount="kv2-staging"

# List available roles
bao list replicator/roles

# Use specific role for sync
bao write replicator/sync/secrets role=prod-vault organizations=[]
```

### Example 2: Scheduled Sync with Cron

```bash
# Create a simple sync script
cat > /usr/local/bin/sync-vault.sh << 'EOF'
#!/bin/bash

# Sync secrets from Vault to OpenBao
OUTPUT=$(bao write -f replicator/sync/secrets organizations=[] 2>&1)

# Check if sync was successful
if echo "$OUTPUT" | grep -q "status.*completed"; then
    echo "Sync completed successfully at $(date)"
else
    echo "Sync failed at $(date): $OUTPUT"
    exit 1
fi
EOF

chmod +x /usr/local/bin/sync-vault.sh

# Add to crontab (run every hour)
# 0 * * * * /usr/local/bin/sync-vault.sh >> /var/log/vault-sync.log 2>&1
```

### Example 3: Monitoring with Metrics

```bash
# Get current metrics
bao read replicator/metrics

# Output:
# Key                   Value
# ---                   -----
# total_requests        150
# total_errors          2
# sync_total            3
# sync_completed        3
# sync_failed           0
# secrets_replicated    1500

# Use in monitoring system (Prometheus example)
# Add to prometheus.yml:
# - job_name: 'openbao-replicator'
#   static_configs:
#     - targets: ['localhost:8200']
#   metrics_path: '/v1/replicator/metrics'
```

### Example 4: Recovery After Failed Sync

```bash
# If sync fails, you can:

# 1. Check what went wrong
bao read replicator/sync/status
bao read replicator/sync/history/2026-03-25T18:00:00Z

# 2. Retry with specific organizations
bao write replicator/sync/secrets organizations="[org1, org2]"

# 3. Force a full re-sync (delete destination and sync fresh)
# WARNING: This will delete all replicated secrets
# Only do this if you have a backup or can re-sync from Vault

# First, check which secrets need resync
bao write replicator/sync/secrets dry_run=true

# Then run full sync
bao write replicator/sync/secrets organizations=[]
```

---

## Integration Examples

### Example 1: Using with Terraform

```hcl
# main.tf
variable "vault_address" {}
variable "approle_role_id" {}
variable "approle_secret_id" {}
variable "openbao_token" {}

provider "bao" {
  address = "https://openbao.example.com:8200"
  token   = var.openbao_token
}

resource "bao_plugin_catalog" "replicator" {
  name        = "vault-replicator"
  type        = "secret"
  command     = "replicator"
  sha256      = file("${path.module}/dist/replicator")
}

resource "bao_secret_backend" "replicator" {
  type        = "plugin"
  plugin_name = "vault-replicator"
  path        = "replicator"
}

resource "bao_plugin_configuration" "config" {
  backend = bao_secret_backend.replicator.path
  config {
    vault_address       = var.vault_address
    vault_mount         = "kv2"
    approle_role_id     = var.approle_role_id
    approle_secret_id   = var.approle_secret_id
    destination_token   = var.openbao_token
    destination_mount   = "kv2"
  }
}
```

---

## Tips and Best Practices

1. **Use descriptive organization names** - Makes sync history easier to understand
2. **Test with dry_run first** - Always preview changes before applying
3. **Monitor metrics** - Set up alerts for failed syncs or errors
4. **Schedule during low traffic** - Large syncs can impact performance
5. **Keep credentials secure** - Use Vault's secrets engine for AppRole creds
6. **Regular health checks** - Monitor the health endpoint for issues