# Copilot Instructions for homelab

## Project Overview
This repository manages homelab infrastructure using Ansible. All automation, configuration, and orchestration are handled via playbooks and roles in the `ansible/` directory.

## Layout
- `ansible/` — Main automation folder
  - `site.yml` — Entry point playbook
  - `group_vars/` — Group variable files
  - `host_vars/` — Host-specific variable files
  - `roles/` — Custom and community roles
- `.gitattributes` — Enforces LF line endings except for Windows scripts
- `.gitignore` — Ignores OS/editor files, Python artifacts, Ansible retries, logs, secrets

## Expectations
- All scripts and configuration files use Unix (LF) line endings unless Windows compatibility is required.
- Sensitive files (secrets, keys, etc.) must be ignored via `.gitignore`.
- Ansible playbooks should be idempotent and well-documented.
- Use descriptive names for hosts, groups, and roles.
- Document any custom roles or playbooks in `README.md`.
- When editing the Ansible inventory, always order group names alphabetically and order hosts within each group alphabetically. Each group should be defined only once, and each host should be defined only once per group.
 - Always include `vars_files: [ ~/.homelab.vault ]` in new playbooks to load secrets from the user's vault file outside the repo.
 - For YAML consistency: use double quotes for variable interpolation (e.g., "{{ var }}"), and single quotes for plain strings or strings containing double quotes.

## Go App Development (nook)

- The Go module for nook is located in `nook/` and uses the import path `github.com/jbweber/homelab/nook`.
- All Go code should follow idiomatic Go practices (see `.github/instructions/go.instructions.md`).
- Use Go modules for dependency management.
- Scaffold binaries in `nook/cmd/` and packages in `nook/internal/`.
- Run `go mod tidy` after adding dependencies.
- Use `gofmt` and `goimports` for formatting and import management.
- To run the web service: `go run ./cmd/nook` from the `nook` directory.
- Database: SQLite via `modernc.org/sqlite` for local metadata storage.

### nook services

This service will let us provide the minimal endpoints needed for managing metadata for cloud-init running on our virtual machines. It also possibly could allow us to do other things in the future so we model both the metadata endpoints for cloud-init and the some api endpoints for our service.

#### API Endpoints

* /2021-01-03/dynamic/instance-identity/document
* /2021-01-03/meta-data/public-keys
* /2021-01-03/meta-data/public-keys/<int:idx>
* /2021-01-03/meta-data/public-keys/<int:idx>/openssh-key
* /latest/api/token
* /meta-data
* /user-data
* /vendor-data
* /api/v0/machines
* /api/v0/networks
* /api/v0/ssh-keys
