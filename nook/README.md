# Nook

Nook is a web service for managing metadata for cloud-init running on virtual machines.

## Project Structure

```
nook/
├── cmd/nook/          # Main application entry point
├── internal/
│   ├── api/          # HTTP API handlers
│   └── datastore/    # Database operations
├── go.mod
├── go.sum
└── Makefile          # Build and test automation
```

## Quick Start

### Prerequisites

- Go 1.24.6 or later
- SQLite (automatically handled by modernc.org/sqlite)

### Building

```bash
# Build the binary
make build

# Build for Linux (cross-compilation)
make build-linux
```

### Running

```bash
# Run directly
make run

# Or build and run
make build && ./nook
```

### Testing

```bash
# Run all tests
make test
# Run tests with coverage validation (requires 80% coverage)
make test-coverage-validate

# Run tests with race detection
make test-race

# Show coverage in terminal
make coverage-func
```

### Real-World Testing with libvirt

Nook has been successfully tested with real VMs using libvirt and Fedora UKI images. This setup validates the cloud-init integration and metadata delivery.

#### Setup Overview
- **VM Image**: Fedora Cloud Base UKI (Unified Kernel Image) for UEFI boot
- **Network**: Bridge interface with dnsmasq for DHCP and DNS
- **Metadata Source**: Nook running on host at `http://10.37.37.2:8080`
- **Configuration**: Dynamic user-data with hostname and SSH keys per VM

#### Key Components
- **Systemd Service**: User service for easy management (`systemctl --user restart nook`)
- **Dynamic User-Data**: Serves customized cloud-config based on VM IP
- **Logging**: Detailed request logging for debugging
- **DHCP**: dnsmasq provides static IPs and gateway/DNS options

#### Testing Steps
1. Configure dnsmasq on bridge for DHCP with static leases
2. Add VM to nook database with expected IP
3. Add SSH keys to VM in nook
4. Define and start VM with nocloud datasource
5. VM boots, fetches metadata, sets hostname, installs SSH keys

#### Verified Endpoints
- `/meta-data`: Instance metadata (YAML)
- `/user-data`: Cloud-config with SSH keys and hostname
- `/vendor-data`: Empty (optional)

#### Files
- `nook/testing/nook.service`: Systemd user service file
- `.gitignore`: Excludes test files from repo

### Coverage

Current test coverage: **75.6%** (api package), **62.5%** (datastore package)

**Recent Improvements (August 2025):**
- SSH key handlers coverage significantly improved:
  - `PublicKeysHandler`: 95.5%
  - `PublicKeyByIdxHandler`: 96.6%
  - `PublicKeyOpenSSHHandler`: 82.8%
- Added comprehensive error branch tests for malformed remote addresses, database errors, and invalid inputs
- Coverage threshold: 80% (enforced in CI via `make test-coverage-validate`)

**Remaining Gaps:**
- Datastore package: Focus on `ListAllSSHKeys` (0% coverage) and other CRUD operations
- API handlers: Some edge cases in metadata and network handlers
- Integration test coverage for real database scenarios

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


The service provides the following endpoints for cloud-init metadata:

### PATCH /api/v0/machines/{id}

Update an existing machine's details.

**Request:**

```
PATCH /api/v0/machines/{id}
Content-Type: application/json

{
	"name": "new-name",
	"hostname": "new-hostname",
	"ipv4": "192.168.1.123"
}
```

**Response (200 OK):**

```
{
	"id": 1,
	"name": "new-name",
	"hostname": "new-hostname",
	"ipv4": "192.168.1.123"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid ID, missing fields, or invalid JSON/IP format
- `404 Not Found`: Machine not found
- `500 Internal Server Error`: Database or update error

**Example:**

```bash
curl -X PATCH http://localhost:8080/api/v0/machines/1 \
	-H "Content-Type: application/json" \
	-d '{"name":"new-name","hostname":"new-hostname","ipv4":"192.168.1.123"}'
```

---


### GET /2021-01-03/meta-data/public-keys

Returns a list of SSH public keys for the requesting machine in cloud-init format (one key per line).

**Request:**

```
GET /2021-01-03/meta-data/public-keys
Header: X-Forwarded-For: <machine-ip>
```

**Response (200 OK):**

Content-Type: text/plain

```
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtestkey1
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2
```

**Error Responses:**

- `404 Not Found`: Machine not found for IP
- `500 Internal Server Error`: Datastore error

**Example:**

```bash
curl -H "X-Forwarded-For: 192.168.1.170" http://localhost:8080/2021-01-03/meta-data/public-keys
```

---

### GET /2021-01-03/meta-data/public-keys/{idx}

Returns the SSH public key at the specified index for the requesting machine.

**Request:**

```
GET /2021-01-03/meta-data/public-keys/{idx}
Header: X-Forwarded-For: <machine-ip>
```

**Response (200 OK):**

Content-Type: text/plain

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2
```

**Error Responses:**

- `400 Bad Request`: Invalid index
- `404 Not Found`: Machine not found for IP, or key index out of range
- `500 Internal Server Error`: Datastore error

**Example:**

```bash
curl -H "X-Forwarded-For: 192.168.1.170" http://localhost:8080/2021-01-03/meta-data/public-keys/1
```

---

### GET /2021-01-03/meta-data/public-keys/{idx}/openssh-key

Returns the OpenSSH-formatted public key at the specified index for the requesting machine.

**Request:**

```
GET /2021-01-03/meta-data/public-keys/{idx}/openssh-key
Header: X-Forwarded-For: <machine-ip>
```

**Response (200 OK):**

Content-Type: text/plain

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestkey2
```

**Error Responses:**

- `400 Bad Request`: Invalid index
- `404 Not Found`: Machine not found for IP, or key index out of range
- `500 Internal Server Error`: Datastore error

**Example:**

```bash
curl -H "X-Forwarded-For: 192.168.1.170" http://localhost:8080/2021-01-03/meta-data/public-keys/1/openssh-key
```


---

### Skipped Endpoint: /latest/api/token

The `/latest/api/token` endpoint (used for IMDSv2 session tokens in AWS EC2) is intentionally skipped in this implementation. It is not required for NoCloud compatibility, but may be added in the future to support enhanced metadata security and compatibility with cloud-init clients expecting IMDSv2.

If you need IMDSv2-style session tokens, see [AWS IMDSv2 documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html) for details. For now, all metadata endpoints are accessible without a session token.

---

Endpoints:

- `/2021-01-03/dynamic/instance-identity/document`
- `/2021-01-03/meta-data/public-keys/<int:idx>`
- `/2021-01-03/meta-data/public-keys/<int:idx}/openssh-key`
- `/meta-data`
- `/user-data`
- `/vendor-data`
- `/api/v0/machines`
- `/api/v0/networks`

### Skipped: /api/v0/networks

The `/api/v0/networks` endpoint is currently a placeholder and does not return real network data. This endpoint is complex and will be revisited in the future to provide actual network configuration and metadata. For now, it is intentionally skipped and documented as such.
- `/api/v0/ssh-keys`

## Database

The service uses SQLite for local metadata storage via `modernc.org/sqlite`.

## Makefile Targets

Run `make help` to see all available targets and their descriptions.
