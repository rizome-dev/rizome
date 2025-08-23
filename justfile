# justfile for github.com/rizome-dev/rizome
# https://just.systems/

# Set shell for Windows compatibility
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# Default recipe to display help
default:
    @just --list

# Go version and Docker image
go_version := "1.24.0"
docker_image := "golang:" + go_version + "-bullseye"
build_dir := "./build"
binary_name := "rizome"
cli_source := "./cmd/rizome"

# Colors for output
bold := '\033[1m'
green := '\033[32m'
yellow := '\033[33m'
red := '\033[31m'
reset := '\033[0m'

# Clean build artifacts
clean:
    @echo "{{yellow}}Cleaning build artifacts...{{reset}}"
    rm -rf {{build_dir}}
    rm -rf vendor/
    @echo "{{green}}✓ Clean complete{{reset}}"

# Vendor dependencies
vendor:
    @echo "{{yellow}}Vendoring dependencies...{{reset}}"
    go mod download
    go mod vendor
    go mod tidy
    @echo "{{green}}✓ Vendor complete{{reset}}"

# Create build directory
_create-build-dir:
    @mkdir -p {{build_dir}}

# Build CLI binary for Darwin/macOS (local)
build-darwin-local: vendor _create-build-dir
    @echo "{{yellow}}Building {{binary_name}} for Darwin (macOS)...{{reset}}"
    @if [ ! -d {{cli_source}} ]; then \
        echo "{{red}}Error: {{cli_source}} not found{{reset}}"; \
        exit 1; \
    fi
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -trimpath \
        -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o {{build_dir}}/{{binary_name}}-darwin-arm64 \
        {{cli_source}}
    @echo "{{green}}✓ Built {{build_dir}}/{{binary_name}}-darwin-arm64{{reset}}"

# Build CLI binary for Linux (local)
build-linux-local: vendor _create-build-dir
    @echo "{{yellow}}Building {{binary_name}} for Linux (local)...{{reset}}"
    @if [ ! -d {{cli_source}} ]; then \
        echo "{{red}}Error: {{cli_source}} not found{{reset}}"; \
        exit 1; \
    fi
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath \
        -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o {{build_dir}}/{{binary_name}}-linux-amd64 \
        {{cli_source}}
    @echo "{{green}}✓ Built {{build_dir}}/{{binary_name}}-linux-amd64{{reset}}"

# Build CLI binary for Linux using Docker (ensures compatibility)
build-linux-docker: vendor _create-build-dir
    @echo "{{yellow}}Building {{binary_name}} for Linux (Docker)...{{reset}}"
    @if [ ! -d {{cli_source}} ]; then \
        echo "{{red}}Error: {{cli_source}} not found{{reset}}"; \
        exit 1; \
    fi
    docker run --rm \
        --platform linux/amd64 \
        -v $(pwd):/workspace \
        -w /workspace \
        {{docker_image}} \
        bash -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -o {{build_dir}}/{{binary_name}}-linux-amd64 {{cli_source}}"
    @echo "{{green}}✓ Built {{build_dir}}/{{binary_name}}-linux-amd64 (via Docker){{reset}}"

# Build all targets
build-all: build-darwin-local build-linux-docker
    @echo "{{green}}✓ All builds complete{{reset}}"
    @ls -la {{build_dir}}/

# Verify CLI binary (macOS)
verify-darwin:
    @echo "{{yellow}}Verifying Darwin binary...{{reset}}"
    @if [ -f {{build_dir}}/{{binary_name}}-darwin-arm64 ]; then \
        file {{build_dir}}/{{binary_name}}-darwin-arm64; \
        {{build_dir}}/{{binary_name}}-darwin-arm64 --help | head -5; \
        echo "{{green}}✓ Binary verification complete{{reset}}"; \
    else \
        echo "{{red}}Error: {{build_dir}}/{{binary_name}}-darwin-arm64 not found{{reset}}"; \
        exit 1; \
    fi

