#!/bin/bash

# Nook API Bash Examples
# This script demonstrates how to interact with Nook API using curl

NOOK_URL="${NOOK_URL:-http://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to create a network
create_network() {
    local name="$1"
    local bridge="$2" 
    local subnet="$3"
    local gateway="$4"
    local description="$5"
    
    log "Creating network: $name"
    
    curl -s -X POST "$NOOK_URL/api/v0/networks" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$name\",
            \"bridge\": \"$bridge\",
            \"subnet\": \"$subnet\",
            \"gateway\": \"$gateway\",
            \"description\": \"$description\"
        }" | jq '.'
}

# Function to create a machine
create_machine() {
    local name="$1"
    local hostname="$2"
    local ipv4="$3"
    
    log "Creating machine: $name"
    
    curl -s -X POST "$NOOK_URL/api/v0/machines" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$name\",
            \"hostname\": \"$hostname\",
            \"ipv4\": \"$ipv4\"
        }" | jq '.'
}

# Function to add SSH key
add_ssh_key() {
    local machine_id="$1"
    local key_file="$2"
    
    if [[ ! -f "$key_file" ]]; then
        error "SSH key file not found: $key_file"
        return 1
    fi
    
    local key_text
    key_text=$(cat "$key_file")
    
    log "Adding SSH key from $key_file to machine $machine_id"
    
    curl -s -X POST "$NOOK_URL/api/v0/ssh-keys" \
        -H "Content-Type: application/json" \
        -d "{
            \"machine_id\": $machine_id,
            \"key_text\": \"$key_text\"
        }" | jq '.'
}

# Function to list machines
list_machines() {
    log "Listing all machines"
    curl -s "$NOOK_URL/api/v0/machines" | jq '.'
}

# Function to list networks
list_networks() {
    log "Listing all networks"
    curl -s "$NOOK_URL/api/v0/networks" | jq '.'
}

# Function to list SSH keys
list_ssh_keys() {
    log "Listing all SSH keys"
    curl -s "$NOOK_URL/api/v0/ssh-keys" | jq '.'
}

# Function to get machine by name
get_machine_by_name() {
    local name="$1"
    log "Getting machine by name: $name"
    curl -s "$NOOK_URL/api/v0/machines/name/$name" | jq '.'
}

# Function to get machine metadata (NoCloud format)
get_machine_metadata() {
    local machine_ip="$1"
    log "Getting metadata for machine at IP: $machine_ip"
    
    # Simulate request from the machine
    echo "Instance ID:"
    curl -s -H "X-Forwarded-For: $machine_ip" "$NOOK_URL/latest/meta-data/instance-id"
    echo
    
    echo "Hostname:"
    curl -s -H "X-Forwarded-For: $machine_ip" "$NOOK_URL/latest/meta-data/local-hostname"
    echo
    
    echo "User Data:"
    curl -s -H "X-Forwarded-For: $machine_ip" "$NOOK_URL/latest/user-data"
    echo
}

# Function to delete machine
delete_machine() {
    local machine_id="$1"
    warn "Deleting machine ID: $machine_id"
    curl -s -X DELETE "$NOOK_URL/api/v0/machines/$machine_id"
    log "Machine deleted"
}

# Function to delete SSH key
delete_ssh_key() {
    local key_id="$1"
    warn "Deleting SSH key ID: $key_id"
    curl -s -X DELETE "$NOOK_URL/api/v0/ssh-keys/$key_id"
    log "SSH key deleted"
}

# Function to test Nook service health
health_check() {
    log "Checking Nook service health"
    
    if curl -s --fail "$NOOK_URL/api/v0/machines" > /dev/null; then
        log "✓ Nook service is healthy"
        return 0
    else
        error "✗ Nook service is not responding"
        return 1
    fi
}

# Example workflow
example_workflow() {
    log "Running example workflow"
    
    # Health check first
    health_check || return 1
    
    # Create network
    local network_result
    network_result=$(create_network "demo-net" "br0" "192.168.200.0/24" "192.168.200.1" "Demo network")
    local network_id
    network_id=$(echo "$network_result" | jq -r '.id')
    
    if [[ "$network_id" == "null" ]]; then
        error "Failed to create network"
        return 1
    fi
    
    # Create machine
    local machine_result
    machine_result=$(create_machine "demo-vm" "demo-vm.local" "192.168.200.10")
    local machine_id
    machine_id=$(echo "$machine_result" | jq -r '.id')
    
    if [[ "$machine_id" == "null" ]]; then
        error "Failed to create machine"
        return 1
    fi
    
    # Add SSH key (if exists)
    if [[ -f ~/.ssh/id_rsa.pub ]]; then
        add_ssh_key "$machine_id" ~/.ssh/id_rsa.pub
    else
        warn "SSH key file ~/.ssh/id_rsa.pub not found, skipping SSH key addition"
    fi
    
    # List resources
    echo
    list_machines
    echo
    list_networks
    echo
    list_ssh_keys
    echo
    
    # Get metadata
    get_machine_metadata "192.168.200.10"
}

# Main menu
main() {
    case "${1:-help}" in
        "create-network")
            create_network "$2" "$3" "$4" "$5" "$6"
            ;;
        "create-machine")
            create_machine "$2" "$3" "$4"
            ;;
        "add-ssh-key")
            add_ssh_key "$2" "$3"
            ;;
        "list-machines")
            list_machines
            ;;
        "list-networks")
            list_networks
            ;;
        "list-ssh-keys")
            list_ssh_keys
            ;;
        "get-machine")
            get_machine_by_name "$2"
            ;;
        "get-metadata")
            get_machine_metadata "$2"
            ;;
        "delete-machine")
            delete_machine "$2"
            ;;
        "delete-ssh-key")
            delete_ssh_key "$2"
            ;;
        "health")
            health_check
            ;;
        "example")
            example_workflow
            ;;
        "help"|*)
            echo "Nook API Bash Client"
            echo
            echo "Usage: $0 <command> [arguments...]"
            echo
            echo "Commands:"
            echo "  create-network <name> <bridge> <subnet> <gateway> <description>"
            echo "  create-machine <name> <hostname> <ipv4>"
            echo "  add-ssh-key <machine_id> <key_file>"
            echo "  list-machines"
            echo "  list-networks"
            echo "  list-ssh-keys"
            echo "  get-machine <name>"
            echo "  get-metadata <machine_ip>"
            echo "  delete-machine <machine_id>"
            echo "  delete-ssh-key <key_id>"
            echo "  health"
            echo "  example"
            echo
            echo "Environment variables:"
            echo "  NOOK_URL - Base URL for Nook service (default: http://localhost:8080)"
            ;;
    esac
}

main "$@"