#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-vault-token}"
OPENBAO_ADDR="${OPENBAO_ADDR:-http://127.0.0.1:8201}"
OPENBAO_TOKEN="${OPENBAO_TOKEN:-openbao-token}"
KV_MOUNT="${KV_MOUNT:-kv2}"
PLUGIN_PATH="${PROJECT_DIR}/dist/replicator"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

wait_for_service() {
    local addr="$1"
    local name="$2"
    local max_attempts=30
    local attempt=1
    
    log_info "Waiting for $name to be ready at $addr..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -o /dev/null -w "%{http_code}" "$addr/v1/sys/health" 2>/dev/null | grep -q "200\|429"; then
            log_info "$name is ready!"
            return 0
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "$name failed to start after $max_attempts attempts"
    return 1
}

build_plugin() {
    log_info "Building plugin..."
    cd "$PROJECT_DIR"
    make build
    log_info "Plugin built at $PLUGIN_PATH"
}

register_plugin() {
    log_info "Registering plugin in OpenBao..."
    
    local plugin_sha
    plugin_sha=$(sha256sum "$PLUGIN_PATH" | cut -d' ' -f1)
    
    $OPENBAO_BIN write sys/plugins/catalog/vault-replicator \
        sha_256="$plugin_sha" \
        command="replicator" 2>/dev/null || {
        log_warn "Plugin already registered, updating..."
        $OPENBAO_BIN write sys/plugins/catalog/vault-replicator \
            sha_256="$plugin_sha" \
            command="replicator"
    }
}

enable_plugin() {
    log_info "Enabling replicator plugin..."
    
    $OPENBAO_BIN secrets enable -path=replicator -plugin-name=vault-replicator plugin 2>/dev/null || {
        log_warn "Plugin already enabled"
    }
}

get_approle_credentials() {
    log_info "Getting AppRole credentials from Vault..."
    
    export VAULT_TOKEN
    
    ROLE_ID=$(vault read -field=role_id auth/approle/role/replicator/role-id)
    SECRET_ID=$(vault write -field=secret_id -f auth/approle/role/replicator/secret-id)
    
    echo "ROLE_ID=$ROLE_ID"
    echo "SECRET_ID=$SECRET_ID"
}

configure_plugin() {
    log_info "Configuring replicator plugin..."
    
    $OPENBAO_BIN write replicator/config \
        vault_address="$VAULT_ADDR" \
        vault_mount="$KV_MOUNT" \
        approle_role_id="$ROLE_ID" \
        approle_secret_id="$SECRET_ID" \
        destination_token="$OPENBAO_TOKEN" \
        destination_mount="$KV_MOUNT" \
        organization_path=""
    
    log_info "Plugin configured successfully"
}

trigger_sync() {
    log_info "Triggering secret sync..."
    
    $OPENBAO_BIN write -f replicator/sync/secrets
}

verify_sync() {
    log_info "Verifying sync results..."
    
    echo ""
    echo "=== Sync Status ==="
    $OPENBAO_BIN read replicator/sync/status
    
    echo ""
    echo "=== Checking replicated secrets in OpenBao ==="
    
    local all_verified=true
    
    for org in org-1 org-2 org-3 org-4 org-5; do
        echo ""
        echo "--- Verifying $org ---"
        
        if $OPENBAO_BIN kv list "$KV_MOUNT/metadata/replicator/$org" 2>/dev/null; then
            local secrets_count=$($OPENBAO_BIN kv list "$KV_MOUNT/metadata/replicator/$org" 2>/dev/null | grep -c "^[a-zA-Z]" || true)
            log_info "Found $secrets_count secrets in $org"
            all_verified=false
        fi
    done
    
    if [ "$all_verified" = true ]; then
        log_info "Checking specific secrets..."
        
        for org in org-1 org-2 org-3 org-4 org-5; do
            case $org in
                org-1)
                    check_secret "$KV_MOUNT/data/replicator/$org/api-key" "key_api-key"
                    check_secret "$KV_MOUNT/data/replicator/$org/database-password" "key_database-password"
                    ;;
                org-2)
                    check_secret "$KV_MOUNT/data/replicator/$org/aws-secret" "key_aws-secret"
                    ;;
                org-3)
                    check_secret "$KV_MOUNT/data/replicator/$org/gcp-credentials" "key_gcp-credentials"
                    ;;
                org-4)
                    check_secret "$KV_MOUNT/data/replicator/$org/production/database" "folder_key"
                    check_secret "$KV_MOUNT/data/replicator/$org/production/cache" "folder_key"
                    ;;
                org-5)
                    check_secret "$KV_MOUNT/data/replicator/$org/app-secret" "key_app-secret"
                    ;;
            esac
        done
    fi
    
    echo ""
    log_info "Sync verification complete"
}

check_secret() {
    local path="$1"
    local key="$2"
    
    if $OPENBAO_BIN kv get "$path" 2>/dev/null; then
        log_info "✓ Secret found: $path"
    else
        log_error "✗ Secret NOT found: $path"
    fi
}

test_dry_run() {
    log_info "Testing dry run mode..."
    
    $OPENBAO_BIN write replicator/sync/secrets dry_run=true
    
    log_info "Dry run completed (no secrets should be written)"
}

cleanup() {
    log_info "Cleaning up..."
    
    $OPENBAO_BIN secrets disable replicator 2>/dev/null || true
    $OPENBAO_BIN delete sys/plugins/catalog/vault-replicator 2>/dev/null || true
}

OPENBAO_BIN="bao -address=$OPENBAO_ADDR -token=$OPENBAO_TOKEN"
export VAULT_TOKEN

main() {
    log_info "=== Starting Integration Tests ==="
    log_info "Vault: $VAULT_ADDR"
    log_info "OpenBao: $OPENBAO_ADDR"
    log_info "Plugin: $PLUGIN_PATH"
    
    wait_for_service "$VAULT_ADDR" "Vault"
    wait_for_service "$OPENBAO_ADDR" "OpenBao"
    
    build_plugin
    
    get_approle_credentials
    
    register_plugin
    enable_plugin
    configure_plugin
    
    echo ""
    log_info "=== Running Sync Tests ==="
    
    echo ""
    echo "--- Test 1: Full Sync ---"
    trigger_sync
    verify_sync
    
    echo ""
    echo "--- Test 2: Dry Run ---"
    test_dry_run
    
    echo ""
    log_info "=== All Integration Tests Complete ==="
}

main "$@"
