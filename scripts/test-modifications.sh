#!/bin/bash
set -e

TOKEN="s.JmTjr3IIxy34BSq9BPD8DpcV"
VAULT_TOKEN="vault-token"
VAULT_ADDR="http://127.0.0.1:8200"
OPENBAO_ADDR="http://127.0.0.1:8201"

echo "=== Test: Modify data in Vault, check sync behavior ==="
echo ""

# Step 1: Delete 20 organizations (org-1 to org-20)
echo "=== 1. Deleting 20 organizations (org-1 to org-20) ==="
for i in $(seq 1 20); do
    org="org-$i"
    # Delete all secrets in org - this should delete the org metadata
    curl -s -X DELETE -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/metadata/$org" > /dev/null 2>&1
done
echo "Deleted orgs 1-20"

# Step 2: Delete 40 root-level secrets (orgs 21-40)
echo ""
echo "=== 2. Deleting 40 root-level secrets (orgs 21-40) ==="
for i in $(seq 21 40); do
    org="org-$i"
    curl -s -X DELETE -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/$org/root-key" > /dev/null 2>&1
done
echo "Deleted root-key from orgs 21-40"

# Step 3: Delete 50 secrets at third level (production/db-password and staging/db-password from orgs 41-65)
echo ""
echo "=== 3. Deleting 50 nested secrets (production/db-password, staging/db-password from orgs 41-65) ==="
for i in $(seq 41 65); do
    org="org-$i"
    curl -s -X DELETE -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/$org/production/db-password" > /dev/null 2>&1
    curl -s -X DELETE -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/$org/staging/db-password" > /dev/null 2>&1
done
echo "Deleted nested secrets from orgs 41-65"

# Step 4: Update 5 secrets at each level
echo ""
echo "=== 4. Updating 5 secrets at each level ==="

# Level 1: root secrets (orgs 66-70)
for i in $(seq 66 70); do
    org="org-$i"
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/root-key" \
        -d "{\"data\": {\"key\": \"UPDATED-root-value-$i\", \"org\": \"$org\", \"updated\": \"yes\"}}" > /dev/null 2>&1
done
echo "Updated root secrets in orgs 66-70"

# Level 2: production secrets (orgs 71-75)
for i in $(seq 71 75); do
    org="org-$i"
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/production/db-password" \
        -d "{\"data\": {\"password\": \"UPDATED-prod-pass-$i\", \"updated\": \"yes\"}}" > /dev/null 2>&1
done
echo "Updated production/db-password in orgs 71-75"

# Level 3: staging secrets (orgs 76-80)
for i in $(seq 76 80); do
    org="org-$i"
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/staging/db-password" \
        -d "{\"data\": {\"password\": \"UPDATED-stage-pass-$i\", \"updated\": \"yes\"}}" > /dev/null 2>&1
done
echo "Updated staging/db-password in orgs 76-80"

# Step 5: Trigger sync
echo ""
echo "=== 5. Running second sync ==="
SYNC_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/replicator/sync/secrets" -d '{"dry_run": false}')
echo "$SYNC_RESP" | jq '.data'

# Step 6: Verify - Compare Vault vs OpenBao
echo ""
echo "=== 6. Verification: Comparing Vault vs OpenBao ==="
echo ""

# Count orgs in each
VAULT_ORGS=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" --request LIST "$VAULT_ADDR/v1/kv2/metadata/" | jq -r '.data.keys | length')
OPENBAO_ORGS=$(curl -s -H "Authorization: Bearer $TOKEN" --request LIST "$OPENBAO_ADDR/v1/kv2/metadata/" | jq -r '.data.keys | length')

echo "Organizations: Vault=$VAULT_ORGS, OpenBao=$OPENBAO_ORGS"

# Check deleted orgs (1-20 should be gone from Vault but still in OpenBao)
echo ""
echo "=== Check deleted organizations (1-20) ==="
echo "Vault org-1:"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" --request LIST "$VAULT_ADDR/v1/kv2/metadata/org-1/" | jq '.data.keys // "NOT FOUND"'
echo "OpenBao org-1:"
curl -s -H "Authorization: Bearer $TOKEN" --request LIST "$OPENBAO_ADDR/v1/kv2/metadata/org-1/" | jq '.data.keys // "NOT FOUND"'

# Check deleted root secrets (21-40)
echo ""
echo "=== Check deleted root secrets (orgs 21-40) ==="
echo "Vault org-21 root-key:"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-21/root-key" | jq '.data.data // "NOT FOUND"'
echo "OpenBao org-21 root-key:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-21/root-key" | jq '.data.data // "NOT FOUND"'

# Check deleted nested secrets (41-65)
echo ""
echo "=== Check deleted nested secrets (orgs 41-65) ==="
echo "Vault org-41 production/db-password:"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-41/production/db-password" | jq '.data.data // "NOT FOUND"'
echo "OpenBao org-41 production/db-password:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-41/production/db-password" | jq '.data.data // "NOT FOUND"'

# Check updated secrets
echo ""
echo "=== Check UPDATED secrets ==="
echo "Vault org-66 root-key (should be UPDATED):"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-66/root-key" | jq '.data.data'
echo "OpenBao org-66 root-key:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-66/root-key" | jq '.data.data'

echo ""
echo "Vault org-71 production/db-password (should be UPDATED):"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-71/production/db-password" | jq '.data.data'
echo "OpenBao org-71 production/db-password:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-71/production/db-password" | jq '.data.data'

echo ""
echo "Vault org-76 staging/db-password (should be UPDATED):"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-76/staging/db-password" | jq '.data.data'
echo "OpenBao org-76 staging/db-password:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-76/staging/db-password" | jq '.data.data'

# Check still existing secrets
echo ""
echo "=== Check untouched secrets ==="
echo "Vault org-50 root-key:"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/kv2/data/org-50/root-key" | jq '.data.data'
echo "OpenBao org-50 root-key:"
curl -s -H "Authorization: Bearer $TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-50/root-key" | jq '.data.data'

echo ""
echo "=== TEST COMPLETE ==="