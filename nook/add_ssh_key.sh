#!/bin/bash

# Script to add SSH public key to nook service for cloud-init provisioning
# Usage: ./add_ssh_key.sh [machine_id] [ssh_key_path]

set -e

# Default values
DEFAULT_SSH_KEY="$HOME/.ssh/id_ed25519.pub"
NOOK_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
usage() {
    echo "Usage: $0 [machine_id] [ssh_key_path]"
    echo ""
    echo "Arguments:"
    echo "  machine_id      ID of the machine to add SSH key to (required)"
    echo "  ssh_key_path    Path to SSH public key file (optional, defaults to ~/.ssh/id_ed25519.pub)"
    echo ""
    echo "Examples:"
    echo "  $0 1                           # Add default SSH key to machine ID 1"
    echo "  $0 2 ~/.ssh/id_rsa.pub        # Add specific SSH key to machine ID 2"
    echo "  $0 1 /path/to/custom_key.pub  # Add custom SSH key to machine ID 1"
    echo ""
    echo "The script will:"
    echo "1. Read the SSH public key from the specified file"
    echo "2. Add it to the specified machine in nook service"
    echo "3. Verify the key was added by checking user-data endpoint"
}

# Check if machine_id is provided
if [ $# -lt 1 ]; then
    print_error "Machine ID is required"
    echo ""
    usage
    exit 1
fi

MACHINE_ID="$1"
SSH_KEY_PATH="${2:-$DEFAULT_SSH_KEY}"

# Check if SSH key file exists
if [ ! -f "$SSH_KEY_PATH" ]; then
    print_error "SSH key file not found: $SSH_KEY_PATH"
    exit 1
fi

# Read the SSH public key
print_status "Reading SSH public key from: $SSH_KEY_PATH"
SSH_KEY=$(cat "$SSH_KEY_PATH")

if [ -z "$SSH_KEY" ]; then
    print_error "SSH key file is empty or could not be read"
    exit 1
fi

print_status "SSH key found: ${SSH_KEY:0:50}..."

# Check if machine exists
print_status "Checking if machine ID $MACHINE_ID exists..."
MACHINE_INFO=$(curl -s -X GET "$NOOK_URL/api/v0/machines/$MACHINE_ID" 2>/dev/null)

if [ $? -ne 0 ] || [ "$MACHINE_INFO" = "null" ] || [ -z "$MACHINE_INFO" ]; then
    print_error "Machine ID $MACHINE_ID not found"
    exit 1
fi

# Extract machine details for display
MACHINE_NAME=$(echo "$MACHINE_INFO" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
MACHINE_HOSTNAME=$(echo "$MACHINE_INFO" | grep -o '"hostname":"[^"]*"' | cut -d'"' -f4)
MACHINE_IP=$(echo "$MACHINE_INFO" | grep -o '"ipv4":"[^"]*"' | cut -d'"' -f4)

print_status "Found machine: $MACHINE_NAME ($MACHINE_HOSTNAME) at $MACHINE_IP"

# Add SSH key to machine
print_status "Adding SSH key to machine ID $MACHINE_ID..."
ADD_RESPONSE=$(curl -s -X POST "$NOOK_URL/api/v0/ssh-keys" \
    -H "Content-Type: application/json" \
    -d "{\"machine_id\": $MACHINE_ID, \"key_text\": \"$SSH_KEY\"}" 2>/dev/null)

if [ $? -ne 0 ] || [ -z "$ADD_RESPONSE" ]; then
    print_error "Failed to add SSH key to nook service"
    exit 1
fi

# Check if the response contains an error
if echo "$ADD_RESPONSE" | grep -q "error"; then
    print_error "Failed to add SSH key: $ADD_RESPONSE"
    exit 1
fi

print_success "SSH key added successfully!"

# Verify the key was added by checking user-data
print_status "Verifying SSH key was added by checking user-data..."
USER_DATA=$(curl -s -X GET "$NOOK_URL/user-data" \
    --header "X-Forwarded-For: $MACHINE_IP" 2>/dev/null)

if [ $? -ne 0 ]; then
    print_warning "Could not verify SSH key in user-data (service may not be running)"
else
    if echo "$USER_DATA" | grep -q "$SSH_KEY"; then
        print_success "SSH key verified in user-data for machine $MACHINE_NAME"
    else
        print_warning "SSH key not found in user-data (may take a moment to propagate)"
    fi
fi

# Show summary
echo ""
print_success "SSH key setup complete!"
echo "Machine: $MACHINE_NAME ($MACHINE_HOSTNAME)"
echo "IP: $MACHINE_IP"
echo "SSH Key: ${SSH_KEY:0:50}..."
echo ""
print_status "The SSH key will be available in cloud-init user-data for this machine"
print_status "You can now provision VMs that will have SSH access with this key"
