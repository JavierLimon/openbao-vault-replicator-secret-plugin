#!/bin/bash
# Comprehensive Integration Test Plan for Vault Replicator Plugin
# Tests: 100-org sync, modifications, deletions, blacklist, dry-run, nested secrets

set -e

# Configuration
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-vault-token}"
OPENBAO_ADDR="${OPENBAO_ADDR:-http://127.0.0.1:8201}"
KV_MOUNT="${KV_MOUNT:-kv2}"
PROJECT_DIR="/Users/javierlimon/Documents/git/openbao-vault-replicator-secret-plugin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_test() { echo -e "\n=== TEST: $1 ===\n"; }

# Get OpenBao token from running container
get_openbao_token() {
    # Try to get from running container
    OPENBAO_TOKEN=$(podman exec openbao-dest cat /tmp/openbao-init.json 2>/dev/null | jq -r '.root_token' || echo "")
    if [ -z "$OPENBAO_TOKEN" ]; then
        log_error "Could not get OpenBao token"
        exit 1
    fi
    echo "$OPENBAO_TOKEN"
}

# Wait for services
wait_for_services() {
    log_info "Waiting for Vault..."
    for i in {1..30}; do
        if curl -s "$VAULT_ADDR/v1/sys/health" | jq -e '.initialized == true' > /dev/null 2>&1; then
            log_info "Vault is ready"
            break
        fi
        sleep 2
    done
    
    log_info "Waiting for OpenBao..."
    for i in {1..30}; do
        if curl -s "$OPENBAO_ADDR/v1/sys/health" | jq -e '.initialized == true and .sealed == false' > /dev/null 2>&1; then
            log_info "OpenBao is ready"
            return 0
        fi
        sleep 2
    done
    log_error "OpenBao not ready"
    return 1
}

