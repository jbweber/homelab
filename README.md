# homelab

# Remote Development Setup

This project is configured for remote development using VS Code and the Remote - SSH extension. Code is edited locally but lives and runs on your Linux machine.

## Getting Started

1. Install the [Remote - SSH extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh) in VS Code.
2. Ensure you have SSH access to your Linux machine.
3. Update `.vscode/settings.json` if your username is not `jweber`:
	- Change `/home/jweber/projects/homelab` to your actual home directory path.
4. Open the folder in VS Code using "Remote-SSH: Connect to Host...".

## Recommended Extensions

Extensions are listed in `.vscode/extensions.json`. VS Code will prompt you to install them when you open the project.

## Notes

- Each user should update the workspace path in `.vscode/settings.json` to match their Linux username if different.

## PreRequisites

```bash
$ sudo dnf install ansible-core ansible-collection-ansible-posix
```

## Secrets and Vault Usage

You must create an Ansible vault file at `~/.homelab.vault` (outside this repository).

This vault file should contain at least the following variables:

- `ansible_user`: The username to SSH as for remote connections.
- `ansible_become_password`: The sudo password for privilege escalation.

Example vault file contents:

```yaml
ansible_user: your_ssh_username
ansible_become_password: your_sudo_password
```

This file is referenced in all playbooks via `vars_files: [ ~/.homelab.vault ]` and must exist for automation to work.

---

# Nook Metadata Service

## Endpoint Separation

The API is split into two groups:

- **Cloud-init endpoints**: Used by cloud-init clients to retrieve instance metadata and user data. These endpoints are versioned and follow the EC2/NoCloud conventions.
- **Management endpoints**: Used for managing machines, networks, and SSH keys. These are not accessed by cloud-init and are intended for administrative use.

### Cloud-init Endpoints
- `/2021-01-03/meta-data` and subpaths
- `/2021-01-03/user-data`
- `/2021-01-03/vendor-data`
- `/2021-01-03/dynamic/instance-identity/document`

### Management Endpoints
- `/api/v0/machines`
- `/api/v0/networks`
- `/api/v0/ssh-keys`

## Network Management Features

Nook now includes comprehensive network management for automatic IP allocation:

### Network Configuration
- **Network Creation**: Define networks with subnet, bridge, and gateway
- **DHCP Ranges**: Configure IP ranges for automatic allocation
- **IP Conflict Detection**: Prevents conflicts between static and dynamic IPs

### Automatic IP Allocation
- **Network-Based**: IPs allocated from appropriate network DHCP ranges
- **Lease Management**: Tracks IP leases with expiration times
- **Cloud-Init Ready**: Allocated IPs automatically included in VM metadata

### Example Network Setup
```bash
# Create network
curl -X POST http://localhost:8080/api/v0/networks \
  -H "Content-Type: application/json" \
  -d '{"name": "virt-net", "bridge": "virt", "subnet": "10.37.37.0/24", "gateway": "10.37.37.1"}'

# Add DHCP range
curl -X POST http://localhost:8080/api/v0/networks/1/dhcp \
  -H "Content-Type: application/json" \
  -d '{"StartIP": "10.37.37.100", "EndIP": "10.37.37.200", "LeaseTime": "12h"}'

# Create machine with auto-IP
curl -X POST http://localhost:8080/api/v0/machines \
  -H "Content-Type: application/json" \
  -d '{"name": "web-server", "hostname": "web.homelab", "network_id": 1}'
```

## Automation Scripts

Two scripts are provided for VM provisioning workflow:

- **`provision_vm.sh`**: Complete VM setup with IP allocation and SSH key injection
- **`add_ssh_key.sh`**: Add SSH keys to existing machines

See `nook/VM_PROVISIONING_README.md` for detailed usage and examples.

## IP Validation Rules

- Cloud-init endpoints require a valid IP address in the `X-Forwarded-For` header. Requests without a valid IP will be rejected.
- Management endpoints do **not** require IP validation and are accessible for administrative operations.

## Documentation Practices

- All endpoints and their behaviors are documented in `API.md`.
- Any changes to endpoint logic or validation rules must be reflected in both `API.md` and this `README.md`.
- For development and testing instructions, see `API.md` for details on running, testing, and validating the service.
