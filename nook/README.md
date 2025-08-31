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

# Run tests with coverage report
make test-coverage

# Run tests with coverage validation (requires 80% coverage)
make test-coverage-validate

# Run tests with race detection
make test-race

# Show coverage in terminal
make coverage-func
```

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

- `/2021-01-03/dynamic/instance-identity/document`
- `/2021-01-03/meta-data/public-keys`
- `/2021-01-03/meta-data/public-keys/<int:idx>`
- `/2021-01-03/meta-data/public-keys/<int:idx>/openssh-key`
- `/latest/api/token`
- `/meta-data`
- `/user-data`
- `/vendor-data`
- `/api/v0/machines`
- `/api/v0/networks`
- `/api/v0/ssh-keys`

## Database

The service uses SQLite for local metadata storage via `modernc.org/sqlite`.

## Makefile Targets

Run `make help` to see all available targets and their descriptions.
