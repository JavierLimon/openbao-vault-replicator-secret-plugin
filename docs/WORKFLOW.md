# Vault Replicator Plugin - Workflow Documentation

This document describes the workflows for using the Vault Replicator plugin with ASCII diagrams.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Initial Setup Workflow](#initial-setup-workflow)
- [Configuration Workflow](#configuration-workflow)
- [Secret Replication Workflow](#secret-replication-workflow)
- [AppRole Authentication Flow](#approle-authentication-flow)
- [Sync Operations](#sync-operations)
- [Monitoring and Health](#monitoring-and-health)
- [Role Management](#role-management)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                      OpenBao Server                                │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │              Vault Replicator Plugin                            ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ ││
│  │  │    Config     │  │     Sync     │  │      Roles           │ ││
│  │  │   (KV Store)  │  │   (State)    │  │   (AppRole Mgmt)     │ ││
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ ││
│  │         │                  │                   │                ││
│  │         └──────────────────┼───────────────────┘                ││
│  │                            ▼                                     ││
│  │  ┌─────────────────────────────────────────────────────────────┐││
│  │  │              Replication Engine                             │││
│  │  │   ┌─────────────────┐    ┌─────────────────┐               │││
│  │  │   │ Vault Client    │    │ OpenBao Client  │               │││
│  │  │   │ (AppRole Auth)  │───▶│  (Local KV v2)  │               │││
│  │  │   └─────────────────┘    └─────────────────┘               │││
│  │  └─────────────────────────────────────────────────────────────┘││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                     HashiCorp Vault                                │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │  KVv2 Engine (kv2/)                                            ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ ││
│  │  │ organization1 │  │ organization2 │  │ organizationN       │ ││
│  │  │   secrets/    │  │   secrets/   │  │   secrets/          │ ││
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘ ││
│  │                                                                  ││
│  │  AppRole Auth Method                                            ││
│  │  ┌──────────────┐                                              ││
│  │  │ role_id      │ ──┐                                          ││
│  │  │ secret_id    │ ──┼──▶ Access secrets                       ││
│  │  └──────────────┘    │                                          ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

---

## Initial Setup Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Initial Setup Workflow                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Step 1: Build Plugin                                              │
│  ─────────────────────                                              │
│  $ make build                                                      │
│       │                                                            │
│       ▼                                                            │
│  dist/replicator (compiled binary)                                  │
│                                                                     │
│  Step 2: Register Plugin                                           │
│  ─────────────────────────                                          │
│  $ bao write sys/plugins/catalog/vault-replicator \                │
│      sha_256=$(sha256sum dist/replicator | cut -d' ' -f1) \        │
│      command="replicator"                                           │
│       │                                                            │
│       ▼                                                            │
│  Plugin registered in catalog                                       │
│                                                                     │
│  Step 3: Enable Plugin                                             │
│  ─────────────────────                                              │
│  $ bao secrets enable -path=replicator -plugin-name=vault-replicator \
│       plugin                                                       │
│       │                                                            │
│       ▼                                                            │
│  Plugin mounted at replicator/                                     │
│                                                                     │
│  Step 4: Verify Health                                             │
│  ─────────────────────                                              │
│  $ bao read replicator/health                                      │
│       │                                                            │
│       ▼                                                            │
│  {                                                                  │
│    "status": "ok",                                                  │
│    "uptime": 3600,                                                 │
│    "version": "1.0.0"                                              │
│  }                                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Configuration Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                  Configuration Workflow                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Configure Source (Vault)                                           │
│  ──────────────────────────                                         │
│  $ bao write replicator/config                                      │
│      vault_address="https://vault.example.com:8200"                │
│      vault_mount="kv2"                                             │
│      approle_role_id="your-role-id"                                │
│      approle_secret_id="your-secret-id"                            │
│      organization_path="data/"                                     │
│                    │                                                │
│                    ▼                                                │
│  ┌──────────────────────────────────────────┐                       │
│  │  Stored in OpenBao storage:             │                       │
│  │  replicator/config                       │                       │
│  │  {                                       │                       │
│  │    "vault_address": "...",               │                       │
│  │    "vault_mount": "kv2",                 │                       │
│  │    "role_id": "...",                     │                       │
│  │    "destination_mount": "kv2"           │                       │
│  │  }                                       │                       │
│  └──────────────────────────────────────────┘                       │
│                                                                     │
│  Configure Destination (OpenBao)                                    │
│  ──────────────────────────────                                     │
│  $ bao write replicator/config                                     │
│      ...                                                           │
│      destination_token="your-openbao-token"                        │
│      destination_mount="kv2"                                      │
│                    │                                                │
│                    ▼                                                │
│  ┌──────────────────────────────────────────┐                       │
│  │  Token stored in storage                 │                       │
│  │  (encrypted at rest)                     │                       │
│  └──────────────────────────────────────────┘                       │
│                                                                     │
│  Verify Configuration                                              │
│  ─────────────────────                                              │
│  $ bao read replicator/config                                      │
│                                                                     │
│  {                                                                  │
│    "vault_address": "https://...",                                 │
│    "vault_mount": "kv2",                                           │
│    "destination_mount": "kv2",                                    │
│    "last_updated": "2026-03-25T..."                               │
│  }                                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Secret Replication Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                 Secret Replication Workflow                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Step 1: Trigger Sync                                              │
│  ─────────────────────                                              │
│  $ bao write replicator/sync/secrets organizations=[]             │
│       │                                                            │
│       ▼                                                            │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Replication Engine Starts                                   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│       │                                                            │
│       ▼                                                            │
│  Step 2: Authenticate to Vault                                     │
│  ─────────────────────────────                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  AppRole Auth:                                              │   │
│  │  1. POST /auth/approle/login (role_id + secret_id)         │   │
│  │  2. Vault returns: {token: "..."}                          │   │
│  │  3. Use token for KV operations                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│       │                                                            │
│       ▼                                                            │
│  Step 3: List Organizations                                         │
│  ──────────────────────────                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  GET /kv2/metadata (list all paths)                        │   │
│  │  Returns: [organization1, organization2, ...]              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│       │                                                            │
│       ▼                                                            │
│  Step 4: For Each Organization                                      │
│  ──────────────────────────                                         │
│                                                                     │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │  For org in organizations:                                │    │
│  │    1. GET /kv2/data/org/secret-path                       │    │
│  │    2. Store in local: kv2/org/secret-path                  │    │
│  │    3. Record sync status                                   │    │
│  └────────────────────────────────────────────────────────────┘    │
│       │                                                            │
│       ▼                                                            │
│  Step 5: Report Results                                            │
│  ───────────────────────                                            │
│  $ bao read replicator/sync/status                                 │
│       │                                                            │
│       ▼                                                            │
│  {                                                                  │
│    "status": "completed",                                          │
│    "total_secrets": 1500,                                          │
│    "synced": 1500,                                                 │
│    "failed": 0,                                                   │
│    "duration": "45s"                                              │
│  }                                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## AppRole Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│               AppRole Authentication Flow                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. Configure AppRole in Vault                                      │
│  ─────────────────────────────────────                              │
│  $ vault auth enable approle                                        │
│       │                                                            │
│       ▼                                                            │
│  $ vault write auth/approle/role/replicator \                      │
│      secret_id_ttl=10m                     \                      │
│      token_ttl=20m                         \                      │
│      secret_id_num_uses=10                 \                      │
│      policies=replicator-policy                                    │
│       │                                                            │
│       ▼                                                            │
│  Role created: role_id generated                                   │
│                                                                     │
│  2. Get Credentials                                                │
│  ────────────────────                                              │
│  $ vault read -field=role_id auth/approle/role/replicator        │
│  role_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx                    │
│                                                                     │
│  $ vault write -f auth/approle/role/replicator/secret-id          │
│  secret_id=yyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy                       │
│                                                                     │
│  3. Plugin Uses Credentials                                        │
│  ───────────────────────────────                                   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  POST https://vault.example.com:8200/v1/auth/approle/login │   │
│  │  {                                                           │   │
│  │    "role_id": "xxx",                                        │   │
│  │    "secret_id": "yyy"                                       │   │
│  │  }                                                           │   │
│  │                               │                               │   │
│  │                               ▼                               │   │
│  │  Response:                                                  │   │
│  │  {                                                           │   │
│  │    "auth": {"client_token": "token-value", ...}            │   │
│  │  }                                                           │   │
│  └─────────────────────────────────────────────────────────────┘   │
│       │                                                            │
│       ▼                                                            │
│  4. Use Token for KV Operations                                    │
│  ──────────────────────────────                                    │
│  GET /v1/kv2/data/org/secret-path                                 │
│  Authorization: Bearer token-value                                │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Sync Operations

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Sync Operations                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Full Sync (All Organizations)                                      │
│  ─────────────────────────────                                      │
│  $ bao write replicator/sync/secrets organizations=[]             │
│       │                                                            │
│       ▼                                                            │
│  Syncs ALL organizations from Vault → OpenBao                      │
│                                                                     │
│  Selective Sync (Specific Organizations)                           │
│  ───────────────────────────────────────                            │
│  $ bao write replicator/sync/secrets \                             │
│      organizations="[org1, org2, org3]"                           │
│       │                                                            │
│       ▼                                                            │
│  Syncs ONLY specified organizations                                │
│                                                                     │
│  Dry Run (Preview Only)                                             │
│  ──────────────────────────                                        │
│  $ bao write replicator/sync/secrets dry_run=true                 │
│       │                                                            │
│       ▼                                                            │
│  Shows what would be synced without making changes                  │
│                                                                     │
│  Check Sync Status                                                  │
│  ────────────────────                                              │
│  $ bao read replicator/sync/status                                 │
│                                                                     │
│  {                                                                  │
│    "last_sync": "2026-03-25T18:00:00Z",                            │
│    "organizations_synced": 50,                                     │
│    "secrets_synced": 1500,                                         │
│    "status": "completed"                                           │
│  }                                                                  │
│                                                                     │
│  View Sync History                                                  │
│  ────────────────────                                              │
│  $ bao list replicator/sync/history                                │
│  2026-03-25T18:00:00Z                                              │
│  2026-03-24T18:00:00Z                                              │
│  2026-03-23T18:00:00Z                                              │
│                                                                     │
│  $ bao read replicator/sync/history/2026-03-25T18:00:00Z          │
│                                                                     │
│  {                                                                  │
│    "timestamp": "2026-03-25T18:00:00Z",                            │
│    "status": "completed",                                          │
│    "organizations": 50,                                            │
│    "secrets_synced": 1500,                                         │
│    "duration": "45s"                                               │
│  }                                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Monitoring and Health

```
┌─────────────────────────────────────────────────────────────────────┐
│               Monitoring and Health                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Health Check Endpoint                                              │
│  ─────────────────────────                                          │
│  $ curl http://localhost:8200/v1/replicator/health                 │
│       │                                                            │
│       ▼                                                            │
│  {                                                                  │
│    "status": "ok",                                                  │
│    "uptime": 3600,                                                  │
│    "total_requests": 150,                                          │
│    "total_errors": 2,                                               │
│    "version": "1.0.0"                                              │
│  }                                                                  │
│                                                                     │
│  Metrics Endpoint                                                   │
│  ─────────────────                                                  │
│  $ curl http://localhost:8200/v1/replicator/metrics               │
│       │                                                            │
│       ▼                                                            │
│  {                                                                  │
│    "total_requests": 150,                                          │
│    "total_errors": 2,                                              │
│    "sync_total": 3,                                                │
│    "sync_completed": 3,                                            │
│    "sync_failed": 0,                                               │
│    "secrets_replicated": 1500                                      │
│  }                                                                  │
│                                                                     │
│  Audit Logs                                                         │
│  ───────────                                                        │
│  All operations logged to OpenBao audit log:                       │
│                                                                     │
│  {                                                                  │
│    "time": "2026-03-25T18:00:00Z",                                 │
│    "type": "request",                                              │
│    "auth": {...},                                                  │
│    "request": {                                                    │
│      "path": "replicator/sync/secrets",                            │
│      "operation": "create"                                         │
│    }                                                               │
│  }                                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Role Management

```
┌─────────────────────────────────────────────────────────────────────┐
│                   Role Management                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Create Role                                                        │
│  ─────────────                                                      │
│  $ bao write replicator/roles/my-role \                            │
│      vault_address="https://vault.example.com:8200"               │
│      vault_mount="kv2"                                             │
│      approle_role_id="xxx"                                         │
│      approle_secret_id="yyy"                                       │
│      destination_mount="kv2"                                      │
│       │                                                            │
│       ▼                                                            │
│  {                                                                  │
│    "role_name": "my-role",                                         │
│    "status": "active"                                              │
│  }                                                                  │
│                                                                     │
│  List Roles                                                         │
│  ───────────                                                        │
│  $ bao list replicator/roles                                       │
│  my-role                                                            │
│  production                                                         │
│  staging                                                            │
│                                                                     │
│  Read Role Details                                                  │
│  ────────────────────                                              │
│  $ bao read replicator/roles/my-role                               │
│                                                                     │
│  {                                                                  │
│    "role_name": "my-role",                                         │
│    "vault_address": "https://...",                                 │
│    "vault_mount": "kv2",                                           │
│    "destination_mount": "kv2",                                    │
│    "status": "active",                                             │
│    "created": "2026-03-25T..."                                    │
│  }                                                                  │
│                                                                     │
│  Delete Role                                                        │
│  ───────────                                                        │
│  $ bao delete replicator/roles/my-role                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Troubleshooting

See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for detailed troubleshooting steps.

---

## Best Practices

1. **Security**
   - Rotate AppRole credentials regularly
   - Use Vault's namespace feature for multi-tenant isolation
   - Enable TLS for Vault communication

2. **Performance**
   - Sync during low-traffic periods
   - Use selective sync for large deployments
   - Monitor metrics endpoint for bottlenecks

3. **Reliability**
   - Configure appropriate TTL for tokens
   - Set up monitoring alerts for failed syncs
   - Review sync history regularly

4. **Recovery**
   - Keep audit logs for compliance
   - Document sync procedures
   - Test recovery procedures regularly