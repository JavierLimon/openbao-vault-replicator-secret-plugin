# Vault Replicator Plugin - Troubleshooting Guide

This guide helps diagnose and resolve common issues with the Vault Replicator plugin.

## Table of Contents

- [Connection Issues](#connection-issues)
- [Authentication Failures](#authentication-failures)
- [Sync Failures](#sync-failures)
- [Performance Issues](#performance-issues)
- [Plugin Errors](#plugin-errors)

---

## Connection Issues

### Issue: "dial tcp: i/o timeout" when connecting to Vault

**Symptoms:**
- Sync operations fail with timeout error
- Health check shows degraded status

**Diagnosis:**
```bash
# Test connectivity to Vault
curl -k -s https://vault.example.com:8200/sys/health

# Check Vault address in config
bao read replicator/config
```

**Solutions:**
1. Verify Vault URL is correct in config
2. Check network/firewall rules
3. Ensure Vault is accessible from OpenBao server
4. Add `vault_address` with correct protocol (https://)

---

### Issue: "x509: certificate signed by unknown authority"

**Symptoms:**
- TLS handshake errors
- Cannot connect to Vault over HTTPS

**Diagnosis:**
```bash
# Test with verbose output
curl -kv https://vault.example.com:8200/sys/health
```

**Solutions:**
1. If using self-signed cert, add `-k` flag (not recommended for production)
2. Import Vault CA certificate to system trust store
3. Configure OpenBao to trust custom CA

---

## Authentication Failures

### Issue: AppRole login fails with "invalid role_id"

**Symptoms:**
```
Error: authentication failed: invalid role_id
```

**Diagnosis:**
```bash
# Verify role exists in Vault
vault read auth/approle/role/replicator

# Check role_id in plugin config
bao read replicator/config
```

**Solutions:**
1. Regenerate role_id:
   ```bash
   vault read -field=role_id auth/approle/role/replicator
   ```
2. Update config with new role_id:
   ```bash
   bao write replicator/config approle_role_id="new-role-id"
   ```

---

### Issue: AppRole login fails with "invalid secret_id"

**Symptoms:**
```
Error: authentication failed: invalid secret_id
```

**Diagnosis:**
```bash
# Check secret_id status in Vault
vault read auth/approle/role/replicator
```

**Solutions:**
1. Generate new secret_id:
   ```bash
   vault write -f auth/approle/role/replicator/secret-id
   ```
2. Update config:
   ```bash
   bao write replicator/config approle_secret_id="new-secret-id"
   ```
3. Check secret_id TTL - may have expired

---

### Issue: "permission denied" when reading KV secrets

**Symptoms:**
- AppRole login succeeds but KV operations fail
- Error: "permission denied"

**Diagnosis:**
```bash
# Check Vault policy
vault read sys/policy/replicator-policy

# Check role's policy assignment
vault read auth/approle/role/replicator
```

**Solutions:**
1. Ensure policy includes:
   ```hcl
   path "kv2/*" {
     capabilities = ["read", "list"]
   }
   ```
2. Update role with correct policy:
   ```bash
   vault write auth/approle/role/replicator policies=replicator-policy
   ```

---

## Sync Failures

### Issue: Sync completes but secrets not replicated

**Symptoms:**
- Sync shows success but secrets missing in OpenBao
- No error messages in response

**Diagnosis:**
```bash
# Check sync status
bao read replicator/sync/status

# Check sync history
bao list replicator/sync/history

# List secrets in destination
bao list kv2/data/
```

**Solutions:**
1. Verify destination token has write permissions:
   ```bash
   bao token capabilities kv2/data/
   ```
2. Check if destination mount exists:
   ```bash
   bao secrets list -detailed
   ```
3. Verify destination mount in config

---

### Issue: Partial sync - some organizations sync, others fail

**Symptoms:**
- Sync reports both success and failure counts
- Some secrets replicated, others missing

**Diagnosis:**
```bash
# Check sync history for details
bao read replicator/sync/history/2026-03-25T18:00:00Z
```

**Solutions:**
1. Check Vault KV mount permissions per organization
2. Verify all organizations exist in Vault
3. Check network connectivity to Vault for all paths

---

### Issue: Sync hangs and never completes

**Symptoms:**
- Sync operation never returns
- Process appears stuck

**Diagnosis:**
```bash
# Check open processes
ps aux | grep replicator

# Check OpenBao logs
bao logs
```

**Solutions:**
1. Kill hanging process:
   ```bash
   # Restart OpenBao or reload plugin
   ```
2. Check for deadlocks in code
3. Increase timeout for large syncs
4. Use smaller organization batches

---

## Performance Issues

### Issue: Sync is very slow

**Symptoms:**
- Sync takes hours for small dataset
- High CPU usage during sync

**Diagnosis:**
```bash
# Check metrics
bao read replicator/metrics

# Monitor during sync
top && htop
```

**Solutions:**
1. Enable concurrent sync for organizations
2. Use selective sync for specific organizations
3. Check for network latency to Vault
4. Review Vault performance (KV backend)

---

### Issue: High memory usage during sync

**Symptoms:**
- Out of memory errors
- Plugin crashes

**Solutions:**
1. Sync in smaller batches
2. Increase OpenBao memory allocation
3. Limit concurrent operations

---

## Plugin Errors

### Issue: Plugin fails to load

**Symptoms:**
```
Plugin not found: vault-replicator
```

**Diagnosis:**
```bash
# Check registered plugins
bao read sys/plugins/catalog/vault-replicator

# Check plugin binary
ls -la dist/replicator
```

**Solutions:**
1. Re-register plugin:
   ```bash
   bao write sys/plugins/catalog/vault-replicator \
       sha_256=$(sha256sum dist/replicator | cut -d' ' -f1) \
       command="replicator"
   ```
2. Rebuild plugin:
   ```bash
   make clean && make build
   ```

---

### Issue: "plugin disabled" error

**Symptoms:**
```
Error: plugin disabled
```

**Solutions:**
1. Enable the plugin:
   ```bash
   bao secrets enable -path=replicator -plugin-name=vault-replicator plugin
   ```
2. Check plugin status:
   ```bash
  bao secrets list -path=replicator
   ```

---

## Debugging Tips

### Enable Debug Logging

1. Check OpenBao logs for detailed error messages
2. Review audit logs for request/response details
3. Use dry_run mode to preview operations

### Collect Information for Support

When reporting issues, include:
1. Plugin version: `bao read replicator/health`
2. Config (redacted): `bao read replicator/config`
3. Sync status: `bao read replicator/sync/status`
4. OpenBao version: `bao version`
5. Relevant logs from audit

### Common Error Messages Quick Reference

| Error | Likely Cause | Solution |
|-------|--------------|----------|
| connection refused | Vault not running | Start Vault |
| i/o timeout | Network/firewall | Check network |
| invalid credentials | AppRole creds invalid | Regenerate creds |
| permission denied | Policy issue | Check policy |
| not found | Wrong path | Verify paths |
| not initialized | Vault not initialized | Initialize Vault |