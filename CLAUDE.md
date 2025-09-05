# Claude Code Instructions for homelab

This file provides context and instructions for Claude Code users working on this homelab infrastructure project.

## Quick Start
- Read `.github/copilot-instructions.md` for comprehensive project guidelines
- Main automation is in `ansible/` directory with `site.yml` as entry point
- Go service "nook" is in `nook/` directory with its own Makefile

## Key Commands
- **Ansible**: Use playbooks in `ansible/` directory
- **Go (nook)**: 
  - `cd nook && make build` - build the binary
  - `cd nook && make test` - run tests
  - `cd nook && make ci` - run full CI pipeline
  - `cd nook && go run ./cmd/nook` - run the web service

## Testing & Linting
- **Go**: Use `make ci` in `nook/` directory for full validation
- **Integration tests**: Run `nook/test_api.sh` for API testing

## Important Notes
- Always include `vars_files: [ ~/.homelab.vault ]` in new Ansible playbooks
- Maintain alphabetical ordering in Ansible inventory files
- Use Unix (LF) line endings for all files except Windows-specific scripts
- Follow Go module path: `github.com/jbweber/homelab/nook`

## Project Structure
```
├── ansible/           # Main automation (playbooks, roles, inventory)
├── nook/             # Go service for VM metadata management
├── .github/          # GitHub workflows and copilot instructions
└── CLAUDE.md         # This file
```

See `.github/copilot-instructions.md` for complete project guidelines and conventions.