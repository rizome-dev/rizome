# Project variables
PROJECT_NAME := rizome
BINARY_NAME := rizome
GO_FILES := $(shell find . -name '*.go' -type f -not -path "./vendor/*")
MAIN_PACKAGE := ./cmd/rizome

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.buildTime=$(BUILD_TIME)'"

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build output
BUILD_DIR := build
COVERAGE_FILE := coverage.out

# Colors for terminal output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_RED := \033[31m

# Default target
.DEFAULT_GOAL := help

.PHONY: all
all: clean lint test build ## Run all main targets (clean, lint, test, build)

.PHONY: help
help: ## Display this help message
	@echo "$(COLOR_BOLD)$(PROJECT_NAME) Makefile$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET) make $(COLOR_GREEN)[target]$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Targets:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-cross
build-cross: ## Build binaries for multiple platforms
	@echo "$(COLOR_BLUE)Building cross-platform binaries...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -buildvcs=false $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Cross-platform binaries built$(COLOR_RESET)"

.PHONY: install
install: build ## Build and install the binary to /usr/local/bin
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(COLOR_RESET)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@# Clear ALL extended attributes and quarantine flags (macOS specific)
	@sudo xattr -d com.apple.quarantine /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@sudo xattr -d com.apple.provenance /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@sudo xattr -cr /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@# Re-sign the binary (macOS specific)
	@sudo codesign --force --deep --sign - /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@echo "$(COLOR_GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(COLOR_RESET)"
	@$(MAKE) post-install

.PHONY: dev
dev: ## Quick development build (no optimization)
	@CGO_ENABLED=0 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Development build complete$(COLOR_RESET)"

.PHONY: run
run: ## Build and run the application
	@CGO_ENABLED=0 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

.PHONY: clean
clean: ## Remove build artifacts
	@echo "$(COLOR_YELLOW)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f *.test
	@rm -f *.out
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

.PHONY: test
test: ## Run unit tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GOTEST) -race -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

.PHONY: test-short
test-short: ## Run unit tests (short mode)
	@echo "$(COLOR_BLUE)Running short tests...$(COLOR_RESET)"
	@$(GOTEST) -short -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	@$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

.PHONY: test-sync
test-sync: build ## Test the sync command functionality
	@echo "$(COLOR_BLUE)Testing sync command...$(COLOR_RESET)"
	@echo "Testing dry-run..."
	@./$(BUILD_DIR)/$(BINARY_NAME) sync --dry-run
	@echo "Testing actual sync..."
	@./$(BUILD_DIR)/$(BINARY_NAME) sync
	@echo "$(COLOR_GREEN)✓ Sync command tests passed$(COLOR_RESET)"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	@$(GOTEST) -bench=. -benchmem ./...

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GOFMT) ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@$(GOVET) ./...
	@echo "$(COLOR_GREEN)✓ Vet passed$(COLOR_RESET)"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(COLOR_BLUE)Running linters...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		if golangci-lint run ./...; then \
			echo "$(COLOR_GREEN)✓ Linting passed$(COLOR_RESET)"; \
		else \
			echo "$(COLOR_RED)✗ Linting failed$(COLOR_RESET)"; \
			exit 1; \
		fi \
	elif [ -x "$(HOME)/go/bin/golangci-lint" ]; then \
		if $(HOME)/go/bin/golangci-lint run ./...; then \
			echo "$(COLOR_GREEN)✓ Linting passed$(COLOR_RESET)"; \
		else \
			echo "$(COLOR_RED)✗ Linting failed$(COLOR_RESET)"; \
			exit 1; \
		fi \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Run 'make setup' to install.$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Hint: If you've run 'make setup', ensure $(HOME)/go/bin is in your PATH$(COLOR_RESET)"; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "$(COLOR_BLUE)Running linters with auto-fix...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./... && \
		echo "$(COLOR_GREEN)✓ Auto-fix applied$(COLOR_RESET)"; \
	elif [ -x "$(HOME)/go/bin/golangci-lint" ]; then \
		$(HOME)/go/bin/golangci-lint run --fix ./... && \
		echo "$(COLOR_GREEN)✓ Auto-fix applied$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Run 'make setup' to install.$(COLOR_RESET)"; \
	fi

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@$(GOMOD) download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: vendor
vendor: ## Create vendor directory
	@echo "$(COLOR_BLUE)Vendoring dependencies...$(COLOR_RESET)"
	@$(GOMOD) vendor
	@echo "$(COLOR_GREEN)✓ Dependencies vendored$(COLOR_RESET)"

.PHONY: setup
setup: ## Install development tools
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Note: Installing tools with current Go version $(shell go version | awk '{print $$3}')$(COLOR_RESET)"
	@$(GOCMD) install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GOCMD) install github.com/goreleaser/goreleaser@latest
	@echo "$(COLOR_GREEN)✓ Development tools installed$(COLOR_RESET)"

