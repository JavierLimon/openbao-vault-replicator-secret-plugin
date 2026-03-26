#!/bin/bash
set -e

echo "=== Full Integration Test Script ==="
echo ""

VAULT_ADDR="http://127.0.0.1:8200"
VAULT_TOKEN="vault-token"
OPENBAO_ADDR="http://127.0.0.1:8201"
KV_MOUNT="kv2"
PROJECT_DIR="/Users/javierlimon/Documents/git/openbao-vault-replicator-secret-plugin"

cd "$PROJECT_DIR"

echo "=== 1. Starting containers ==="
docker-compose up -d
sleep 15

echo ""
echo "=== 2. Setup OpenBao (check existing or init) ==="
SEAL_STATUS=$(curl -s "$OPENBAO_ADDR/v1/sys/seal-status")
INITIALIZED=$(echo $SEAL_STATUS | jq -r '.initialized')
SEALED=$(echo $SEAL_STATUS | jq -r '.sealed')

if [ "$INITIALIZED" = "false" ]; then
    INIT_RESP=$(curl -s -X POST "$OPENBAO_ADDR/v1/sys/init" -d '{"secret_shares": 1, "secret_threshold": 1}')
    OPENBAO_TOKEN=$(echo $INIT_RESP | jq -r '.root_token')
    OPENBAO_KEY=$(echo $INIT_RESP | jq -r '.keys[0]')
    curl -s -X POST "$OPENBAO_ADDR/v1/sys/unseal" -d "{\"key\": \"$OPENBAO_KEY\"}" > /dev/null
    echo "Initialized new OpenBao: $OPENBAO_TOKEN"
else
    if [ "$SEALED" = "true" ]; then
        curl -s -X POST "$OPENBAO_ADDR/v1/sys/unseal" -d '{"key": "e764bcd7ee9de19b74760ebb39ac94c702fd72c9a9155283a2c51e1006e7430a"}' > /dev/null
    fi
    OPENBAO_TOKEN="s.zr7Jf2ZqVZSb865HAPhlzd2"
    echo "Using existing OpenBao token: $OPENBAO_TOKEN"
fi

echo ""
echo "=== 3. Setting up Vault ==="
curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/sys/mounts/$KV_MOUNT" -d '{"type": "kv", "options": {"version": "2"}}' 2>/dev/null || true

curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/sys/auth/approle" -d '{"type": "approle"}' 2>/dev/null || true

curl -s -X PUT -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/sys/policies/acl/replicator" \
    -d '{"policy": "path \"kv2/data/*\" { capabilities = [\"read\", \"list\"] } path \"kv2/metadata/*\" { capabilities = [\"read\", \"list\"] }"}' 2>/dev/null || true

curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/auth/approle/role/replicator" -d '{"policies": ["replicator"], "token_ttl": "1h"}' 2>/dev/null || true

ROLE_ID=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/auth/approle/role/replicator/role-id" | jq -r '.data.role_id')
SECRET_ID=$(curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/auth/approle/role/replicator/secret-id" | jq -r '.data.secret_id')

echo "AppRole: $ROLE_ID"

echo ""
echo "=== 4. Setting up OpenBao mounts ==="
curl -s -X POST -H "Authorization: Bearer $OPENBAO_TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/sys/mounts/$KV_MOUNT" -d '{"type": "kv", "options": {"version": "2"}}' 2>/dev/null || true

echo ""
echo "=== 5. Building and registering plugin ==="
GOOS=linux GOARCH=amd64 go build -o dist/replicator ./cmd/vault-replicator/
docker cp dist/replicator openbao-dest:/vault/plugins/vault-openbao-replicator 2>/dev/null || true

SHA256=$(sha256sum dist/replicator | cut -d' ' -f1)
curl -s -X POST -H "Authorization: Bearer $OPENBAO_TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/sys/plugins/catalog/vault-replicator" \
    -d "{\"sha256\": \"$SHA256\", \"command\": \"vault-openbao-replicator\"}" 2>/dev/null || true

curl -s -X POST -H "Authorization: Bearer $OPENBAO_TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/sys/mounts/replicator" -d '{"type": "plugin", "plugin_name": "vault-replicator"}' 2>/dev/null || true

echo ""
echo "=== 6. Configuring plugin ==="
curl -s -X POST -H "Authorization: Bearer $OPENBAO_TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/replicator/config" \
    -d "{
        \"vault_address\": \"http://vault:8200\",
        \"vault_mount\": \"kv2\",
        \"approle_role_id\": \"$ROLE_ID\",
        \"approle_secret_id\": \"$SECRET_ID\",
        \"destination_token\": \"$OPENBAO_TOKEN\",
        \"destination_mount\": \"kv2\",
        \"organization_path\": \".\"
    }" 2>/dev/null || true

echo "Plugin configured"

echo ""
echo "=== 7. Populating Vault with 100 organizations ==="
for i in $(seq 1 100); do
    org="org-$i"
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/root-key" \
        -d "{\"data\": {\"key\": \"root-value-$i\", \"org\": \"$org\"}}" > /dev/null 2>&1
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/production/db-password" \
        -d "{\"data\": {\"password\": \"prod-pass-$i\"}}" > /dev/null 2>&1
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/production/api-token" \
        -d "{\"data\": {\"token\": \"prod-token-$i\"}}" > /dev/null 2>&1
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
        "$VAULT_ADDR/v1/kv2/data/$org/staging/db-password" \
        -d "{\"data\": {\"password\": \"stage-pass-$i\"}}" > /dev/null 2>&1
    
    if [ $((i % 20)) -eq 0 ]; then
        echo "  Created $i orgs..."
    fi
done
echo "  Done: 100 orgs created"

echo ""
echo "=== 8. Running sync ==="
SYNC_RESP=$(curl -s -X POST -H "Authorization: Bearer $OPENBAO_TOKEN" -H "Content-Type: application/json" \
    "$OPENBAO_ADDR/v1/replicator/sync/secrets" -d '{"dry_run": false}')

ORG_SYNCED=$(echo $SYNC_RESP | jq -r '.data.organizations_synced // 0')
SECRET_SYNCED=$(echo $SYNC_RESP | jq -r '.data.secrets_synced // 0')
STATUS=$(echo $SYNC_RESP | jq -r '.data.status // .errors[0]')

echo "Sync complete: $ORG_SYNCED orgs, $SECRET_SYNCED secrets, status: $STATUS"

echo ""
echo "=== 9. Verification ==="
DEST_ORGS=$(curl -s -H "Authorization: Bearer $OPENBAO_TOKEN" --request LIST "$OPENBAO_ADDR/v1/kv2/metadata/" | jq -r '.data.keys | length // 0')
echo "Organizations in OpenBao: $DEST_ORGS"

echo ""
echo "Secrets in org-50:"
curl -s -H "Authorization: Bearer $OPENBAO_TOKEN" --request LIST "$OPENBAO_ADDR/v1/kv2/metadata/org-50/" | jq '.data.keys // []'

echo ""
echo "Secrets in org-100:"
curl -s -H "Authorization: Bearer $OPENBAO_TOKEN" --request LIST "$OPENBAO_ADDR/v1/kv2/metadata/org-100/" | jq '.data.keys // []'

echo ""
echo "Sample secret from org-1:"
curl -s -H "Authorization: Bearer $OPENBAO_TOKEN" "$OPENBAO_ADDR/v1/kv2/data/org-1/root-key" | jq '.data.data // empty'

echo ""
echo "=== TEST COMPLETE ==="