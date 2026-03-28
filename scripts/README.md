# Integration Testing

This directory contains scripts for running integration tests with real Vault and OpenBao instances.

## Prerequisites

- Podman (Docker replacement)
- Built plugin (`make build`)
- `podman` and `podman-compose` installed

## Quick Start

```bash
# 1. Start containers using podman-compose
podman-compose up -d

# 2. Wait for services to be ready (check podman is running)

# 3. Populate Vault with test secrets
./scripts/populate-vault.sh

# 4. Run integration test
./scripts/run-integration-test.sh

# Or run comprehensive tests (100 orgs, deletion sync, etc.)
./scripts/run-comprehensive-test.sh
```

## Scripts

### docker-compose.yml (used by podman-compose)
Runs two containers:
- **vault-source** (port 8200) - Source HashiCorp Vault
- **openbao-dest** (port 8201) - Destination OpenBao

### populate-vault.sh
Populates the source Vault with test secrets:
- org-1: api-key, database-password, jwt-secret
- org-2: aws-secret, azure-key  
- org-3: gcp-credentials, private-key
- org-4: folder structure (production/database, production/cache, staging/database)
- org-5: app-secret, encryption-key, webhook-url

Creates an AppRole with credentials for the replicator.

### run-integration-test.sh
Full integration test that:
1. Builds the plugin
2. Registers and enables the plugin in OpenBao
3. Configures the plugin with Vault credentials
4. Triggers a sync
5. Verifies secrets were replicated correctly
6. Tests dry-run mode

## Manual Testing

```bash
# Connect to Vault
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=vault-token

# Connect to OpenBao
export OPENBAO_ADDR=http://127.0.0.1:8201
export OPENBAO_TOKEN=openbao-token

# Check status
bao -address=$OPENBAO_ADDR -token=$OPENBAO_TOKEN read replicator/sync/status
bao -address=$OPENBAO_ADDR -token=$OPENBAO_TOKEN list replicator/sync/history

# Check replicated secrets
bao -address=$OPENBAO_ADDR -token=$OPENBAO_TOKEN kv list kv2/metadata/replicator/org-1
```

## Troubleshooting

### Services not starting
- Ensure Podman is running
- Check logs: `podman-compose logs vault-source` or `podman-compose logs openbao-dest`

### Cannot connect
- Wait for services to be ready (they take a few seconds to start)
- Check port availability: `lsof -i :8200` and `lsof -i :8201`

### Plugin registration fails
- Ensure plugin is built: `make build`
- Check the plugin binary exists: `ls -la dist/replicator`
- If using podman, ensure you built for Linux: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/replicator ./cmd/vault-replicator/`
