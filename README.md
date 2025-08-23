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

#### 1. Create a RIZOME.md file

Create a `RIZOME.md` file in your project directory:

```markdown
# RIZOME.md

This is the master configuration file for your project.

## Common Instructions

These are common instructions that apply to all AI providers:

- This is a TypeScript project using React and Next.js
- Follow the existing code patterns and conventions
- Use TypeScript best practices and proper error handling
- Maintain clean, readable, and well-documented code

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

### WINDSURF
Windsurf-specific instructions:
- Prioritize user experience and interface design
- Ensure responsive and accessible components
- Follow modern UI/UX patterns
```

Then:

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

Each file will contain both the common instructions and provider-specific overrides.

The generated files can be used directly with your AI development tools:

- **Claude Code**: Automatically reads `CLAUDE.md`
- **Cursor**: Uses `CURSOR.md` for context
- **Qwen**: References `QWEN.md` 
- **Gemini CLI**: Loads `GEMINI.md`
- **Windsurf**: Uses `WINDSURF.md`

### RIZOME.md Format

The `RIZOME.md` file uses a structured format:

#### Required Sections

- **Common Instructions**: Instructions that apply to all AI providers
- **Provider Overrides**: Provider-specific instructions organized by provider name

#### Supported Providers

- `CLAUDE` - Claude Code and Claude API
- `QWEN` - Qwen Code and Qwen models
- `CURSOR` - Cursor AI IDE
- `GEMINI` - Gemini CLI and Gemini models
- `WINDSURF` - Windsurf AI development environment

#### Example Structure

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

### Available Make Targets

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

Built with ❤️ by [Rizome Labs](https://rizome.dev)
