#!/bin/bash
set -e

VAULT_ADDR="${VAULT_ADDR:-http://127.00.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-vault-token}"
KV_MOUNT="${KV_MOUNT:-kv2}"
NUM_ORGS="${NUM_ORGS:-100}"
MAX_DEPTH="${MAX_DEPTH:-4}"
MAX_SECRETS_PER_FOLDER="${MAX_SECRETS_PER_FOLDER:-5}"

export VAULT_ADDR VAULT_TOKEN

echo "=== Extensive Vault Population Test ==="
echo "Organizations: $NUM_ORGS"
echo "Max depth: $MAX_DEPTH"
echo "Max secrets per folder: $MAX_SECRETS_PER_FOLDER"

# Ensure KVv2 is enabled
echo "Ensuring KVv2 is enabled at $KV_MOUNT..."
curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/sys/mounts/$KV_MOUNT" \
    -d '{"type": "kv", "options": {"version": "2"}}' 2>/dev/null || true

# Ensure AppRole is enabled
echo "Ensuring AppRole is enabled..."
curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
    "$VAULT_ADDR/v1/sys/auth/approle" \
    -d '{"type": "approle"}' 2>/dev/null || true

# Function to generate random secret data
generate_secret_data() {
    local secret_name="$1"
    local org="$2"
    local path="$3"
    
    # Generate random key-value pairs
    local num_fields=$((RANDOM % 5 + 3))  # 3-7 fields
    
    local data=""
    for i in $(seq 1 $num_fields); do
        local key="field_$i"
        local value="val_${org}_${path}_${secret_name}_$RANDOM"
        data="$data\"$key\": \"$value\", "
    done
    
    # Add some metadata
    data="$data\"created_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\", "
    data="$data\"environment\": \"$([ \"$((RANDOM % 2))\" -eq 0 ] && echo \"production\" || echo \"staging\")\", "
    data="$data\"version\": \"1.0.$RANDOM\""
    
    echo "{$data}"
}

# Function to create random path structure
create_random_paths() {
    local org="$1"
    local current_path=""
    local depth=$((RANDOM % MAX_DEPTH + 1))
    
    for d in $(seq 1 $depth); do
        local folder_name="layer_$((RANDOM % 10))_$([ "$((RANDOM % 3))" -eq 0 ] && echo "prod" || echo "$([ "$((RANDOM % 3))" -eq 0 ] && echo "dev" || echo "shared")")"
        
        if [ -z "$current_path" ]; then
            current_path="$folder_name"
        else
            current_path="$current_path/$folder_name"
        fi
        
        # Create secrets at this level
        local num_secrets=$((RANDOM % MAX_SECRETS_PER_FOLDER + 1))
        
        for s in $(seq 1 $num_secrets); do
            local secret_name="secret_$([ "$((RANDOM % 3))" -eq 0 ] && echo "key" || echo "$([ "$((RANDOM % 3))" -eq 0 ] && echo "token" || echo "cert")")_$RANDOM"
            local secret_path="$KV_MOUNT/data/$org/$current_path/$secret_name"
            local secret_data=$(generate_secret_data "$secret_name" "$org" "$current_path")
            
            # Replace spaces in path for curl
            secret_path_encoded=$(echo "$secret_path" | sed 's/ /%20/g')
            
            curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
                "$VAULT_ADDR/v1/$secret_path_encoded" \
                -d "{\"data\": $secret_data}" > /dev/null 2>&1
            
            if [ $? -eq 0 ]; then
                echo "  Created: $org/$current_path/$secret_name"
            fi
        done
    done
}

# Create root-level secrets for each org (without subfolder)
create_root_secrets() {
    local org="$1"
    local num_root_secrets=$((RANDOM % 3 + 1))  # 1-3 root secrets
    
    for s in $(seq 1 $num_root_secrets); do
        local secret_name="root_secret_$([ "$((RANDOM % 4))" -eq 0 ] && echo "api" || echo "$([ "$((RANDOM % 4))" -eq 0 ] && echo "db" || echo "$([ "$((RANDOM % 4))" -eq 0 ] && echo "aws" || echo "jwt")")")_$RANDOM"
        local secret_path="$KV_MOUNT/data/$org/$secret_name"
        local secret_data=$(generate_secret_data "$secret_name" "$org" "root")
        
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" -H "Content-Type: application/json" \
            "$VAULT_ADDR/v1/$secret_path" \
            -d "{\"data\": $secret_data}" > /dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            echo "  Created root: $org/$secret_name"
        fi
    done
}

echo ""
echo "=== Creating $NUM_ORGS organizations with random paths ==="

total_start=$(date +%s)
created_orgs=0
created_secrets=0

for i in $(seq 1 $NUM_ORGS); do
    org="org-$i"
    
    # Create root-level secrets
    create_root_secrets "$org"
    created_secrets=$((created_secrets + $?))
    
    # Create nested path structure
    create_random_paths "$org"
    
    created_orgs=$((created_orgs + 1))
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "Progress: $i/$NUM_ORGS orgs..."
    fi
done

total_end=$(date +%s)
duration=$((total_end - total_start))

echo ""
echo "=== Population Complete ==="
echo "Organizations: $created_orgs"
echo "Time: ${duration}s"

# Verify by listing some orgs
echo ""
echo "=== Verification ==="
echo "Sample organizations (first 10):"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" --request LIST "$VAULT_ADDR/v1/$KV_MOUNT/metadata/" | jq -r '.data.keys[:10] // []' 2>/dev/null || echo "Could not list"

echo ""
echo "Sample secrets in org-1:"
curl -s -H "X-Vault-Token: $VAULT_TOKEN" --request LIST "$VAULT_ADDR/v1/$KV_MOUNT/metadata/org-1/" | jq -r '.data.keys[:10] // []' 2>/dev/null || echo "Could not list secrets"

echo ""
echo "=== Population script complete ==="