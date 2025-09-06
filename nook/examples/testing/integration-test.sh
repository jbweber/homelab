#!/bin/bash

# Nook Integration Test Script
# This script validates that Nook service is working correctly

set -e

NOOK_URL="${NOOK_URL:-http://localhost:8080}"
TEST_PREFIX="test-$(date +%s)"
FAILED_TESTS=0
TOTAL_TESTS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test counter
test_count() {
    ((TOTAL_TESTS++))
}

# Wait for service to be ready
wait_for_service() {
    local retries=30
    local count=0
    
    log "Waiting for Nook service at $NOOK_URL..."
    
    while [ $count -lt $retries ]; do
        if curl -s --fail "$NOOK_URL/api/v0/machines" > /dev/null 2>&1; then
            success "Nook service is ready"
            return 0
        fi
        
        sleep 1
        ((count++))
        echo -n "."
    done
    
    fail "Nook service did not become ready within $retries seconds"
    return 1
}

# Test API endpoints
test_api_endpoints() {
    log "Testing API endpoints..."
    
    # Test machines endpoint
    test_count
    if curl -s --fail "$NOOK_URL/api/v0/machines" > /dev/null; then
        success "Machines API endpoint accessible"
    else
        fail "Machines API endpoint not accessible"
    fi
    
    # Test networks endpoint
    test_count
    if curl -s --fail "$NOOK_URL/api/v0/networks" > /dev/null; then
        success "Networks API endpoint accessible"
    else
        fail "Networks API endpoint not accessible"
    fi
    
    # Test SSH keys endpoint
    test_count
    if curl -s --fail "$NOOK_URL/api/v0/ssh-keys" > /dev/null; then
        success "SSH Keys API endpoint accessible"
    else
        fail "SSH Keys API endpoint not accessible"
    fi
    
    # Test NoCloud metadata endpoints
    test_count
    if curl -s --fail "$NOOK_URL/latest/meta-data/" > /dev/null; then
        success "NoCloud metadata directory endpoint accessible"
    else
        fail "NoCloud metadata directory endpoint not accessible"
    fi
}

# Test CRUD operations
test_crud_operations() {
    log "Testing CRUD operations..."
    
    # Create network
    test_count
    local network_result
    network_result=$(curl -s -X POST "$NOOK_URL/api/v0/networks" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"${TEST_PREFIX}-net\",
            \"bridge\": \"br0\",
            \"subnet\": \"192.168.250.0/24\",
            \"gateway\": \"192.168.250.1\",
            \"description\": \"Test network\"
        }")
    
    if echo "$network_result" | jq -e '.id' > /dev/null; then
        success "Created test network"
        NETWORK_ID=$(echo "$network_result" | jq -r '.id')
    else
        fail "Failed to create test network"
        return 1
    fi
    
    # Create machine
    test_count
    local machine_result
    machine_result=$(curl -s -X POST "$NOOK_URL/api/v0/machines" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"${TEST_PREFIX}-vm\",
            \"hostname\": \"${TEST_PREFIX}-vm.local\",
            \"ipv4\": \"192.168.250.100\"
        }")
    
    if echo "$machine_result" | jq -e '.id' > /dev/null; then
        success "Created test machine"
        MACHINE_ID=$(echo "$machine_result" | jq -r '.id')
    else
        fail "Failed to create test machine"
        return 1
    fi
    
    # Create temporary SSH key for testing
    test_count
    local temp_key_file="/tmp/${TEST_PREFIX}-key.pub"
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test-key" > "$temp_key_file"
    
    local ssh_key_result
    ssh_key_result=$(curl -s -X POST "$NOOK_URL/api/v0/ssh-keys" \
        -H "Content-Type: application/json" \
        -d "{
            \"machine_id\": $MACHINE_ID,
            \"key_text\": \"$(cat "$temp_key_file")\"
        }")
    
    if echo "$ssh_key_result" | jq -e '.id' > /dev/null; then
        success "Created SSH key"
        SSH_KEY_ID=$(echo "$ssh_key_result" | jq -r '.id')
    else
        fail "Failed to create SSH key"
    fi
    
    rm -f "$temp_key_file"
    
    # Test machine retrieval by name
    test_count
    local machine_by_name
    machine_by_name=$(curl -s "$NOOK_URL/api/v0/machines/name/${TEST_PREFIX}-vm")
    
    if echo "$machine_by_name" | jq -e '.id' > /dev/null; then
        success "Retrieved machine by name"
    else
        fail "Failed to retrieve machine by name"
    fi
    
    # Test machine update
    test_count
    local update_result
    update_result=$(curl -s -X PATCH "$NOOK_URL/api/v0/machines/$MACHINE_ID" \
        -H "Content-Type: application/json" \
        -d "{
            \"hostname\": \"${TEST_PREFIX}-vm-updated.local\"
        }")
    
    if echo "$update_result" | jq -e '.hostname' | grep -q "updated"; then
        success "Updated machine hostname"
    else
        fail "Failed to update machine"
    fi
}

