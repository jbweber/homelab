# Nook

Nook is a web service for managing metadata for cloud-init running on virtual machines.

This is a CURSED experiment at using AI for software development. This code may work but I would not recommend people who are not me use it for anything important.

## Project Structure

```
nook/
├── cmd/nook/              # Main application entry point
├── internal/
│   ├── api/              # HTTP API handlers
│   ├── domain/           # Domain models
│   ├── repository/       # Data access layer
│   ├── migrations/       # Database migrations
│   └── testutil/         # Testing utilities
├── testing/
│   └── nook.service      # Systemd user service file
├── go.mod
├── go.sum
├── Makefile              # Build and test automation
├── test_api.sh           # Integration test script
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.24.6 or later
- SQLite (automatically handled by modernc.org/sqlite)

### Building

```bash
# Build the binary (copies to ~/nook/bin/)
make build

# Build for Linux (cross-compilation)
make build-linux
```

### Running

#### Development Mode
```bash
# Run directly with default settings
make run

# Or build and run with custom settings
make build && ./nook server --db-path ./nook.db --port 8080
```

#### Production Mode (Systemd User Service)
```bash
# Copy service file and start
mkdir -p ~/.config/systemd/user
cp testing/nook.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable nook
systemctl --user start nook

# Check status
systemctl --user status nook

# View logs
journalctl --user -u nook -f
```

### Testing

```bash
# Run all tests
make test

# Run integration tests (starts/stops server automatically)
./test_api.sh

# Run tests with coverage validation (requires 80% coverage)
make test-coverage-validate

# Run tests with race detection
make test-race

# Show coverage in terminal
make coverage-func
```
make test-coverage-validate

# Run tests with race detection
make test-race

# Show coverage in terminal
make coverage-func
```

### Real-World Testing with libvirt

Nook has been successfully tested with real VMs using libvirt and Fedora UKI images. This setup validates the cloud-init integration and metadata delivery.

#### Production Deployment Structure
```
~/nook/
├── bin/
│   └── nook          # Production binary (auto-deployed by make build)
└── data/
    └── nook.db       # Production database
```

#### Setup Overview
- **VM Image**: Fedora Cloud Base UKI (Unified Kernel Image) for UEFI boot
- **Network**: Bridge interface with dnsmasq for DHCP and DNS
- **Metadata Source**: Nook running via systemd user service on port 8080
- **Configuration**: Dynamic user-data with hostname and SSH keys per VM

#### Key Components
- **Systemd User Service**: Automatic management (`systemctl --user restart nook`)
- **Dynamic User-Data**: Serves customized cloud-config based on VM IP
- **Logging**: Detailed request logging via `journalctl --user -u nook`
- **DHCP**: dnsmasq provides static IPs and gateway/DNS options
- **Test Isolation**: Separate test database and port (8081) for development

#### Deployment Steps
1. **Build and Deploy:**
   ```bash
   make build  # Builds and copies binary to ~/nook/bin/
   ```

2. **Setup Systemd Service:**
   ```bash
   cp testing/nook.service ~/.config/systemd/user/
   systemctl --user daemon-reload
   systemctl --user enable nook
   systemctl --user start nook
   ```

3. **Configure dnsmasq** on bridge for DHCP with static leases

4. **Add VM to nook database:**
   ```bash
   curl -X POST http://localhost:8080/api/v0/machines \
     -H "Content-Type: application/json" \
     -d '{"name":"test-vm","hostname":"test.example.com","ipv4":"10.37.37.100"}'
   ```

5. **Add SSH keys to VM:**
   ```bash
   # Add SSH key via API
   curl -X POST http://localhost:8080/api/v0/ssh-keys \
     -H "Content-Type: application/json" \
     -d '{"machine_id":1,"key_text":"ssh-rsa AAAAB3NzaC1yc2E..."}'
   ```

6. **Define and start VM** with nocloud datasource pointing to nook

#### Verified Endpoints
- `/meta-data`: Instance metadata (YAML format)
- `/user-data`: Cloud-config with SSH keys and hostname
- `/vendor-data`: Empty (optional)
- `/api/v0/machines`: Machine management API
- `/api/v0/ssh-keys`: SSH key management API

