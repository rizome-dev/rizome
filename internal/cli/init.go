package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/spf13/cobra"

	"github.com/rizome-dev/rizome/internal/config"
	"github.com/rizome-dev/rizome/internal/sync"
	"github.com/rizome-dev/rizome/internal/tui"
)

// generateDefaultTemplate creates a default RIZOME.md template based on available providers
func generateDefaultTemplate() string {
	providers := sync.GetAvailableProviders()
	
	var parts []string
	parts = append(parts, "# RIZOME.md")
	parts = append(parts, "")
	parts = append(parts, "Project overview and context.")
	parts = append(parts, "")
	parts = append(parts, "## Common Instructions")
	parts = append(parts, "")
	parts = append(parts, "Instructions that apply to all providers:")
	parts = append(parts, "- Project type and technology stack")
	parts = append(parts, "- Coding standards and conventions")
	parts = append(parts, "- General best practices")
	parts = append(parts, "")
	parts = append(parts, "## Provider Overrides")
	parts = append(parts, "")
	
	for _, provider := range providers {
		parts = append(parts, "### "+provider)
		parts = append(parts, fmt.Sprintf("%s-specific instructions", provider))
		parts = append(parts, "")
	}
	
	// Remove trailing newline
	content := strings.Join(parts, "\n")
	return strings.TrimSuffix(content, "\n")
}

