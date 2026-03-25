# OpenBAO Vault Replicator Plugin

Secret engine plugin that replicates secrets from HashiCorp Vault (KVv2) to OpenBao (KVv2).

## Overview

The Vault Replicator plugin enables one-way replication of secrets from a HashiCorp Vault instance to an OpenBao instance. It uses AppRole authentication to read from Vault and writes to the local OpenBao KVv2 mount.

## Architecture

- **Source**: HashiCorp Vault with KVv2 at kv2/ (1500+ organizations, shared mount)
- **Destination**: OpenBao with KVv2 at kv2/
- **Mount Path**: replicator/

## Auth Methods

- **Source (Vault)**: AppRole - role_id + secret_id
- **Destination (OpenBao)**: Token (stored in config)

## Installation

```bash
# Build the plugin
make build

# Register the plugin
bao write sys/plugins/catalog/vault-replicator \
    sha_256=$(sha256sum vault-openbao-replicator | cut -d' ' -f1) \
    command="vault-openbao-replicator"

# Enable as secrets engine
bao secrets enable -path=replicator -plugin-name=vault-replicator plugin
```

## Configuration

```bash
# Configure the plugin
bao write replicator/config \
    vault_address="https://vault.example.com:8200" \
    vault_mount="kv2" \
    approle_role_id="your-role-id" \
    approle_secret_id="your-secret-id" \
    destination_token="your-openbao-token" \
    destination_mount="kv2" \
    organization_path="data/"
```

## Usage

### Trigger Secret Sync

```bash
# Sync all organizations
bao write replicator/sync/secrets organizations=[]

# Sync specific organizations
bao write replicator/sync/secrets organizations="[org1,org2]"

# Dry run (preview only)
bao write replicator/sync/secrets dry_run=true
```

### Check Sync Status

```bash
bao read replicator/sync/status
```

### View Sync History

```bash
# List past syncs
bao list replicator/sync/history

# Get specific sync details
bao read replicator/sync/history/2026-03-25T10:00:00Z
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| /config | CRUD | Plugin configuration |
| /sync/secrets | POST | Trigger secret replication |
| /sync/status | GET | Show sync status |
| /sync/history | LIST | List past operations |
| /sync/history/:timestamp | GET | Get operation details |

## Requirements

- Go 1.25+
- OpenBao SDK v2
- HashiCorp Vault API

## License

MIT