# Verify CLI binary (Linux)
verify-linux:
    @echo "{{yellow}}Verifying Linux binary...{{reset}}"
    @if [ -f {{build_dir}}/{{binary_name}}-linux-amd64 ]; then \
        file {{build_dir}}/{{binary_name}}-linux-amd64; \
        if [ "$(uname -s)" = "Linux" ]; then \
            {{build_dir}}/{{binary_name}}-linux-amd64 --help | head -5; \
        else \
            echo "Cannot test Linux binary on non-Linux system"; \
        fi; \
        echo "{{green}}✓ Binary verification complete{{reset}}"; \
    else \
        echo "{{red}}Error: {{build_dir}}/{{binary_name}}-linux-amd64 not found{{reset}}"; \
        exit 1; \
    fi

# Run tests
test:
    @echo "{{yellow}}Running tests...{{reset}}"
    go test -v -race ./...
    @echo "{{green}}✓ Tests complete{{reset}}"

# Run linter
lint:
    @echo "{{yellow}}Running linter...{{reset}}"
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run ./...; \
    else \
        echo "{{red}}golangci-lint not installed{{reset}}"; \
        echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
        exit 1; \
    fi
    @echo "{{green}}✓ Lint complete{{reset}}"

# Format code
fmt:
    @echo "{{yellow}}Formatting code...{{reset}}"
    go fmt ./...
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    fi
    @echo "{{green}}✓ Format complete{{reset}}"

# Check if Docker is available
check-docker:
    @if ! command -v docker >/dev/null 2>&1; then \
        echo "{{red}}Error: Docker is not installed or not in PATH{{reset}}"; \
        exit 1; \
    fi
    @if ! docker info >/dev/null 2>&1; then \
        echo "{{red}}Error: Docker daemon is not running{{reset}}"; \
        exit 1; \
    fi
    @echo "{{green}}✓ Docker is available{{reset}}"

# Development setup
setup:
    @echo "{{yellow}}Setting up development environment...{{reset}}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go mod download
    @echo "{{green}}✓ Setup complete{{reset}}"

# Show build information
info:
    @echo "{{bold}}Build Information:{{reset}}"
    @echo "  Go Version:     {{go_version}}"
    @echo "  Docker Image:   {{docker_image}}"
    @echo "  Build Dir:      {{build_dir}}"
    @echo "  Binary Name:    {{binary_name}}"
    @echo "  CLI Source:     {{cli_source}}"
    @echo ""
    @echo "{{bold}}System Information:{{reset}}"
    @echo "  OS:             $(uname -s)"
    @echo "  Arch:           $(uname -m)"
    @echo "  Go:             $(go version)"

# CI/CD pipeline (runs all checks)
ci: vendor lint test
    @echo "{{green}}✓ CI checks passed{{reset}}"

# Quick build for current platform
build: vendor _create-build-dir
    @echo "{{yellow}}Building {{binary_name}} for current platform...{{reset}}"
    go build \
        -trimpath \
        -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o {{build_dir}}/{{binary_name}} \
        {{cli_source}}
    @echo "{{green}}✓ Built {{build_dir}}/{{binary_name}}{{reset}}"

# Interactive shell in Docker build environment
shell: check-docker
    docker run --rm -it \
        --platform linux/amd64 \
        -v $(pwd):/workspace \
        -w /workspace \
        {{docker_image}} \
        bash

# Install the binary locally
install: build
    @echo "{{yellow}}Installing {{binary_name}} to /usr/local/bin...{{reset}}"
    sudo cp {{build_dir}}/{{binary_name}} /usr/local/bin/{{binary_name}}
    @echo "{{green}}✓ {{binary_name}} installed to /usr/local/bin/{{binary_name}}{{reset}}"

# Run the sync command
sync: build
    @echo "{{yellow}}Running rizome sync...{{reset}}"
    {{build_dir}}/{{binary_name}} sync
    @echo "{{green}}✓ Sync completed{{reset}}"