// InitCmd creates the init command
func InitCmd() *cobra.Command {
	var force bool
	var templateKey string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive RIZOME.md creation from templates",
		Long: `The init command interactively creates a new RIZOME.md file in the current directory.
You can select from available templates, create new templates on the fly, or use existing ones.
Templates are stored in ~/.rizome/config.yaml and managed with the 'rizome tmpl' command.

The command provides a beautiful interactive interface for template selection and preview.
If RIZOME.md already exists, use --force to overwrite it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitInteractive(force, templateKey)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing RIZOME.md file")
	cmd.Flags().StringVar(&templateKey, "template", "", "Use specific template by key (non-interactive)")

	return cmd
}

// templateSelectionItem represents a template in the selection list
type templateSelectionItem struct {
	key       string
	template  config.Template
	isDefault bool
}

func (t templateSelectionItem) FilterValue() string { return t.template.Name }
func (t templateSelectionItem) Title() string {
	if t.isDefault {
		return t.template.Name + " (default)"
	}
	return t.template.Name
}
func (t templateSelectionItem) Description() string { return t.template.Description }

// runInitInteractive runs the interactive init command
func runInitInteractive(force bool, templateKey string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	rizomePath := filepath.Join(cwd, "RIZOME.md")

	// Check if RIZOME.md already exists
	if _, err := os.Stat(rizomePath); err == nil && !force {
		overwrite, err := tui.Confirm("RIZOME.md already exists. Overwrite?")
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Printf("%s Operation cancelled\n", infoStyle.Render("ℹ"))
			return nil
		}
	}

	// Provider registry setup before template selection (only in interactive mode)
	if templateKey == "" {
		setupProviders, err := tui.Confirm("Would you like to configure your AI provider registry before creating RIZOME.md?")
		if err != nil {
			return err
		}
		
		if setupProviders {
			if err := runProviderSetupFlow(); err != nil {
				fmt.Printf("%s Provider setup failed: %v\n", infoStyle.Render("ℹ"), err)
				fmt.Println("Continuing with template selection...")
			}
		}
	}

	var selectedTemplate *config.Template

	// If template key is specified, use it directly (non-interactive mode)
	if templateKey != "" {
		tm, err := config.NewTemplateManager()
		if err != nil {
			// Fallback to dynamically generated template if template manager fails
			fmt.Printf("%s Template manager not available, using default template\n", infoStyle.Render("ℹ"))
			selectedTemplate = &config.Template{
				Name:    "Default Template",
				Content: generateDefaultTemplate(),
			}
		} else {
			template, err := tm.GetTemplate(templateKey)
			if err != nil {
				return fmt.Errorf("template '%s' not found: %w", templateKey, err)
			}
			selectedTemplate = template
		}
	} else {
		// Interactive template selection
		template, err := selectTemplateForInit()
		if err != nil {
			return err
		}
		selectedTemplate = template
	}

	fmt.Printf("%s Initializing RIZOME.md in %s\n", infoStyle.Render("ℹ"), cwd)
	fmt.Printf("  Using template: %s\n", selectedTemplate.Name)

	// Inject timestamp for model grounding and write the template file
	contentWithTimestamp := config.InjectTimestamp(selectedTemplate.Content)
	if err := os.WriteFile(rizomePath, []byte(contentWithTimestamp), 0644); err != nil {
		return fmt.Errorf("%s Failed to create RIZOME.md: %w", errorStyle.Render("✗"), err)
	}

	fmt.Printf("\n%s RIZOME.md created successfully!\n", successStyle.Render("✅"))

	// Ask about .claudedocs initialization
	initClaudedocs, err := tui.Confirm("Would you like to initialize .claudedocs for project documentation?")
	if err != nil {
		// Don't fail the whole init if this errors, just skip it
		fmt.Printf("%s Skipping .claudedocs initialization\n", infoStyle.Render("ℹ"))
	} else if initClaudedocs {
		if err := initializeClaudedocs(cwd); err != nil {
			fmt.Printf("%s Failed to initialize .claudedocs: %v\n", errorStyle.Render("✗"), err)
			fmt.Println("You can manually create the .claudedocs folder later")
		} else {
			fmt.Printf("%s .claudedocs initialized successfully!\n", successStyle.Render("✅"))
		}
	}

	fmt.Printf("\n%s Next steps:\n", infoStyle.Render("ℹ"))
	fmt.Printf("  1. Edit RIZOME.md with your project details\n")
	fmt.Printf("  2. Run 'rizome sync' to generate provider-specific files\n")
	fmt.Printf("  3. Use 'rizome tmpl' to manage templates\n")
	if initClaudedocs {
		fmt.Printf("  4. Update .claudedocs/IMPLEMENTATION_PLAN.md with your project plan\n")
	}

	return nil
}

// selectTemplateForInit shows an interactive template selection for init
func selectTemplateForInit() (*config.Template, error) {
	tm, err := config.NewTemplateManager()
	if err != nil {
		// Fallback to dynamically generated template if template manager fails
		fmt.Printf("%s Template manager not available, using default template\n", infoStyle.Render("ℹ"))
		return &config.Template{
			Name:    "Default Template",
			Content: generateDefaultTemplate(),
		}, nil
	}

	templates, err := tm.ListTemplates()
	if err != nil {
		// Fallback to dynamically generated template
		fmt.Printf("%s Failed to load templates, using default template\n", infoStyle.Render("ℹ"))
		return &config.Template{
			Name:    "Default Template",
			Content: generateDefaultTemplate(),
		}, nil
	}

	if len(templates) == 0 {
		// Fallback to dynamically generated template
		fmt.Printf("%s No templates available, using default template\n", infoStyle.Render("ℹ"))
		return &config.Template{
			Name:    "Default Template",
			Content: generateDefaultTemplate(),
		}, nil
	}

	// Convert templates to list items
	var items []list.Item

	// Sort templates by name for consistent display
	var keys []string
	for key := range templates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		template := templates[key]
		item := templateSelectionItem{
			key:       key,
			template:  template,
			isDefault: key == "default",
		}
		items = append(items, item)
	}

	// Add option to create new template
	newTemplateItem := templateSelectionItem{
		key: "new",
		template: config.Template{
			Name:        "Create New Template",
			Description: "Create a new template interactively",
		},
		isDefault: false,
	}
	items = append(items, newTemplateItem)

	selectedItem, err := tui.ListSelection("🚀 Welcome to Rizome! Select a template for your RIZOME.md:", items)
	if err != nil {
		return nil, err
	}

	if selectedTemplateItem, ok := selectedItem.(templateSelectionItem); ok {
		if selectedTemplateItem.key == "new" {
			// Create new template interactively
			return createNewTemplateInteractive()
		}

		// Show template preview
		fmt.Printf("\n%s Template Preview: %s\n", infoStyle.Render("ℹ"), selectedTemplateItem.template.Name)
		if selectedTemplateItem.template.Description != "" {
			fmt.Printf("Description: %s\n", selectedTemplateItem.template.Description)
		}
		fmt.Printf("\nContent preview:\n")
		fmt.Printf("────────────────────────────────────────\n")

		// Show first few lines of content
		lines := strings.Split(selectedTemplateItem.template.Content, "\n")
		maxLines := 10
		if len(lines) > maxLines {
			for i := 0; i < maxLines; i++ {
				fmt.Printf("%s\n", lines[i])
			}
			fmt.Printf("... (%d more lines)\n", len(lines)-maxLines)
		} else {
			fmt.Printf("%s\n", selectedTemplateItem.template.Content)
		}
		fmt.Printf("────────────────────────────────────────\n")

		// Confirm selection
		confirmed, err := tui.Confirm("Use this template?")
		if err != nil {
			return nil, err
		}

		if !confirmed {
			fmt.Printf("%s Template selection cancelled\n", infoStyle.Render("ℹ"))
			return selectTemplateForInit() // Try again
		}

		return &selectedTemplateItem.template, nil
	}

	return nil, fmt.Errorf("no template selected")
}

// createNewTemplateInteractive creates a new template interactively during init
func createNewTemplateInteractive() (*config.Template, error) {
	fmt.Printf("\n%s Creating a new template\n", infoStyle.Render("ℹ"))

	// Get template name
	name, err := tui.Prompt("Enter template name:")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}

	// Get template description
	description, err := tui.Prompt("Enter template description:")
	if err != nil {
		return nil, err
	}

	// Create template with structured input
	template, err := createStructuredTemplateInit(name, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Ask if user wants to save this template for future use
	saveTemplate, err := tui.Confirm("Save this template for future use?")
	if err != nil {
		return nil, err
	}

	if saveTemplate {
		tm, err := config.NewTemplateManager()
		if err != nil {
			fmt.Printf("%s Warning: Could not save template: %v\n", infoStyle.Render("ℹ"), err)
		} else {
			// Convert name to key
			key := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

			// Check if template already exists
			exists, err := tm.TemplateExists(key)
			if err == nil && exists {
				overwrite, err := tui.Confirm(fmt.Sprintf("Template '%s' already exists. Overwrite?", name))
				if err != nil {
					fmt.Printf("%s Warning: Could not save template: %v\n", infoStyle.Render("ℹ"), err)
				} else if overwrite {
					if err := tm.SaveTemplate(key, *template); err != nil {
						fmt.Printf("%s Warning: Could not save template: %v\n", infoStyle.Render("ℹ"), err)
					} else {
						fmt.Printf("%s Template '%s' saved successfully!\n", successStyle.Render("✅"), name)
					}
				}
			} else if err == nil {
				if err := tm.SaveTemplate(key, *template); err != nil {
					fmt.Printf("%s Warning: Could not save template: %v\n", infoStyle.Render("ℹ"), err)
				} else {
					fmt.Printf("%s Template '%s' saved successfully!\n", successStyle.Render("✅"), name)
				}
			}
		}
	}

	return template, nil
}

// createStructuredTemplateInit creates a template with structured, optional sections during init
func createStructuredTemplateInit(name, description string) (*config.Template, error) {
	fmt.Printf("\n%s Creating structured template: %s\n", infoStyle.Render("ℹ"), name)
	fmt.Println("Each section is optional. You can skip sections to use generic fallback text.")
	fmt.Println()

	var templateParts []string

	// Add header
	templateParts = append(templateParts, "# "+name)
	templateParts = append(templateParts, "")

	// Ask for project overview
	overview, err := tui.Prompt("Enter project overview (optional, press Enter to skip):")
	if err != nil {
		return nil, err
	}
	if overview == "" {
		overview = "Project overview and context."
	}
	templateParts = append(templateParts, overview)
	templateParts = append(templateParts, "")

	// Ask for common instructions
	templateParts = append(templateParts, "## Common Instructions")
	templateParts = append(templateParts, "")

	addCommon, err := tui.Confirm("Do you want to add custom common instructions?")
	if err != nil {
		return nil, err
	}

	if addCommon {
		fmt.Printf("\n%s Enter common instructions (press Ctrl+D when finished):\n", infoStyle.Render("ℹ"))
		fmt.Println("These apply to all AI providers.")
		fmt.Println()

		commonInstructions, err := tui.ReadMultilineInput()
		if err != nil {
			return nil, fmt.Errorf("failed to read common instructions: %w", err)
		}
		if strings.TrimSpace(commonInstructions) == "" {
			commonInstructions = config.DefaultCommonInstructions
		}
		templateParts = append(templateParts, commonInstructions)
	} else {
		templateParts = append(templateParts, config.DefaultCommonInstructions)
	}

	templateParts = append(templateParts, "")
	templateParts = append(templateParts, "## Provider Overrides")
	templateParts = append(templateParts, "")

	// Get providers from registry
	providers := sync.GetAvailableProviders()

	for _, provider := range providers {
		templateParts = append(templateParts, "### "+provider)

		addProvider, err := tui.Confirm(fmt.Sprintf("Do you want to add custom %s instructions?", provider))
		if err != nil {
			return nil, err
		}

		if addProvider {
			fmt.Printf("\n%s Enter %s-specific instructions (press Ctrl+D when finished):\n", infoStyle.Render("ℹ"), provider)
			fmt.Println("These will override or supplement the common instructions for " + provider + ".")
			fmt.Println()

			providerInstructions, err := tui.ReadMultilineInput()
			if err != nil {
				return nil, fmt.Errorf("failed to read %s instructions: %w", provider, err)
			}
			if strings.TrimSpace(providerInstructions) == "" {
				providerInstructions = config.DefaultProviderInstructions
			}
			templateParts = append(templateParts, providerInstructions)
		} else {
			templateParts = append(templateParts, config.DefaultProviderInstructions)
		}
		templateParts = append(templateParts, "")
	}

	// Build the final template content
	content := strings.Join(templateParts, "\n")

	// Remove trailing newlines
	content = strings.TrimSuffix(content, "\n\n")

	// Inject timestamp for model grounding
	contentWithTimestamp := config.InjectTimestamp(content)

	return &config.Template{
		Name:        name,
		Description: description,
		Content:     contentWithTimestamp,
	}, nil
}

// runProviderSetupFlow provides the provider registry setup functionality integrated into init
func runProviderSetupFlow() error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	for {
		action, err := selectSetupActionForInit()
		if err != nil {
			return err
		}

		switch action {
		case "manage":
			if err := manageProvidersForInit(tm); err != nil {
				return err
			}
		case "add":
			if err := addCustomProviderForInit(tm); err != nil {
				return err
			}
		case "remove":
			if err := removeProviderForInit(tm); err != nil {
				return err
			}
		case "status":
			if err := showProviderStatusForInit(tm); err != nil {
				return err
			}
		case "done":
			fmt.Printf("%s Provider setup complete!\n", successStyle.Render("✅"))
			return nil
		}

		// Ask if user wants to continue
		if action != "status" {
			shouldContinue, err := tui.Confirm("Continue configuring providers?")
			if err != nil {
				return err
			}
			if !shouldContinue {
				fmt.Printf("%s Provider setup complete!\n", successStyle.Render("✅"))
				return nil
			}
		}
	}
}

// setupActionItemForInit represents an action option in the setup menu for init
type setupActionItemForInit struct {
	key         string
	title       string
	description string
}

func (s setupActionItemForInit) FilterValue() string { return s.title }
func (s setupActionItemForInit) Title() string       { return s.title }
func (s setupActionItemForInit) Description() string { return s.description }

// selectSetupActionForInit prompts user to select what they want to do in the provider setup
func selectSetupActionForInit() (string, error) {
	items := []list.Item{
		setupActionItemForInit{key: "manage", title: "Manage Provider Settings", description: "Enable/disable providers and view current settings"},
		setupActionItemForInit{key: "add", title: "Add Custom Provider", description: "Add a new provider to the registry"},
		setupActionItemForInit{key: "remove", title: "Remove Provider", description: "Remove a provider from the registry"},
		setupActionItemForInit{key: "status", title: "Show Provider Status", description: "View current provider status and settings"},
		setupActionItemForInit{key: "done", title: "Finish Provider Setup", description: "Continue with RIZOME.md template selection"},
	}

	selected, err := tui.ListSelection("Configure your AI provider registry:", items)
	if err != nil {
		return "", err
	}

	if actionItem, ok := selected.(setupActionItemForInit); ok {
		return actionItem.key, nil
	}

	return "", fmt.Errorf("unexpected selection type")
}

// manageProvidersForInit provides interactive provider enable/disable functionality for init
func manageProvidersForInit(tm *config.TemplateManager) error {
	registry, err := tm.GetProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to load provider registry: %w", err)
	}

	// Create checkbox options from providers
	var options []tui.CheckboxOption
	for _, provider := range registry.Providers {
		description := provider.Description
		if provider.Category != "" {
			description = fmt.Sprintf("[%s] %s", provider.Category, provider.Description)
		}

		options = append(options, tui.CheckboxOption{
			Label:       provider.Name,
			Description: description,
			Value:       provider.Name,
			Checked:     provider.Enabled,
		})
	}

	if len(options) == 0 {
		fmt.Println("No providers found in registry.")
		return nil
	}

	selected, err := tui.CheckboxSelection("Select providers to enable (deselected will be disabled by default):", options)
	if err != nil {
		return err
	}

	// Update provider enabled status
	for _, provider := range registry.Providers {
		enabled := false
		for _, selectedKey := range selected {
			if provider.Name == selectedKey {
				enabled = true
				break
			}
		}

		if err := tm.SetProviderEnabled(provider.Name, enabled); err != nil {
			return fmt.Errorf("failed to update provider %s: %w", provider.Name, err)
		}
	}

	fmt.Printf("Updated provider settings. %d providers enabled.\n", len(selected))
	return nil
}

// categoryOptionItemForInit represents a category option in the category selection for init
type categoryOptionItemForInit struct {
	key         string
	title       string
	description string
}

func (c categoryOptionItemForInit) FilterValue() string { return c.title }
func (c categoryOptionItemForInit) Title() string       { return c.title }
func (c categoryOptionItemForInit) Description() string { return c.description }

// addCustomProviderForInit allows users to add a new provider for init
func addCustomProviderForInit(tm *config.TemplateManager) error {
	fmt.Printf("\n%s Add Custom Provider\n", infoStyle.Render("ℹ"))

	name, err := tui.Prompt("Provider name (e.g., GPT4, COPILOT):")
	if err != nil {
		return err
	}

	name = strings.TrimSpace(strings.ToUpper(name))
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Check if provider already exists
	registry, err := tm.GetProviderRegistry()
	if err != nil {
		return err
	}

	if _, exists := registry.GetProvider(name); exists {
		return fmt.Errorf("provider '%s' already exists", name)
	}

	description, err := tui.Prompt(fmt.Sprintf("Description for %s:", name))
	if err != nil {
		return err
	}

	categoryOptions := []list.Item{
		categoryOptionItemForInit{key: "Chat", title: "Chat", description: "AI chat/conversation models"},
		categoryOptionItemForInit{key: "Code Editor", title: "Code Editor", description: "AI-powered code editors and IDEs"},
		categoryOptionItemForInit{key: "API", title: "API", description: "AI API services"},
		categoryOptionItemForInit{key: "Other", title: "Other", description: "Other AI providers"},
		categoryOptionItemForInit{key: "custom", title: "Custom Category", description: "Enter a custom category"},
	}

	selectedCategory, err := tui.ListSelection("Select provider category:", categoryOptions)
	if err != nil {
		return err
	}

	categoryItem, ok := selectedCategory.(categoryOptionItemForInit)
	if !ok {
		return fmt.Errorf("unexpected category selection type")
	}

	category := categoryItem.key
	if category == "custom" {
		customCategory, err := tui.Prompt("Enter custom category:")
		if err != nil {
			return err
		}
		category = strings.TrimSpace(customCategory)
	}

	enabled, err := tui.Confirm("Enable this provider by default?")
	if err != nil {
		return err
	}

	// Add the provider
	provider := config.Provider{
		Name:        name,
		Description: description,
		Category:    category,
		Enabled:     enabled,
	}

	if err := tm.AddProvider(provider); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	fmt.Printf("%s Successfully added provider '%s'!\n", successStyle.Render("✅"), name)
	return nil
}

// providerOptionItemForInit represents a provider option in the provider selection for init
type providerOptionItemForInit struct {
	name        string
	title       string
	description string
}

func (p providerOptionItemForInit) FilterValue() string { return p.title }
func (p providerOptionItemForInit) Title() string       { return p.title }
func (p providerOptionItemForInit) Description() string { return p.description }

// removeProviderForInit allows users to remove a provider for init
func removeProviderForInit(tm *config.TemplateManager) error {
	registry, err := tm.GetProviderRegistry()
	if err != nil {
		return err
	}

	if len(registry.Providers) == 0 {
		fmt.Println("No providers to remove.")
		return nil
	}

	// Create list of providers to remove (excluding defaults to prevent accidental deletion)
	var options []list.Item
	defaultNames := make(map[string]bool)
	for _, defaultProvider := range config.GetDefaultProviders() {
		defaultNames[defaultProvider.Name] = true
	}

	for _, provider := range registry.Providers {
		title := provider.Name
		description := provider.Description
		
		if defaultNames[provider.Name] {
			description += " (default provider - not recommended to remove)"
		}

		if provider.Category != "" {
			description = fmt.Sprintf("[%s] %s", provider.Category, description)
		}

		options = append(options, providerOptionItemForInit{
			name:        provider.Name,
			title:       title,
			description: description,
		})
	}

	if len(options) == 0 {
		fmt.Println("No custom providers to remove.")
		return nil
	}

	selectedItem, err := tui.ListSelection("Select provider to remove:", options)
	if err != nil {
		return err
	}

	providerItem, ok := selectedItem.(providerOptionItemForInit)
	if !ok {
		return fmt.Errorf("unexpected provider selection type")
	}

	// Confirm removal
	confirmed, err := tui.Confirm(fmt.Sprintf("Are you sure you want to remove provider '%s'?", providerItem.name))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Removal cancelled.")
		return nil
	}

	if err := tm.RemoveProvider(providerItem.name); err != nil {
		return fmt.Errorf("failed to remove provider: %w", err)
	}

	fmt.Printf("%s Successfully removed provider '%s'.\n", successStyle.Render("✅"), providerItem.name)
	return nil
}

// initializeClaudedocs creates the .claudedocs folder with IMPLEMENTATION_PLAN.md and IMPLEMENTATION_PROGRESS.md
func initializeClaudedocs(projectPath string) error {
	claudedocsPath := filepath.Join(projectPath, ".claudedocs")
	
	// Create .claudedocs directory
	if err := os.MkdirAll(claudedocsPath, 0755); err != nil {
		return fmt.Errorf("failed to create .claudedocs directory: %w", err)
	}
	
	// Create IMPLEMENTATION_PLAN.md
	implementationPlanPath := filepath.Join(claudedocsPath, "IMPLEMENTATION_PLAN.md")
	implementationPlanContent := `# Implementation Plan

## Project Overview
<!-- Brief description of the project and its goals -->

## Architecture
<!-- High-level architecture and design decisions -->

## Implementation Phases

### Phase 1: Foundation
<!-- Initial setup and core components -->
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Phase 2: Core Features
<!-- Main functionality implementation -->
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Phase 3: Polish & Testing
<!-- Testing, optimization, and final touches -->
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Technical Decisions
<!-- Document key technical decisions and rationale -->

## Dependencies
<!-- List external dependencies and libraries -->

## Testing Strategy
<!-- Approach to testing the implementation -->

## Documentation Requirements
<!-- What documentation needs to be created -->

## Success Criteria
<!-- How to measure successful implementation -->
`
	
	if err := os.WriteFile(implementationPlanPath, []byte(implementationPlanContent), 0644); err != nil {
		return fmt.Errorf("failed to create IMPLEMENTATION_PLAN.md: %w", err)
	}
	
	// Create IMPLEMENTATION_PROGRESS.md
	implementationProgressPath := filepath.Join(claudedocsPath, "IMPLEMENTATION_PROGRESS.md")
	implementationProgressContent := `# Implementation Progress

## Current Status
<!-- Overall project status and current phase -->
**Phase:** Foundation
**Status:** Not Started
**Last Updated:** ` + time.Now().Format("2006-01-02") + `

## Completed Tasks
<!-- Track completed implementation tasks -->

## In Progress
<!-- Tasks currently being worked on -->

## Blockers
<!-- Any issues blocking progress -->

## Next Steps
<!-- Immediate next actions -->

## Notes
<!-- Additional notes and observations -->

## Change Log
<!-- Track major changes and decisions -->

### ` + time.Now().Format("2006-01-02") + `
- Project initialized with .claudedocs
`
	
	if err := os.WriteFile(implementationProgressPath, []byte(implementationProgressContent), 0644); err != nil {
		return fmt.Errorf("failed to create IMPLEMENTATION_PROGRESS.md: %w", err)
	}
	
	return nil
}

// showProviderStatusForInit displays current provider configuration for init
func showProviderStatusForInit(tm *config.TemplateManager) error {
	registry, err := tm.GetProviderRegistry()
	if err != nil {
		return err
	}

	if len(registry.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	fmt.Println("\n=== Provider Status ===")
	
	// Group by category
	categories := registry.GetCategories()
	if len(categories) == 0 {
		// No categories, show all providers
		categories = []string{""}
	}

	for _, category := range categories {
		if category != "" {
			fmt.Printf("\n[%s]\n", category)
		}

		var providersInCategory []config.Provider
		if category == "" {
			// Show providers with no category
			for _, provider := range registry.Providers {
				if provider.Category == "" {
					providersInCategory = append(providersInCategory, provider)
				}
			}
		} else {
			providersInCategory = registry.GetProvidersByCategory(category)
		}

		for _, provider := range providersInCategory {
			status := "✗ Disabled"
			if provider.Enabled {
				status = "✓ Enabled"
			}
			
			fmt.Printf("  %s %s\n", status, provider.Name)
			if provider.Description != "" {
				fmt.Printf("    %s\n", provider.Description)
			}
		}
	}

	enabledCount := len(registry.GetEnabledProviders())
	totalCount := len(registry.Providers)
	fmt.Printf("\nSummary: %d/%d providers enabled\n", enabledCount, totalCount)

	return nil
}
