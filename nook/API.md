# Nook API Documentation

## Overview
Nook provides two distinct sets of API endpoints:
- **Cloud-init metadata endpoints** (for VM bootstrapping)
- **Management endpoints** (for managing metadata and keys)

---

## Cloud-init Metadata Endpoints
These endpoints are compatible with cloud-init and EC2-style metadata services. They use the requestor's IP address to look up the associated machine and return metadata specific to that machine.

- `/2021-01-03/dynamic/instance-identity/document` — EC2-compatible instance identity document (IP-based lookup)
- `/2021-01-03/meta-data/public-keys` — List of public keys for the requesting machine (IP-based lookup)
- `/2021-01-03/meta-data/public-keys/{idx}` — Specific public key by index (IP-based lookup)
- `/2021-01-03/meta-data/public-keys/{idx}/openssh-key` — OpenSSH-formatted key by index (IP-based lookup)
- `/meta-data` — Instance metadata (YAML with hostname, instance-id, etc.) (IP-based lookup)
- `/user-data` — Dynamic cloud-config with SSH keys and hostname (IP-based lookup)
- `/vendor-data` — Vendor-specific data (currently empty) (IP-based lookup)

**Note:** These endpoints always validate the requestor's IP and return 404 if the machine is not found.

---

## Management Endpoints
These endpoints are for managing the data behind the service. They do **not** perform IP validation or machine matching. They operate on all data in the system, regardless of the requestor.

- `GET /api/v0/machines` — List all machines
- `POST /api/v0/machines` — Create a new machine
- `GET /api/v0/machines/{id}` — Get machine by ID
- `DELETE /api/v0/machines/{id}` — Delete machine by ID (cascades to SSH keys)
- `GET /api/v0/machines/name/{name}` — Get machine by name
- `GET /api/v0/machines/ipv4/{ipv4}` — Get machine by IPv4

- `GET /api/v0/ssh-keys` — List all SSH keys
- `POST /api/v0/ssh-keys` — Create a new SSH key
- `GET /api/v0/ssh-keys/{id}` — Get SSH key by ID
- `DELETE /api/v0/ssh-keys/{id}` — Delete SSH key by ID

- `GET /api/v0/networks` — List all networks
- `POST /api/v0/networks` — Create a new network
- `GET /api/v0/networks/{id}` — Get network by ID
- `DELETE /api/v0/networks/{id}` — Delete network by ID

**Note:** These endpoints are for administrative and automation use, not for cloud-init.

---

## Development & Testing Instructions
- Management endpoints should **never** use IP-based machine lookup.
- Cloud-init endpoints **must** use IP-based machine lookup for correct metadata delivery.
- Unit tests expect `/api/v0/ssh-keys` to always return a 200 and a JSON array, even if empty.
- Lint and coverage checks should be run for both endpoint groups.
- **Coverage Status (September 2025):** 58.4% overall coverage (temporarily lowered from 75% during active development). All critical paths tested, focus on functionality over metrics. SSH key handlers at 95.5%+ coverage. Recent improvements include comprehensive error handling and cascade deletion testing.

---

## Updating Documentation
- When adding new endpoints, specify whether they are for cloud-init or management.
- Document any changes to IP validation logic or endpoint behavior.
- Keep this file and the main README up to date with endpoint details and usage notes.

---

For more details, see the main `README.md` and code comments in `nook/internal/api/`.
