# Makefile for Nook project

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOIMPORTS=$(HOME)/go/bin/goimports

# Main package path
MAIN_PACKAGE=./cmd/nook
BINARY_NAME=nook
BINARY_UNIX=$(BINARY_NAME)_unix

# Coverage parameters
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
COVERAGE_THRESHOLD=55

# Build the project
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

# Build for Linux
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PACKAGE)

# Test
.PHONY: test
test:
	$(GOTEST) -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Test with coverage and validate threshold
.PHONY: test-coverage-validate
test-coverage-validate:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ "$$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l)" -eq 1 ]; then \
		echo "ERROR: Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "SUCCESS: Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

# Show coverage in terminal
.PHONY: coverage-func
coverage-func:
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GOTEST) -v -race ./...

# Clean
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Run
.PHONY: run
run:
	$(GOCMD) run $(MAIN_PACKAGE)

# Dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy


# Format code (go fmt + goimports)
.PHONY: fmt
fmt:
	$(GOFMT) ./...
	$(GOIMPORTS) -w .

# Vet code
.PHONY: vet
vet:
	$(GOVET) ./...


# Lint (uses explicit path for portability)
.PHONY: lint
lint:
	$(HOME)/go/bin/golangci-lint run


# Check formatting (go fmt + goimports)
.PHONY: check-fmt
check-fmt:
	@if [ -n "$$($(GOIMPORTS) -l .)" ]; then \
		echo "The following files are not formatted by goimports:"; \
		$(GOIMPORTS) -l .; \
		exit 1; \
	fi

# Security check (requires gosec to be installed)
.PHONY: security
security:
	gosec ./...

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Development environment setup complete"

# CI pipeline simulation
.PHONY: ci
ci: check-fmt vet test-race test-coverage-validate


# Install development tools
.PHONY: install-tools
install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(HOME)/go/bin v2.4.0
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
# Static analysis (staticcheck)
.PHONY: staticcheck
staticcheck:
	staticcheck ./...


# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build                 - Build the binary"
	@echo "  build-linux           - Build for Linux"
	@echo "  test                  - Run tests"
	@echo "  test-coverage         - Run tests with coverage report"
	@echo "  test-coverage-validate - Run tests with coverage validation (threshold: $(COVERAGE_THRESHOLD)%)"
	@echo "  coverage-func         - Show coverage in terminal"
	@echo "  test-race             - Run tests with race detection"
	@echo "  clean                 - Clean build artifacts"
	@echo "  run                   - Run the application"
	@echo "  deps                  - Download and tidy dependencies"
	@echo "  fmt                   - Format code (go fmt + goimports)"
	@echo "  vet                   - Vet code"
	@echo "  lint                  - Lint code (requires golangci-lint)"
	@echo "  staticcheck           - Static analysis (requires staticcheck)"
	@echo "  check-fmt             - Check code formatting (go fmt + goimports)"
	@echo "  security              - Security check (requires gosec)"
	@echo "  dev-setup             - Setup development environment"
	@echo "  ci                    - Run CI pipeline (fmt check, vet, race test, coverage)"
	@echo "  install-tools         - Install development tools (golangci-lint, gosec, goimports, staticcheck)"
	@echo "  help                  - Show this help"
