# Nook Metadata Service Implementation

## Overview
Nook is a Go web service that provides metadata endpoints compatible with cloud-init's NoCloud data source. It is designed to serve dynamic instance metadata, user-data, vendor-data, and network configuration for virtual machines, with all data managed via a local SQLite database.

## Key Features
- **Dynamic Metadata**: `/meta-data` endpoint serves instance metadata based on the requestor's IP address, including instance-id, hostname, local-hostname, local-ipv4, public-hostname, and security-groups.
- **User Data**: `/user-data` endpoint serves a static or configurable cloud-init user-data script.
- **Vendor Data**: `/vendor-data` endpoint serves vendor-specific data (currently empty).
- **Network Config**: `/network-config` endpoint serves a basic network configuration in YAML format.
- **Machine Management API**: `/api/v0/machines` endpoints allow CRUD operations for VM metadata records, including name, hostname, and IPv4 address.

## How It Works
- When a VM requests `/meta-data`, the service extracts the requestor's IP and looks up the corresponding machine in the database.
- Metadata is generated dynamically from the machine record and returned in NoCloud-compatible YAML format.
- Other endpoints (`/user-data`, `/vendor-data`, `/network-config`) serve static or configurable data for cloud-init.
- The API allows you to create, list, update, and delete machine records, which are used to drive metadata responses.

## Endpoints
### NoCloud Metadata
- `GET /meta-data` — Returns instance metadata for the requestor's IP
- `GET /user-data` — Returns cloud-init user-data
- `GET /vendor-data` — Returns vendor-data (empty by default)
- `GET /network-config` — Returns network configuration

### Machine Management
- `GET /api/v0/machines` — List all machines
- `POST /api/v0/machines` — Create a new machine
- `GET /api/v0/machines/{id}` — Get machine by ID
- `DELETE /api/v0/machines/{id}` — Delete machine by ID
- `GET /api/v0/machines/name/{name}` — Get machine by name
- `GET /api/v0/machines/ipv4/{ipv4}` — Get machine by IPv4

## Database Schema
- **Machine**: `id`, `name`, `hostname`, `ipv4`
- **SSHKey**: `id`, `machine_id`, `key_text`

## Testing & Verification
- All endpoints are covered by Go unit tests using the testify library.
- Tests verify correct metadata generation, machine CRUD, and error handling.
- To run tests: `make test` or `go test ./internal/...`

## Improving & Extending
- The documentation and tests should be updated whenever new features or endpoints are added.
- Future improvements may include:
  - Customizable user-data and vendor-data per machine
  - More advanced network configuration
  - Additional metadata fields
  - API authentication and RBAC

## Usage
- Start the service and point your VM's cloud-init NoCloud data source to the Nook server.
- Use the machine management API to register VM metadata before booting.
- Metadata will be served dynamically based on the VM's IP address.

---

_Last updated: August 31, 2025_

## References

- [cloud-init NoCloud DataSource documentation](https://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html)
- [cloud-init Instance Metadata documentation](https://cloudinit.readthedocs.io/en/latest/topics/instancedata.html)
- [NoCloud metadata format example (Canonical)](https://discourse.ubuntu.com/t/nocloud-cloud-init-datasource/15312)
- [cloud-init source code (GitHub)](https://github.com/canonical/cloud-init)