# Setup Vault with AppRole
setup_vault() {
    log_info "Setting up Vault..."
    
    # Enable KVv2
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/sys/mounts/$KV_MOUNT" \
        -d '{"type":"kv","options":{"version":"2"}}' 2>/dev/null || true
    
    # Enable AppRole
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/sys/auth/approle" \
        -d '{"type":"approle"}' 2>/dev/null || true
    
    # Create policy
    curl -s -X PUT -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/sys/policies/acl/replicator" \
        -d '{"policy":"path \"kv2/data/*\" { capabilities = [\"read\", \"list\"] } path \"kv2/metadata/*\" { capabilities = [\"read\", \"list\"] } path \"kv2/metadata\" { capabilities = [\"list\"] }"}' 2>/dev/null || true
    
    # Create AppRole
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/auth/approle/role/replicator" \
        -d '{"policies":["replicator"],"token_ttl":"1h"}' 2>/dev/null || true
    
    # Get credentials
    ROLE_ID=$(curl -s -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/auth/approle/role/replicator/role-id" | jq -r '.data.role_id')
    
    SECRET_ID=$(curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/auth/approle/role/replicator/secret-id" | jq -r '.data.secret_id')
    
    echo "ROLE_ID=$ROLE_ID"
    echo "SECRET_ID=$SECRET_ID"
}

# Setup OpenBao
setup_openbao() {
    local TOKEN="$1"
    log_info "Setting up OpenBao..."
    
    # Enable KVv2
    curl -s -X POST -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/sys/mounts/$KV_MOUNT" \
        -d '{"type":"kv","options":{"version":"2"}}' 2>/dev/null || true
}

# Register and enable plugin
setup_plugin() {
    local TOKEN="$1"
    local ROLE_ID="$2"
    local SECRET_ID="$3"
    
    log_info "Building and registering plugin..."
    
    # Build for Linux
    cd "$PROJECT_DIR"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin.Version=1.0.0" -o dist/replicator ./cmd/vault-replicator/
    
    # Copy to container
    podman cp dist/replicator openbao-dest:/vault/plugins/replicator 2>/dev/null || true
    
    # Get SHA
    SHA256=$(podman exec openbao-dest sha256sum /vault/plugins/replicator | cut -d' ' -f1)
    
    # Register plugin
    curl -s -X PUT -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/sys/plugins/catalog/secret/replicator" \
        -d "{\"sha256\":\"$SHA256\",\"command\":\"replicator\"}" 2>/dev/null || true
    
    # Enable plugin
    curl -s -X POST -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/sys/mounts/replicator" \
        -d '{"type":"plugin","plugin_name":"replicator"}' 2>/dev/null || true
    
    # Configure plugin
    curl -s -X PUT -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/config" \
        -d "{
            \"vault_address\": \"http://vault-source:8200\",
            \"vault_mount\": \"$KV_MOUNT\",
            \"approle_role_id\": \"$ROLE_ID\",
            \"approle_secret_id\": \"$SECRET_ID\",
            \"destination_token\": \"$TOKEN\",
            \"destination_mount\": \"$KV_MOUNT\",
            \"organization_path\": \"\"
        }" 2>/dev/null || true
    
    log_info "Plugin configured"
}

# Create test secrets
create_test_secrets() {
    local NUM_ORGS="${1:-100}"
    log_info "Creating $NUM_ORGS organizations with secrets..."
    
    for i in $(seq 1 $NUM_ORGS); do
        org="org-$i"
        
        # Root secrets
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/root-key" \
            -d "{\"data\": {\"key\": \"root-value-$i\"}}" > /dev/null
        
        # Production secrets
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/production/db-password" \
            -d "{\"data\": {\"password\": \"prod-pass-$i\"}}" > /dev/null
        
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/production/api-token" \
            -d "{\"data\": {\"token\": \"prod-token-$i\"}}" > /dev/null
        
        # Staging secrets
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/staging/db-password" \
            -d "{\"data\": {\"password\": \"stage-pass-$i\"}}" > /dev/null
        
        if [ $((i % 20)) -eq 0 ]; then
            echo "  Created $i orgs..."
        fi
    done
    log_info "Created $NUM_ORGS orgs with secrets"
}

# Run sync
run_sync() {
    local TOKEN="$1"
    local DRY_RUN="${2:-false}"
    local ORG_FILTER="${3:-}"
    
    local BODY="{\"dry_run\":$DRY_RUN}"
    if [ -n "$ORG_FILTER" ]; then
        BODY="{\"dry_run\":$DRY_RUN,\"organizations\":[$ORG_FILTER]}"
    fi
    
    curl -s -X POST -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/sync/secrets" \
        -d "$BODY"
}

# Count secrets in destination
count_dest_secrets() {
    local TOKEN="$1"
    local org="$2"
    
    curl -s -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/$KV_MOUNT/metadata/$org" 2>/dev/null | jq '.data.keys | length' || echo "0"
}

# Get secret value
get_secret() {
    local TOKEN="$1"
    local path="$2"
    
    curl -s -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/$KV_MOUNT/data/$path" | jq -r '.data.data // empty'
}

# =============================================================================
# TEST CASES
# =============================================================================

test_initial_sync() {
    local TOKEN="$1"
    
    log_test "INITIAL SYNC (100 orgs, 400 secrets)"
    
    run_sync "$TOKEN" | jq '.data'
    
    local SYNCED=$(curl -s -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/sync/status" | jq -r '.data.secrets_synced')
    local FAILED=$(curl -s -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/sync/status" | jq -r '.data.failed')
    
    echo "Secrets synced: $SYNCED"
    echo "Failed: $FAILED"
    
    if [ "$SYNCED" -ge 400 ]; then
        log_info "PASS: Initial sync successful"
        return 0
    else
        log_error "FAIL: Expected 400+ secrets synced, got $SYNCED"
        return 1
    fi
}

test_modifications() {
    local TOKEN="$1"
    
    log_test "MODIFICATION SYNC"
    
    # Update secrets in 10 orgs
    for i in $(seq 1 10); do
        org="org-$i"
        curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/root-key" \
            -d "{\"data\": {\"key\": \"MODIFIED-$i\"}}" > /dev/null
    done
    
    # Sync
    run_sync "$TOKEN" | jq '.data'
    
    # Verify modifications
    local all_correct=true
    for i in $(seq 1 10); do
        org="org-$i"
        value=$(get_secret "$TOKEN" "$org/root-key" | jq -r '.key')
        if [[ "$value" != "MODIFIED-$i" ]]; then
            log_error "FAIL: org-$i root-key not updated: $value"
            all_correct=false
        fi
    done
    
    if $all_correct; then
        log_info "PASS: All modifications synced"
        return 0
    else
        return 1
    fi
}

test_nested_secrets() {
    local TOKEN="$1"
    
    log_test "NESTED SECRETS SYNC"
    
    # Verify nested secrets were synced (org-1/production/db-password)
    local nested=$(get_secret "$TOKEN" "org-1/production/db-password" | jq -r '.password')
    
    if [[ "$nested" == "prod-pass-1" ]]; then
        log_info "PASS: Nested secrets synced correctly"
        return 0
    else
        log_error "FAIL: Nested secret not synced correctly: $nested"
        return 1
    fi
}

test_deletion_sync() {
    local TOKEN="$1"
    
    log_test "DELETION SYNC"
    
    # Delete 10 secrets from Vault
    for i in $(seq 91 100); do
        org="org-$i"
        curl -s -X DELETE -H "X-Vault-Token: $VAULT_TOKEN" \
            "$VAULT_ADDR/v1/$KV_MOUNT/data/$org/production/api-token" > /dev/null
    done
    
    # Enable deletion sync in config
    curl -s -X PUT -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/config" \
        -d "{
            \"vault_address\": \"http://vault-source:8200\",
            \"vault_mount\": \"$KV_MOUNT\",
            \"approle_role_id\": \"$(curl -s -H \"X-Vault-Token: $VAULT_TOKEN\" \"$VAULT_ADDR/v1/auth/approle/role/replicator/role-id\" | jq -r '.data.role_id')\",
            \"approle_secret_id\": \"$(curl -s -X POST -H \"X-Vault-Token: $VAULT_TOKEN\" \"$VAULT_ADDR/v1/auth/approle/role/replicator/secret-id\" | jq -r '.data.secret_id')\",
            \"destination_token\": \"$TOKEN\",
            \"destination_mount\": \"$KV_MOUNT\",
            \"organization_path\": \"\",
            \"allow_deletion_sync\": true
        }" > /dev/null
    
    # Sync with deletion
    run_sync "$TOKEN" | jq '.data'
    
    # Verify deletions - these secrets should NOT exist in destination
    local still_exists=0
    for i in $(seq 91 100); do
        org="org-$i"
        if curl -s -H "X-Vault-Token: $TOKEN" \
            "$OPENBAO_ADDR/v1/$KV_MOUNT/data/$org/production/api-token" | jq -e '.data' > /dev/null 2>&1; then
            ((still_exists++))
        fi
    done
    
    if [ "$still_exists" -eq 0 ]; then
        log_info "PASS: Deleted secrets removed from destination"
        return 0
    else
        log_error "FAIL: $still_exists deleted secrets still exist in destination"
        return 1
    fi
}

test_blacklist() {
    local TOKEN="$1"
    
    log_test "ORG BLACKLIST"
    
    # Add org-1 and org-2 to skip list
    curl -s -X PUT -H "X-Vault-Token: $TOKEN" \
        "$OPENBAO_ADDR/v1/replicator/config" \
        -d "{
            \"vault_address\": \"http://vault-source:8200\",
            \"vault_mount\": \"$KV_MOUNT\",
            \"approle_role_id\": \"$(curl -s -H \"X-Vault-Token: $VAULT_TOKEN\" \"$VAULT_ADDR/v1/auth/approle/role/replicator/role-id\" | jq -r '.data.role_id')\",
            \"approle_secret_id\": \"$(curl -s -X POST -H \"X-Vault-Token: $VAULT_TOKEN\" \"$VAULT_ADDR/v1/auth/approle/role/replicator/secret-id\" | jq -r '.data.secret_id')\",
            \"destination_token\": \"$TOKEN\",
            \"destination_mount\": \"$KV_MOUNT\",
            \"organization_path\": \"\",
            \"org_skip_list\": [\"org-1\", \"org-2\"]
        }" > /dev/null
    
    # Verify org-1 and org-2 are NOT in destination
    local count=0
    for org in org-1 org-2; do
        if curl -s -H "X-Vault-Token: $TOKEN" \
            "$OPENBAO_ADDR/v1/$KV_MOUNT/metadata/$org" 2>/dev/null | jq -e '.data.keys | length > 0' > /dev/null 2>&1; then
            ((count++))
        fi
    done
    
    if [ "$count" -eq 0 ]; then
        log_info "PASS: Blacklisted orgs not synced"
        return 0
    else
        log_error "FAIL: Blacklisted orgs still exist in destination"
        return 1
    fi
}

test_dry_run() {
    local TOKEN="$1"
    
    log_test "DRY RUN MODE"
    
    # Make a change
    curl -s -X POST -H "X-Vault-Token: $VAULT_TOKEN" \
        "$VAULT_ADDR/v1/$KV_MOUNT/data/org-50/root-key" \
        -d "{\"data\": {\"key\": \"DRYRUN-TEST\"}}" > /dev/null
    
    # Get current value
    local before=$(get_secret "$TOKEN" "org-50/root-key" | jq -r '.key')
    
    # Dry run sync
    run_sync "$TOKEN" true | jq '.data'
    
    # Value should NOT have changed
    local after=$(get_secret "$TOKEN" "org-50/root-key" | jq -r '.key')
    
    if [[ "$before" == "$after" && "$after" != "DRYRUN-TEST" ]]; then
        log_info "PASS: Dry run did not modify secrets"
        return 0
    else
        log_error "FAIL: Dry run modified secrets (before=$before, after=$after)"
        return 1
    fi
}

test_specific_org_sync() {
    local TOKEN="$1"
    
    log_test "SPECIFIC ORG SYNC (org-50 only)"
    
    # Sync only org-50
    run_sync "$TOKEN" false '"org-50"' | jq '.data'
    
    # Verify org-50 has secrets, others don't
    local org50_count=$(count_dest_secrets "$TOKEN" "org-50")
    local org51_count=$(count_dest_secrets "$TOKEN" "org-51")
    
    if [ "$org50_count" -gt 0 ] && [ "$org51_count" -eq 0 ]; then
        log_info "PASS: Specific org sync works"
        return 0
    else
        log_error "FAIL: org-50=$org50_count, org-51=$org51_count"
        return 1
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    echo "=========================================="
    echo "  VAULT REPLICATOR COMPREHENSIVE TEST"
    echo "=========================================="
    
    # Get OpenBao token
    OPENBAO_TOKEN=$(get_openbao_token)
    log_info "OpenBao token: ${OPENBAO_TOKEN:0:20}..."
    
    # Wait for services
    wait_for_services
    
    # Setup
    CREDS=$(setup_vault)
    ROLE_ID=$(echo "$CREDS" | grep ROLE_ID | cut -d= -f2)
    SECRET_ID=$(echo "$CREDS" | grep SECRET_ID | cut -d= -f2)
    
    setup_openbao "$OPENBAO_TOKEN"
    setup_plugin "$OPENBAO_TOKEN" "$ROLE_ID" "$SECRET_ID"
    
    # Create 100 orgs with secrets
    create_test_secrets 100
    
    # Run tests
    RESULTS=()
    
    test_initial_sync "$OPENBAO_TOKEN" && RESULTS+=("initial_sync:PASS") || RESULTS+=("initial_sync:FAIL")
    test_nested_secrets "$OPENBAO_TOKEN" && RESULTS+=("nested:PASS") || RESULTS+=("nested:FAIL")
    test_modifications "$OPENBAO_TOKEN" && RESULTS+=("modifications:PASS") || RESULTS+=("modifications:FAIL")
    test_blacklist "$OPENBAO_TOKEN" && RESULTS+=("blacklist:PASS") || RESULTS+=("blacklist:FAIL")
    test_dry_run "$OPENBAO_TOKEN" && RESULTS+=("dry_run:PASS") || RESULTS+=("dry_run:FAIL")
    test_specific_org_sync "$OPENBAO_TOKEN" && RESULTS+=("specific_org:PASS") || RESULTS+=("specific_org:FAIL")
    test_deletion_sync "$OPENBAO_TOKEN" && RESULTS+=("deletion:PASS") || RESULTS+=("deletion:FAIL")
    
    # Summary
    echo ""
    echo "=========================================="
    echo "  TEST SUMMARY"
    echo "=========================================="
    for result in "${RESULTS[@]}"; do
        if [[ "$result" == *":PASS" ]]; then
            log_info "$result"
        else
            log_error "$result"
        fi
    done
    
    local failed=0
    for result in "${RESULTS[@]}"; do
        if [[ "$result" == *":FAIL" ]]; then
            ((failed++))
        fi
    done
    
    echo ""
    if [ $failed -eq 0 ]; then
        log_info "ALL TESTS PASSED!"
        return 0
    else
        log_error "$failed TESTS FAILED"
        return 1
    fi
}

main "$@"
