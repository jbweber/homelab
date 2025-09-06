# Nook Examples

This directory contains examples and documentation for using the Nook VM metadata service.

## Contents

- `cloud-init-examples/` - Cloud-init configuration examples
- `network-configs/` - Network configuration examples  
- `vm-provisioning/` - VM provisioning workflows
- `api-usage/` - API usage examples
- `testing/` - Testing and validation examples

## Quick Start

1. **Start the Nook service:**
   ```bash
   nook server --port 8080
   ```

2. **Add a network:**
   ```bash
   nook add-network --name "lan" --bridge "br0" --subnet "192.168.1.0/24"
   ```

3. **Add a machine:**
   ```bash
   nook add-machine --name "vm01" --hostname "vm01.local" --ip "192.168.1.10"
   ```

4. **Add SSH keys:**
   ```bash
   nook add-ssh-key --machine-name "vm01" --key-file ~/.ssh/id_rsa.pub
   ```

5. **Configure VM to use Nook as metadata service:**
   Set the metadata service URL to `http://YOUR_NOOK_IP:8080/`

## API Endpoints

- `GET /latest/meta-data/` - NoCloud metadata directory
- `GET /latest/meta-data/instance-id` - Instance ID
- `GET /latest/meta-data/local-hostname` - Local hostname
- `GET /latest/user-data` - User data (cloud-init)
- `GET /latest/vendor-data` - Vendor data
- `GET /latest/network-config` - Network configuration

## Management API

- `GET /api/v0/machines` - List machines
- `POST /api/v0/machines` - Create machine
- `GET /api/v0/networks` - List networks
- `POST /api/v0/networks` - Create network
- `GET /api/v0/ssh-keys` - List SSH keys
- `POST /api/v0/ssh-keys` - Add SSH key