#### Testing Commands
```bash
# Test production service
curl http://localhost:8080/
curl -H "X-Forwarded-For: 10.37.37.100" http://localhost:8080/meta-data

# Test development version (isolated)
./test_api.sh

# Check service status
systemctl --user status nook
journalctl --user -u nook -f
```

#### Files
- `testing/nook.service`: Systemd user service configuration
- `test_api.sh`: Integration test script with automatic cleanup
- `.gitignore`: Excludes test files (`test_nook.db`, etc.)

### Coverage

Current test coverage: **75.6%** (api package), **62.5%** (datastore package)

**Recent Improvements (September 2025):**
- **Streamlined codebase**: Removed unused EC2-style endpoints for nocloud compatibility
- **Enhanced test script**: `test_api.sh` with automatic cleanup and process management
- **Systemd integration**: Production deployment with user service management
- **Database isolation**: Separate test and production databases
- **Build automation**: Auto-deployment to `~/nook/bin/` on build

**Coverage Focus Areas:**
- API handlers: Core cloud-init endpoints (`/meta-data`, `/user-data`, `/vendor-data`)
- Machine and SSH key management APIs
- Error handling for malformed requests and database errors
- Integration tests with real database scenarios

**Test Infrastructure:**
- **Unit Tests**: Comprehensive coverage of individual components
- **Integration Tests**: `test_api.sh` validates full API workflow
- **Process Management**: Automatic server lifecycle management in tests
- **Database Cleanup**: Test databases removed after execution

### Development

```bash
# Format code
make fmt

# Vet code for issues
make vet

# Check formatting (fails if not formatted)
make check-fmt

# Setup development environment
make dev-setup

# Run full CI pipeline locally
make ci
```

### Code Quality Tools

```bash
# Install development tools (golangci-lint, gosec)
make install-tools

# Lint code (requires golangci-lint)
make lint

# Security check (requires gosec)
make security
```

### Cleaning

```bash
# Clean build artifacts and coverage files
make clean
```

## API Endpoints

The service provides the following endpoints for cloud-init metadata and management:

### Cloud-Init Metadata Endpoints

#### GET /meta-data
Returns instance metadata in cloud-init format.

**Request:**
```
GET /meta-data
X-Forwarded-For: <client-ip>
```

**Response (200 OK):**
```yaml
instance-id: iid-00000001
hostname: my-host
local-hostname: my-host
local-ipv4: 192.168.1.100
public-hostname: my-host
security-groups: default
```

#### GET /user-data
Returns user-data in cloud-config format with SSH keys and hostname.

**Request:**
```
GET /user-data
X-Forwarded-For: <client-ip>
```

**Response (200 OK):**
```yaml
#cloud-config
hostname: my-host
manage_etc_hosts: true
ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2E...
```

#### GET /vendor-data
Returns vendor-data (currently empty for nocloud compatibility).

**Request:**
```
GET /vendor-data
```

**Response (200 OK):**
```yaml
# Empty vendor data
```

### Management API Endpoints

#### GET /api/v0/machines
List all machines.

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "web-server",
    "hostname": "web.example.com",
    "ipv4": "192.168.1.100"
  }
]
```

#### POST /api/v0/machines
Create a new machine.

**Request:**
```json
{
  "name": "new-server",
  "hostname": "new.example.com",
  "ipv4": "192.168.1.101"
}
```

#### GET /api/v0/machines/{id}
Get a specific machine by ID.

#### PATCH /api/v0/machines/{id}
Update an existing machine.

**Request:**
```json
{
  "name": "updated-name",
  "hostname": "updated.example.com",
  "ipv4": "192.168.1.102"
}
```

#### DELETE /api/v0/machines/{id}
Delete a machine.

#### GET /api/v0/ssh-keys
List all SSH keys.

#### POST /api/v0/ssh-keys
Add an SSH key to a machine.

**Request:**
```json
{
  "machine_id": 1,
  "key_text": "ssh-rsa AAAAB3NzaC1yc2E..."
}
```

#### DELETE /api/v0/ssh-keys/{id}
Delete an SSH key.

### Health Check

#### GET /
Basic health check endpoint.

**Response (200 OK):**
```
Nook web service is running!
```

## Database

The service uses SQLite for local metadata storage via `modernc.org/sqlite`.

## Makefile Targets

Run `make help` to see all available targets and their descriptions.
