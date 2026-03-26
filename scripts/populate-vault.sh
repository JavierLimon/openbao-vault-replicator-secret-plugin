#!/bin/bash
set -e

VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-vault-token}"
KV_MOUNT="${KV_MOUNT:-kv2}"

export VAULT_ADDR VAULT_TOKEN

echo "=== Populating Vault with test secrets ==="
echo "Vault Address: $VAULT_ADDR"

enable_kv() {
    echo "Enabling KVv2 at $KV_MOUNT..."
    vault secrets enable -path="$KV_MOUNT" -version=2 kv 2>/dev/null || echo "KV mount already exists"
}

create_approle() {
    echo "Creating AppRole..."
    vault auth enable approle 2>/dev/null || echo "AppRole already enabled"
    
    vault write auth/approle/role/replicator \
        secret_id_ttl=1h \
        token_ttl=1h \
        token_max_ttl=4h \
        policies=default 2>/dev/null || echo "Role may already exist"
    
    ROLE_ID=$(vault read -field=role_id auth/approle/role/replicator)
    SECRET_ID=$(vault write -field=secret_id -f auth/approle/role/replicator/secret-id)
    
    echo "ROLE_ID=$ROLE_ID"
    echo "SECRET_ID=$SECRET_ID"
}

populate_org() {
    local org_name="$1"
    shift
    local secrets=("$@")
    
    for secret in "${secrets[@]}"; do
        local path="$KV_MOUNT/data/$org_name/$secret"
        echo "Writing secret: $path"
        vault kv put "$path" \
            key_"$secret"="value_for_${secret}_in_${org_name}" \
            description="Test secret for $org_name/$secret" \
            created="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    done
}

populate_org_with_folders() {
    local org_name="$1"
    shift
    local folders=("$@")
    
    for folder in "${folders[@]}"; do
        local path="$KV_MOUNT/data/$org_name/$folder"
        echo "Writing folder secrets: $path"
        vault kv put "$path" \
            folder_key="value_in_${org_name}_${folder}" \
            description="Test secrets in folder $folder" \
            created="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    done
}

echo ""
echo "=== Creating test organizations ==="

populate_org "org-1" "api-key" "database-password" "jwt-secret"
populate_org "org-2" "aws-secret" "azure-key"
populate_org "org-3" "gcp-credentials" "private-key"

populate_org_with_folders "org-4" "production/database" "production/cache" "staging/database"

populate_org "org-5" "app-secret" "encryption-key" "webhook-url"

echo ""
echo "=== Verifying secrets ==="
vault kv list "$KV_MOUNT/metadata/"

echo ""
echo "=== Listing secrets per org ==="
for org in org-1 org-2 org-3 org-4 org-5; do
    echo "--- $org ---"
    vault kv list "$KV_MOUNT/metadata/$org/" 2>/dev/null || vault kv list "$KV_MOUNT/metadata/$org"
done

echo ""
echo "=== Test data populated successfully ==="
echo "KV Mount: $KV_MOUNT"
echo ""
echo "AppRole Role ID:"
vault read -field=role_id auth/approle/role/replicator
