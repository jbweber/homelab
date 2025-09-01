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
- `/meta-data`, `/user-data`, `/vendor-data` — Other metadata endpoints (IP-based lookup)

**Note:** These endpoints always validate the requestor's IP and return 404 if the machine is not found.

---

## Management Endpoints
These endpoints are for managing the data behind the service. They do **not** perform IP validation or machine matching. They operate on all data in the system, regardless of the requestor.

- `/api/v0/machines` — CRUD operations for machines
- `/api/v0/networks` — Network management
- `/api/v0/ssh-keys` — List all SSH keys in the system (returns all keys as JSON, no IP filtering)

**Note:** These endpoints are for administrative and automation use, not for cloud-init.

---

## Development & Testing Instructions
- Management endpoints should **never** use IP-based machine lookup.
- Cloud-init endpoints **must** use IP-based machine lookup for correct metadata delivery.
- Unit tests expect `/api/v0/ssh-keys` to always return a 200 and a JSON array, even if empty.
- Lint and coverage checks should be run for both endpoint groups.
- **Coverage Status (August 2025):** 75.6% overall, with SSH key handlers at 95.5%+ coverage. Focus on error branches and edge cases in recent improvements.

---

## Updating Documentation
- When adding new endpoints, specify whether they are for cloud-init or management.
- Document any changes to IP validation logic or endpoint behavior.
- Keep this file and the main README up to date with endpoint details and usage notes.

---

For more details, see the main `README.md` and code comments in `nook/internal/api/`.