# Test NoCloud metadata functionality
test_nocloud_metadata() {
    log "Testing NoCloud metadata functionality..."
    
    # Test metadata directory
    test_count
    local meta_dir
    meta_dir=$(curl -s -H "X-Forwarded-For: 192.168.250.100" "$NOOK_URL/latest/meta-data/")
    
    if echo "$meta_dir" | grep -q "instance-id"; then
        success "NoCloud metadata directory contains expected entries"
    else
        fail "NoCloud metadata directory missing expected entries"
    fi
    
    # Test instance ID
    test_count
    local instance_id
    instance_id=$(curl -s -H "X-Forwarded-For: 192.168.250.100" "$NOOK_URL/latest/meta-data/instance-id")
    
    if [[ -n "$instance_id" && "$instance_id" != "null" ]]; then
        success "Instance ID returned: $instance_id"
    else
        fail "Instance ID not returned or null"
    fi
    
    # Test hostname
    test_count
    local hostname
    hostname=$(curl -s -H "X-Forwarded-For: 192.168.250.100" "$NOOK_URL/latest/meta-data/local-hostname")
    
    if echo "$hostname" | grep -q "${TEST_PREFIX}-vm"; then
        success "Local hostname returned: $hostname"
    else
        fail "Local hostname not returned correctly"
    fi
    
    # Test user data
    test_count
    local user_data
    user_data=$(curl -s -H "X-Forwarded-For: 192.168.250.100" "$NOOK_URL/latest/user-data")
    
    if echo "$user_data" | grep -q "cloud-config"; then
        success "User data returned with cloud-config format"
    else
        fail "User data not returned in expected format"
    fi
    
    # Test network config
    test_count
    local network_config
    network_config=$(curl -s -H "X-Forwarded-For: 192.168.250.100" "$NOOK_URL/latest/network-config")
    
    if echo "$network_config" | grep -q "version: 2"; then
        success "Network config returned in Netplan format"
    else
        fail "Network config not returned in expected format"
    fi
}

# Test error conditions
test_error_conditions() {
    log "Testing error conditions..."
    
    # Test 404 for non-existent machine
    test_count
    local status_code
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$NOOK_URL/api/v0/machines/99999")
    
    if [[ "$status_code" == "404" ]]; then
        success "Returns 404 for non-existent machine"
    else
        fail "Expected 404 for non-existent machine, got $status_code"
    fi
    
    # Test invalid JSON
    test_count
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$NOOK_URL/api/v0/machines" \
        -H "Content-Type: application/json" \
        -d "invalid json")
    
    if [[ "$status_code" == "400" ]]; then
        success "Returns 400 for invalid JSON"
    else
        fail "Expected 400 for invalid JSON, got $status_code"
    fi
    
    # Test metadata for non-existent machine
    test_count
    status_code=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "X-Forwarded-For: 1.2.3.4" "$NOOK_URL/latest/meta-data/instance-id")
    
    if [[ "$status_code" == "500" ]]; then
        success "Returns 500 for metadata of non-existent machine"
    else
        warn "Expected 500 for metadata of non-existent machine, got $status_code (may be expected)"
    fi
}

# Cleanup test resources
cleanup_test_resources() {
    log "Cleaning up test resources..."
    
    if [[ -n "${SSH_KEY_ID:-}" ]]; then
        curl -s -X DELETE "$NOOK_URL/api/v0/ssh-keys/$SSH_KEY_ID" > /dev/null
        log "Deleted test SSH key"
    fi
    
    if [[ -n "${MACHINE_ID:-}" ]]; then
        curl -s -X DELETE "$NOOK_URL/api/v0/machines/$MACHINE_ID" > /dev/null
        log "Deleted test machine"
    fi
    
    if [[ -n "${NETWORK_ID:-}" ]]; then
        curl -s -X DELETE "$NOOK_URL/api/v0/networks/$NETWORK_ID" > /dev/null
        log "Deleted test network"
    fi
}

# Main test runner
main() {
    echo "============================================"
    echo "Nook Integration Test Suite"
    echo "============================================"
    echo "Testing Nook service at: $NOOK_URL"
    echo "Test prefix: $TEST_PREFIX"
    echo
    
    # Set trap for cleanup
    trap cleanup_test_resources EXIT
    
    # Run tests
    wait_for_service || exit 1
    
    test_api_endpoints
    test_crud_operations
    test_nocloud_metadata
    test_error_conditions
    
    # Summary
    echo
    echo "============================================"
    echo "Test Summary"
    echo "============================================"
    echo "Total tests: $TOTAL_TESTS"
    echo "Failed tests: $FAILED_TESTS"
    echo "Passed tests: $((TOTAL_TESTS - FAILED_TESTS))"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        success "All tests passed! ✓"
        exit 0
    else
        fail "$FAILED_TESTS test(s) failed! ✗"
        exit 1
    fi
}

# Run tests
main "$@"