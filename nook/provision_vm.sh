#!/bin/bash

# Script to provision a new VM using nook service for IP allocation and SSH key injection
# Usage: ./provision_vm.sh <vm_name> <hostname>

set -e

# Default values
NOOK_URL="http://localhost:8080"
NETWORK_ID=1  # Default to first network (virt-net)

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
    echo "Usage: $0 <vm_name> <hostname>"
    echo ""
    echo "Arguments:"
    echo "  vm_name    Name of the VM (required)"
    echo "  hostname   Hostname for the VM (required)"
    echo ""
    echo "Examples:"
    echo "  $0 my-vm my-vm.homelab"
    echo "  $0 web-server web-server.homelab"
    echo ""
    echo "This script will:"
    echo "1. Create a machine entry in nook with network-based IP allocation"
    echo "2. Add your default SSH key to the machine"
    echo "3. Display the allocated IP and cloud-init metadata"
    echo ""
    echo "After running this script, you can use the allocated IP with cloud-init"
    echo "to provision your VM with automatic network configuration and SSH access."
}

# Check arguments
if [ $# -lt 2 ]; then
    print_error "VM name and hostname are required"
    echo ""
    usage
    exit 1
fi

VM_NAME="$1"
VM_HOSTNAME="$2"

# Check if SSH key exists
SSH_KEY_PATH="$HOME/.ssh/id_ed25519.pub"
if [ ! -f "$SSH_KEY_PATH" ]; then
    print_error "SSH public key not found at: $SSH_KEY_PATH"
    print_error "Please ensure you have an SSH key pair generated"
    exit 1
fi

print_status "Provisioning VM: $VM_NAME ($VM_HOSTNAME)"
print_status "Using network ID: $NETWORK_ID"

# Create machine with network-based IP allocation
print_status "Creating machine in nook service..."
CREATE_RESPONSE=$(curl -s -X POST "$NOOK_URL/api/v0/machines" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$VM_NAME\", \"hostname\": \"$VM_HOSTNAME\", \"network_id\": $NETWORK_ID}" 2>/dev/null)

if [ $? -ne 0 ] || [ -z "$CREATE_RESPONSE" ]; then
    print_error "Failed to create machine in nook service"
    exit 1
fi

# Check for error in response
if echo "$CREATE_RESPONSE" | grep -q "error"; then
    print_error "Failed to create machine: $CREATE_RESPONSE"
    exit 1
fi

# Extract machine details
MACHINE_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)
ALLOCATED_IP=$(echo "$CREATE_RESPONSE" | grep -o '"ipv4":"[^"]*"' | cut -d'"' -f4)

if [ -z "$MACHINE_ID" ] || [ -z "$ALLOCATED_IP" ]; then
    print_error "Failed to parse machine creation response"
    exit 1
fi

print_success "Machine created with ID: $MACHINE_ID"
print_success "Allocated IP: $ALLOCATED_IP"

# Add SSH key to the machine
print_status "Adding SSH key to machine..."
SSH_KEY=$(cat "$SSH_KEY_PATH")
ADD_KEY_RESPONSE=$(curl -s -X POST "$NOOK_URL/api/v0/ssh-keys" \
    -H "Content-Type: application/json" \
    -d "{\"machine_id\": $MACHINE_ID, \"key_text\": \"$SSH_KEY\"}" 2>/dev/null)

if [ $? -ne 0 ] || [ -z "$ADD_KEY_RESPONSE" ]; then
    print_warning "Failed to add SSH key (machine still created successfully)"
else
    print_success "SSH key added to machine"
fi

# Display provisioning information
echo ""
print_success "VM Provisioning Complete!"
echo "=========================================="
echo "VM Name:     $VM_NAME"
echo "Hostname:    $VM_HOSTNAME"
echo "IP Address:  $ALLOCATED_IP"
echo "Machine ID:  $MACHINE_ID"
echo ""
echo "Cloud-init URLs:"
echo "  Metadata:  $NOOK_URL/meta-data"
echo "  User-data: $NOOK_URL/user-data"
echo ""
echo "To provision your VM, use these URLs in your cloud-init configuration"
echo "with X-Forwarded-For header set to: $ALLOCATED_IP"
echo ""
print_status "Example cloud-init network config:"
echo "  network:"
echo "    config: disabled"
echo "    wait-online: true"
echo ""
print_status "Your VM will automatically get network access and SSH access with your key!"
