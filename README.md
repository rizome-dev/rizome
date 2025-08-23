# Rizome CLI - Agentic Development Environment Workspace Management

[![GoDoc](https://pkg.go.dev/badge/github.com/rizome-dev/rizome)](https://pkg.go.dev/github.com/rizome-dev/rizome)
[![Go Report Card](https://goreportcard.com/badge/github.com/rizome-dev/rizome)](https://goreportcard.com/report/github.com/rizome-dev/rizome)
[![CI](https://github.com/rizome-dev/rizome/actions/workflows/ci.yml/badge.svg)](https://github.com/rizome-dev/rizome/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-GPL--2.0-blue.svg)](LICENSE)

**built by:** [rizome labs](https://rizome.dev) | **contact:** [hi@rizome.dev](mailto:hi@rizome.dev)

## Installation

### Homebrew

```bash
brew tap rizome-dev/brews && brew install rizome
```

### Source

```bash
git clone https://github.com/rizome-dev/rizome && cd rizome && sudo make install
```

### Binary Download

Download the latest binary for your platform from the [releases page](https://github.com/rizome-dev/rizome/releases).

## Commands

### `rizome sync`

```bash
# Preview what will be changed
rizome sync --dry-run

# Apply the synchronization
rizome sync

# Force overwrite existing files
rizome sync --force
```

This will create/update individual provider configuration files:
- `CLAUDE.md`
- `QWEN.md`
- `CURSOR.md`
- `GEMINI.md`
- `WINDSURF.md`

#### RIZOME.md Format

The `RIZOME.md` file uses a structured format:

##### Required Sections

- **Common Instructions**: Instructions that apply to all AI providers
- **Provider Overrides**: Provider-specific instructions organized by provider name

##### Supported Providers

- `CLAUDE` - Claude Code and Claude API
- `QWEN` - Qwen Code and Qwen models
- `CURSOR` - Cursor AI IDE
- `GEMINI` - Gemini CLI and Gemini models
- `WINDSURF` - Windsurf AI development environment

##### Example Structure

```markdown
# RIZOME.md

Project overview and context.

## Common Instructions

Instructions that apply to all providers:
- Project type and technology stack
- Coding standards and conventions
- General best practices

## Provider Overrides

### CLAUDE
Claude-specific instructions

### QWEN
Qwen-specific instructions

### CURSOR
Cursor-specific instructions

### GEMINI
Gemini-specific instructions

### WINDSURF
Windsurf-specific instructions
```

## Use Cases

### Multi-Provider Development

Keep all your AI development environments synchronized with a single source of truth:

```bash
# Update your RIZOME.md with new project requirements
# Then sync across all providers
rizome sync
```

### Team Collaboration

Share consistent AI provider configurations across your team:

```bash
# Team lead updates RIZOME.md with coding standards
git add RIZOME.md
git commit -m "Update coding standards"
git push

# Team members sync their environments
git pull
rizome sync
```

### Project Templates

Create reusable project templates with pre-configured AI provider settings:

```bash
# In your project template repository
echo "Standard RIZOME.md configuration" > RIZOME.md
rizome sync

# Users of your template get consistent AI behavior
git clone your-template
cd your-template
rizome sync
```

---

```bash
make help              # Show all available targets
make build             # Build the binary
make test              # Run unit tests
make test-coverage     # Run tests with coverage report
make test-sync         # Test sync command functionality
make lint              # Run linters
make fmt               # Format code
make install           # Install to /usr/local/bin
make uninstall         # Remove installed binary
make clean             # Remove build artifacts
```

---

Built with ❤️ by [Rizome Labs](https://rizome.dev)
