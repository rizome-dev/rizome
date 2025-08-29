<!-- Current Date: 2025-08-29 14:39:08 UTC -->

# RIZOME.md

Rizome is a Go CLI tool for managing AI provider configuration files across multiple development environments. It synchronizes a master configuration (RIZOME.md) to individual provider-specific files (CLAUDE.md, QWEN.md, CURSOR.md, etc.) to maintain consistent AI coding assistant behavior.

## Common Instructions

### Development Commands

```bash
make build              # Build the binary to build/ directory
make dev                # Quick development build (no optimization)
make install            # Build and install to /usr/local/bin with proper permissions
make build-cross        # Build for multiple platforms (darwin, linux, windows)
```

```bash
make test               # Run unit tests with race detection
make test-short         # Run short tests only
make test-coverage      # Generate test coverage report (creates coverage.html)
make test-sync          # Test the sync command functionality
```

```bash
make lint               # Run golangci-lint (extensive linting rules configured)
make lint-fix           # Run golangci-lint with auto-fix
make fmt                # Format code with gofmt
make vet                # Run go vet
make check              # Run all checks (fmt, vet, lint, test)
```

```bash
make setup              # Install development tools (golangci-lint, gosec, goreleaser)
make deps               # Download dependencies
make vendor             # Create vendor directory
```

```bash
go test -v -run TestName ./internal/package_name/
```

### Architecture Overview

#### Command Structure
- **Main Entry**: `cmd/rizome/main.go` - Uses charmbracelet/fang for enhanced CLI with graceful shutdown
- **Root Command**: `internal/cli/root.go` - Cobra-based command structure with custom help templates
- **Core Commands**:
  - `init`: Interactive RIZOME.md creation with provider setup and template selection
  - `sync`: Synchronize RIZOME.md to provider-specific files (CLAUDE.md, QWEN.md, etc.)
  - `tmpl`: Template management (list, add, edit, show, delete)

#### Key Packages
- **internal/cli**: Command implementations using Cobra
- **internal/config**: Template and provider registry management (stored in ~/.rizome/config.yaml)
- **internal/sync**: Core synchronization logic between RIZOME.md and provider files
- **internal/tui**: Terminal UI components using Charmbracelet Bubbletea (checkbox, list, multiline input)
- **internal/version**: Version information management

#### Provider System
- **Provider Registry**: Configurable via `rizome init`, stored in ~/.rizome/config.yaml
- **Default Providers**: CLAUDE, QWEN, CURSOR, GEMINI, WINDSURF
- **Custom Providers**: Users can add custom AI providers with descriptions and categories
- **Sync Selection**: Interactive checkbox UI for selecting which providers to sync

#### RIZOME.md Structure
The master configuration file follows this format:
1. **Common Instructions**: Applied to all AI providers
2. **Provider Overrides**: Provider-specific instructions (e.g., ### CLAUDE, ### QWEN)
3. **Date Timestamp**: Automatically updated with current date on each rizome command

#### Configuration Storage
- **Location**: `~/.rizome/config.yaml`
- **Contents**: Templates, provider registry, enabled/disabled providers
- **Persistence**: Settings persist across projects and sessions

### Key Implementation Details

- **Go Version**: 1.24.0+ required
- **No CGO**: Built with CGO_ENABLED=0 for maximum portability
- **Error Handling**: Comprehensive error handling with context-appropriate messages
- **Signal Handling**: Graceful shutdown on SIGINT/SIGTERM
- **TUI Framework**: Charmbracelet Bubbletea for interactive elements
- **Config Management**: Viper for configuration with YAML format
- **Testing**: Unit tests with race detection, integration tests for CLI commands

### Important Notes from RIZOME.md

This project's CLAUDE.md is managed by a RIZOME.md sync file. The RIZOME.md contains:
- Common instructions for all AI providers
- Provider-specific overrides for CLAUDE, QWEN, CURSOR, GEMINI, and WINDSURF
- Focus on clean architecture, separation of concerns, and proper dependency injection
- Emphasis on Go best practices and proper error handling
- Use existing code patterns from the opun project as reference

## Provider Overrides

### CLAUDE
Claude-specific instructions:
- Focus on clean architecture and separation of concerns
- Use proper dependency injection patterns
- Ensure comprehensive error handling

### QWEN
Qwen-specific instructions:
- Pay attention to performance optimizations
- Use efficient algorithms and data structures
- Consider memory usage in implementations

### CURSOR  
Cursor-specific instructions:
- Emphasize code readability and maintainability
- Provide clear inline documentation
- Use descriptive variable and function names

### GEMINI
Gemini-specific instructions:
- Focus on modularity and reusability
- Implement proper testing strategies
- Consider edge cases in implementations