.PHONY: clean-tools
clean-tools: ## Remove installed development tools
	@echo "$(COLOR_YELLOW)Removing development tools...$(COLOR_RESET)"
	@rm -f $(HOME)/go/bin/golangci-lint
	@rm -f $(HOME)/go/bin/gosec
	@rm -f $(HOME)/go/bin/goreleaser
	@echo "$(COLOR_GREEN)✓ Development tools removed$(COLOR_RESET)"

.PHONY: deep-clean
deep-clean: clean clean-tools ## Deep clean including Go cache and installed tools
	@echo "$(COLOR_BLUE)Performing deep clean...$(COLOR_RESET)"
	@$(GOCMD) clean -cache
	@$(GOCMD) clean -testcache
	@echo "$(COLOR_GREEN)✓ Go build and test caches cleaned$(COLOR_RESET)"

.PHONY: generate
generate: ## Run go generate
	@echo "$(COLOR_BLUE)Running go generate...$(COLOR_RESET)"
	@$(GOCMD) generate ./...
	@echo "$(COLOR_GREEN)✓ Generation complete$(COLOR_RESET)"

.PHONY: release-dry-run
release-dry-run: ## Perform a dry run of goreleaser
	@echo "$(COLOR_BLUE)Running release dry run...$(COLOR_RESET)"
	@goreleaser release --snapshot --skip-publish --rm-dist

.PHONY: release
release: ## Create a new release (requires tag)
	@echo "$(COLOR_BLUE)Creating release...$(COLOR_RESET)"
	@goreleaser release --rm-dist

.PHONY: post-install
post-install: ## Post-installation setup
	@echo ""
	@echo "$(COLOR_GREEN)✨ Rizome CLI installation complete!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Quick Start:$(COLOR_RESET)"
	@echo "  1. Run '$(COLOR_GREEN)rizome init$(COLOR_RESET)' to interactively create a RIZOME.md from templates"
	@echo "  2. Edit RIZOME.md with your project details and provider overrides"
	@echo "  3. Run '$(COLOR_GREEN)rizome sync$(COLOR_RESET)' to interactively synchronize provider configurations"
	@echo ""
	@echo "$(COLOR_BOLD)Template Management:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)rizome tmpl$(COLOR_RESET)        List available templates"
	@echo "  $(COLOR_GREEN)rizome tmpl add$(COLOR_RESET)    Create a new template interactively"
	@echo "  $(COLOR_GREEN)rizome tmpl edit$(COLOR_RESET)   Edit an existing template"
	@echo "  $(COLOR_GREEN)rizome tmpl show$(COLOR_RESET)   Show template content"
	@echo "  $(COLOR_GREEN)rizome tmpl delete$(COLOR_RESET) Delete a template"
	@echo ""
	@echo "$(COLOR_BOLD)Example RIZOME.md format:$(COLOR_RESET)"
	@echo "  # RIZOME.md"
	@echo ""
	@echo "  Project overview and context."
	@echo ""
	@echo "  ## Common Instructions"
	@echo "  Instructions that apply to all AI providers:"
	@echo "  - Project type and technology stack"
	@echo "  - Coding standards and conventions"
	@echo ""
	@echo "  ## Provider Overrides"
	@echo "  ### CLAUDE"
	@echo "  Claude-specific instructions"
	@echo ""
	@echo "  ### QWEN"
	@echo "  Qwen-specific instructions"
	@echo ""
	@echo "$(COLOR_BOLD)Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)rizome --help$(COLOR_RESET)        Show help information"
	@echo "  $(COLOR_GREEN)rizome init$(COLOR_RESET)          Interactive RIZOME.md creation from templates"
	@echo "  $(COLOR_GREEN)rizome tmpl$(COLOR_RESET)          Manage RIZOME.md templates"
	@echo "  $(COLOR_GREEN)rizome sync$(COLOR_RESET)          Interactive provider configuration sync"
	@echo "  $(COLOR_GREEN)rizome sync --dry-run$(COLOR_RESET)    Preview changes without applying"
	@echo ""
	@echo "$(COLOR_BLUE)For more information: https://github.com/rizome-dev/rizome$(COLOR_RESET)"

.PHONY: uninstall
uninstall: ## Remove the installed binary
	@echo "$(COLOR_YELLOW)Uninstalling $(BINARY_NAME)...$(COLOR_RESET)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ $(BINARY_NAME) uninstalled$(COLOR_RESET)"

# CI/CD pipeline (runs all checks)
.PHONY: ci
ci: vendor lint test
	@echo "$(COLOR_GREEN)✓ CI checks passed$(COLOR_RESET)"