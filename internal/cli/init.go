package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	fmt.Printf("\n%s Next steps:\n", infoStyle.Render("ℹ"))
	fmt.Printf("  1. Edit RIZOME.md with your project details\n")
	fmt.Printf("  2. Run 'rizome sync' to generate provider-specific files\n")
	fmt.Printf("  3. Use 'rizome tmpl' to manage templates\n")

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
