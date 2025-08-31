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

### Coverage Goals & Workflow

- **Minimum Coverage Goal:** All code in `internal/api` and `internal/datastore` must maintain at least **75% test coverage**. This is enforced as part of the development workflow.
- Coverage is checked with `make test-coverage` and reviewed before merging or releasing new features.
- All endpoints are covered by Go unit tests using the testify library.
- Integration tests are provided in `test_api.sh` and cover:
  - Metadata lookup by IP (including X-Forwarded-For)
  - Machine CRUD operations
  - Error cases: invalid IPv4, duplicate IPv4, missing fields, non-existent IP, invalid machine ID
- Error handling and validation:
  - Machine creation validates required fields and IPv4 format
  - Duplicate IPv4 addresses are rejected
  - All error responses are returned as JSON with clear messages and appropriate status codes
  - Metadata endpoint supports X-Forwarded-For for proxy compatibility
- To run tests: `make test` or `go test ./internal...`
- To run integration tests: `bash nook/test_api.sh`

#### Coverage Visualization in VS Code

For an improved workflow, it is recommended to install the **Go Coverage Gutters** extension in VS Code. This extension highlights covered and uncovered lines directly in the editor, making it easy to target gaps in test coverage. See the recommended extensions in the workspace config for details.

## Improving & Extending

## Advanced Endpoints Implementation Plan

### Overview
To achieve full parity with the Python reference and ensure compatibility with cloud-init, the following advanced endpoints and features will be implemented:

1. `/2021-01-03/dynamic/instance-identity/document`: Returns a JSON document with instance identity info. Not required for NoCloud, but useful for EC2 compatibility and future extensibility.
2. `/2021-01-03/meta-data/public-keys`, `/2021-01-03/meta-data/public-keys/<int:idx>`, `/2021-01-03/meta-data/public-keys/<int:idx>/openssh-key`: Returns public keys in plain text or JSON. Not required for NoCloud, but useful for compatibility with cloud-init’s EC2 mode.
3. `/latest/api/token`: Returns a session token. Not required for NoCloud, but used for EC2-compatible datasources.
4. DHCP mapping: Network-config can be generated dynamically per machine, supporting both v1 and v2 YAML formats.
5. Dynamic user-data/meta-data: All endpoints support per-machine dynamic data, not static files.
6. Raw data endpoints: Optional, for debugging or advanced use cases.
7. YAML migration: All metadata endpoints support YAML where required (meta-data, network-config).

### Endpoint Formats
- `meta-data`: YAML, must contain at least `instance-id`, can include `local-hostname`, `network-interfaces`, etc.
- `user-data`: Cloud-config or script, plain text.
- `vendor-data`: Optional, plain text or cloud-config.
- `network-config`: YAML, supports v1 or v2 network config formats.
- `instance-identity/document`: JSON document with instance identity info (see EC2 format for reference).
- `public-keys`: Plain text or JSON, compatible with EC2 metadata format.
- `api/token`: Plain text token string.

### Design Decisions
- All endpoints will be implemented as HTTP handlers in Go, using the chi router.
- Data will be sourced dynamically from the SQLite database, supporting per-machine configuration.
- Formats will match cloud-init documentation and EC2-compatible datasources where applicable.
- Integration tests in `test_api.sh` will be updated to cover all new endpoints and error cases.
- Documentation will be updated as new features are added.

---

## Development Workflow

All contributors should follow these workflow steps:

- **Coding Standards:**
  - All Go code must follow idiomatic Go practices and formatting (see `.github/copilot-instructions.md`).
  - Use `gofmt` and `goimports` for formatting and import management.
  - Organize binaries in `nook/cmd/` and packages in `nook/internal/`.

- **Testing & Coverage:**
  - All new features and bugfixes must include unit tests.
  - Minimum coverage for `internal/api` and `internal/datastore` is 75% (check with `make test-coverage`).
  - After running coverage, open `nook/coverage.html` and review uncovered lines. Add targeted tests for any remaining uncovered code before merging or releasing.
  - Integration tests in `test_api.sh` should be updated for new endpoints and error cases.

- **Build & CI:**
  - Use the Makefile for all build, test, coverage, and CI operations:
    - `make build` — build the binary
    - `make test` — run all unit tests
    - `make test-coverage` — run tests and generate coverage report
    - `make ci` — run full CI pipeline locally

- **Documentation:**
  - Update `IMPLEMENTATION.md` and endpoint documentation for all new features.
  - Document any custom roles or playbooks in `README.md` if relevant.

- **Roadmap & Standards:**
  - See [`nook/ROADMAP.md`](./ROADMAP.md) for priorities and future plans.
  - See `.github/copilot-instructions.md` for detailed standards and expectations.

---
## Usage


_Last updated: August 31, 2025_

## References

## Roadmap
See [`nook/ROADMAP.md`](./ROADMAP.md) for the current development roadmap and priorities.

- [cloud-init NoCloud DataSource documentation](https://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html)
- [cloud-init Instance Metadata documentation](https://cloudinit.readthedocs.io/en/latest/topics/instancedata.html)
- [NoCloud metadata format example (Canonical)](https://discourse.ubuntu.com/t/nocloud-cloud-init-datasource/15312)
- [cloud-init source code (GitHub)](https://github.com/canonical/cloud-init)
