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
