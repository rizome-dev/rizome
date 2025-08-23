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

### `rizome init`

```bash
# Create a new RIZOME.md interactively from templates
rizome init

# Force overwrite existing RIZOME.md
rizome init --force

# Use a specific template (non-interactive)
rizome init --template my-template-key
```

The `init` command provides an interactive experience for both provider setup and template selection. It first optionally allows you to configure your AI provider registry (enable/disable providers, add custom providers, remove providers, view status) and then provides template selection from saved templates or create a new one on the fly.

By default, all RIZOME.md files and the corresponding sync'd PROVIDER.md files have a Current Date comment pinned to the top; all `rizome` commands update this tag, making it simple to inject context about the current date for research purposes.

### `rizome tmpl`

```bash
# List all available templates
rizome tmpl
rizome tmpl list

# Add a new template interactively
rizome tmpl add
rizome tmpl add "My Template Name"

# Edit an existing template
rizome tmpl edit
rizome tmpl edit "Template Name"

# Show template content
rizome tmpl show
rizome tmpl show "Template Name"

# Delete a template
rizome tmpl delete
rizome tmpl delete "Template Name" --force
```

Templates are stored in `~/.rizome/config.yaml` and can be reused across projects. The template management system allows you to create, edit, and organize reusable RIZOME.md templates.

### `rizome sync`

```bash
# Interactive provider selection and preview changes
rizome sync --dry-run

# Interactive provider selection and apply synchronization
rizome sync

# Force overwrite existing files
rizome sync --force

# Non-interactive mode with specific providers
rizome sync --providers claude,qwen,cursor

# Non-interactive mode (syncs enabled providers from registry)
rizome sync --non-interactive
```

The `sync` command provides an interactive checkbox interface for selecting which providers to sync. Providers enabled in your registry (see `rizome init` provider setup phase) are pre-selected by default. You can override these selections during interactive sync.

This will create/update individual provider configuration files for enabled providers. Default providers include:
- `CLAUDE.md` - Claude Code and Claude API
- `QWEN.md` - Qwen Code and Qwen models  
- `CURSOR.md` - Cursor AI IDE
- `GEMINI.md` - Gemini CLI and Gemini models
- `WINDSURF.md` - Windsurf AI development environment

You can add custom providers or modify these defaults using `rizome init` (provider setup phase).

#### RIZOME.md Format

The `RIZOME.md` file uses a structured format:

##### Required Sections

- **Common Instructions**: Instructions that apply to all AI providers
- **Provider Overrides**: Provider-specific instructions organized by provider name

##### Default Providers

The following providers are included by default (manageable via `rizome init`):

- `CLAUDE` - Claude Code and Claude API
- `QWEN` - Qwen Code and Qwen models
- `CURSOR` - Cursor AI IDE
- `GEMINI` - Gemini CLI and Gemini models
- `WINDSURF` - Windsurf AI development environment

**Note**: You can add custom providers, enable/disable default providers, and manage your provider registry using the `rizome init` command (provider setup phase).

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

### Provider Registry Management

Configure your personal AI provider preferences and manage custom providers:

```bash
# Set up your provider preferences (during init provider setup phase)
rizome init
# When prompted, choose "Yes" to configure provider registry
# Then select from: Manage Provider Settings, Add Custom Provider, 
# Remove Provider, Show Provider Status, or Finish Provider Setup
```

The provider setup phase in `rizome init` lets you:
- Enable or disable providers (affects default selections in `sync`)
- Add custom AI providers with descriptions and categories  
- Remove providers you no longer need
- View comprehensive provider status and settings

All settings are stored in `~/.rizome/config.yaml` and persist across projects.

### Multi-Provider Development

Keep all your AI development environments synchronized with a single source of truth:

```bash
# Initialize a new RIZOME.md from templates (interactive)
rizome init

# Or use a specific template
rizome init --template my-project-template

# Edit RIZOME.md with your project requirements
# Then interactively sync across selected providers
rizome sync
```

### Team Collaboration

Share consistent AI provider configurations across your team:

```bash
# Team lead creates and saves a team template
rizome tmpl add "Team Standards"
# Edit template with coding standards
# Then create project RIZOME.md from template
rizome init --template team-standards
git add RIZOME.md
git commit -m "Add coding standards"
git push

# Team members sync their environments
git pull
rizome sync  # Interactive provider selection
```

### Project Templates

Create and manage reusable project templates with pre-configured AI provider settings:

```bash
# Create and save reusable templates
rizome tmpl add "Go Backend"
rizome tmpl add "Python ML"
rizome tmpl add "Frontend Project"

# Use templates in new projects
rizome init --template go-backend
# or select interactively
rizome init

# List and manage your templates
rizome tmpl list
rizome tmpl show "React Project"
rizome tmpl edit "React Project"

# Share templates across team (stored in ~/.rizome/config.yaml)
scp ~/.rizome/config.yaml teammate@host:~/.rizome/
```

### Template Management Workflow

```bash
# Create templates for different project types
rizome tmpl add "Frontend Project"  # Add React/Vue/Angular specific instructions
rizome tmpl add "Backend API"       # Add API development guidelines  
rizome tmpl add "ML Pipeline"       # Add data science and ML instructions

# Use templates when starting new projects
cd new-frontend-project
rizome init --template frontend-project
rizome sync --providers claude,cursor  # Select specific providers
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
