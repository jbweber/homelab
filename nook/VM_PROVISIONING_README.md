# Nook VM Provisioning Scripts

This directory contains scripts to help provision VMs using the nook metadata service.

## Scripts

### `add_ssh_key.sh`
Adds an SSH public key to a machine in the nook service for cloud-init provisioning.

**Usage:**
```bash
./add_ssh_key.sh <machine_id> [ssh_key_path]
```

**Examples:**
```bash
# Add default SSH key (~/.ssh/id_ed25519.pub) to machine ID 1
./add_ssh_key.sh 1

# Add specific SSH key to machine ID 2
./add_ssh_key.sh 2 ~/.ssh/id_rsa.pub

# Add custom SSH key to machine ID 1
./add_ssh_key.sh 1 /path/to/custom_key.pub
```

**What it does:**
1. Reads the SSH public key from the specified file (defaults to ~/.ssh/id_ed25519.pub)
2. Verifies the machine exists
3. Adds the SSH key to the machine in nook
4. Verifies the key was added by checking the user-data endpoint

### `provision_vm.sh`
Complete VM provisioning script that creates a machine, allocates an IP, and adds SSH access.

**Usage:**
```bash
./provision_vm.sh <vm_name> <hostname>
```

**Examples:**
```bash
./provision_vm.sh my-web-server web-server.homelab
./provision_vm.sh database db-server.homelab
```

**What it does:**
1. Creates a machine in nook with network-based IP allocation
2. Adds your default SSH key to the machine
3. Displays all the information needed for cloud-init provisioning
4. Shows the allocated IP address and cloud-init URLs

## Cloud-Init Configuration

After provisioning a VM with these scripts, use the provided information in your cloud-init configuration:

```yaml
# Example cloud-init config
#cloud-config
hostname: web-server.homelab
users:
  - name: your-user
    ssh_authorized_keys:
      - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGQ8RqPbtHAG8UT4J/YeeZnSVfDwFRNjAHwuK2nPpkhX jweber@gravity.fe.cofront.xyz
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: sudo
    shell: /bin/bash

# Network configuration (let cloud-init handle it)
network:
  config: disabled

# Package updates
package_update: true
package_upgrade: true

# Your custom setup here
runcmd:
  - echo "VM provisioning complete!"
```

## Manual API Usage

If you prefer to use the API directly:

**Create Network:**
```bash
curl -X POST http://localhost:8080/api/v0/networks \
  -H "Content-Type: application/json" \
  -d '{"name": "virt-net", "bridge": "virt", "subnet": "10.37.37.0/24", "gateway": "10.37.37.1"}'
```

**Add DHCP Range:**
```bash
curl -X POST http://localhost:8080/api/v0/networks/1/dhcp \
  -H "Content-Type: application/json" \
  -d '{"StartIP": "10.37.37.100", "EndIP": "10.37.37.200", "LeaseTime": "12h"}'
```

**Create Machine:**
```bash
curl -X POST http://localhost:8080/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "my-vm", "hostname": "my-vm.homelab", "network_id": 1}'
```

**Add SSH Key:**
```bash
curl -X POST http://localhost:8080/api/v0/ssh-keys \
  -H "Content-Type: application/json" \
  -d '{"machine_id": 1, "key_text": "ssh-ed25519 AAAAC3..."}'
```

## Current Setup

- **Network:** virt-net (bridge: virt, subnet: 10.37.37.0/24)
- **DHCP Range:** 10.37.37.100 - 10.37.37.200
- **Gateway:** 10.37.37.1
- **SSH Key:** Your default ed25519 key is configured for all machines

## Troubleshooting

**Service not running:**
```bash
systemctl --user status nook
systemctl --user restart nook
```

**Check machine list:**
```bash
curl http://localhost:8080/api/v0/machines
```

**Check network configuration:**
```bash
curl http://localhost:8080/api/v0/networks
```

**Test cloud-init endpoints:**
```bash
curl -H "X-Forwarded-For: 10.37.37.100" http://localhost:8080/meta-data
curl -H "X-Forwarded-For: 10.37.37.100" http://localhost:8080/user-data
```
