package cli

// Copyright (C) 2025 Rizome Labs, Inc.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/spf13/cobra"

	"github.com/rizome-dev/rizome/internal/config"
	"github.com/rizome-dev/rizome/internal/tui"
)

// templateListItem represents a template in the list view
type templateListItem struct {
	key         string
	template    config.Template
	description string
}

func (t templateListItem) FilterValue() string { return t.template.Name }
func (t templateListItem) Title() string       { return t.template.Name }
func (t templateListItem) Description() string { return t.template.Description }

// TmplCmd creates the template management command
func TmplCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmpl",
		Short: "Manage RIZOME.md templates",
		Long: `The tmpl command provides template management functionality for RIZOME.md files.
Templates are stored in ~/.rizome/config.yaml and can be reused across projects.

Available subcommands:
  list    List all available templates
  add     Add a new template interactively  
  edit    Edit an existing template
  delete  Delete a template
  show    Show template content`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default to list if no subcommand specified
			runTmplList()
		},
	}

	// Add subcommands
	cmd.AddCommand(
		tmplListCmd(),
		tmplAddCmd(),
		tmplEditCmd(),
		tmplDeleteCmd(),
		tmplShowCmd(),
	)

	return cmd
}

// tmplListCmd lists all available templates
func tmplListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		Run: func(cmd *cobra.Command, args []string) {
			runTmplList()
		},
	}
}

// tmplAddCmd adds a new template interactively
func tmplAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new template interactively",
		Long: `Add a new template interactively. If no name is provided,
you'll be prompted to enter one.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runTmplAdd(name)
		},
	}
}

// tmplEditCmd edits an existing template
func tmplEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit [name]",
		Short: "Edit an existing template",
		Long: `Edit an existing template. If no name is provided,
you'll be prompted to select one from a list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runTmplEdit(name)
		},
	}
}

// tmplDeleteCmd deletes a template
func tmplDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a template",
		Long: `Delete a template. If no name is provided,
you'll be prompted to select one from a list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runTmplDelete(name, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")
	return cmd
}

// tmplShowCmd shows template content
func tmplShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show template content",
		Long: `Show the full content of a template. If no name is provided,
you'll be prompted to select one from a list.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runTmplShow(name)
		},
	}
}

// runTmplList lists all available templates
func runTmplList() {
	tm, err := config.NewTemplateManager()
	if err != nil {
		fmt.Printf("%s Failed to initialize template manager: %v\n", errorStyle.Render("✗"), err)
		return
	}

	templates, err := tm.ListTemplates()
	if err != nil {
		fmt.Printf("%s Failed to load templates: %v\n", errorStyle.Render("✗"), err)
		return
	}

	if len(templates) == 0 {
		fmt.Printf("%s No templates found\n", infoStyle.Render("ℹ"))
		return
	}

	fmt.Printf("%s Available templates:\n\n", infoStyle.Render("ℹ"))

	// Sort templates by name for consistent display
	var keys []string
	for key := range templates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		template := templates[key]
		fmt.Printf("%s %s\n", successStyle.Render("•"), template.Name)
		if template.Description != "" {
			fmt.Printf("  %s\n", template.Description)
		}
		fmt.Printf("  Key: %s\n\n", key)
	}
}

// runTmplAdd adds a new template interactively
func runTmplAdd(name string) error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Get template name
	if name == "" {
		name, err = tui.Prompt("Enter template name:")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("template name cannot be empty")
		}
	}

	// Convert name to key (lowercase, spaces to dashes)
	key := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Check if template already exists
	exists, err := tm.TemplateExists(key)
	if err != nil {
		return fmt.Errorf("failed to check if template exists: %w", err)
	}

	if exists {
		confirmed, err := tui.Confirm(fmt.Sprintf("Template '%s' already exists. Overwrite?", name))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Printf("%s Template creation cancelled\n", infoStyle.Render("ℹ"))
			return nil
		}
	}

	// Get template description
	description, err := tui.Prompt("Enter template description:")
	if err != nil {
		return err
	}

	// Create template with structured input
	template, err := createStructuredTemplate(name, description)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	// Save template
	if err := tm.SaveTemplate(key, *template); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	fmt.Printf("\n%s Template '%s' saved successfully!\n", successStyle.Render("✅"), name)
	fmt.Printf("  Key: %s\n", key)
	fmt.Printf("  Use with: rizome init\n")

	return nil
}

// runTmplEdit edits an existing template
func runTmplEdit(name string) error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	templates, err := tm.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Printf("%s No templates available to edit\n", infoStyle.Render("ℹ"))
		return nil
	}

	// If no name provided, show selection list
	var key string
	if name == "" {
		key, err = selectTemplate(templates, "Select template to edit:")
		if err != nil {
			return err
		}
	} else {
		// Convert name to key or find matching template
		key = findTemplateKey(templates, name)
		if key == "" {
			return fmt.Errorf("template '%s' not found", name)
		}
	}

	template, err := tm.GetTemplate(key)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	fmt.Printf("\n%s Editing template: %s\n", infoStyle.Render("ℹ"), template.Name)

	// Edit name
	newName, err := tui.Prompt(fmt.Sprintf("Enter template name (current: %s):", template.Name))
	if err != nil {
		return err
	}
	if newName == "" {
		newName = template.Name
	}

	// Edit description
	newDescription, err := tui.Prompt(fmt.Sprintf("Enter template description (current: %s):", template.Description))
	if err != nil {
		return err
	}
	if newDescription == "" {
		newDescription = template.Description
	}

	// Edit content
	editContent, err := tui.Confirm("Do you want to edit the template content?")
	if err != nil {
		return err
	}

	newContent := template.Content
	if editContent {
		fmt.Printf("\n%s Current template content:\n", infoStyle.Render("ℹ"))
		fmt.Println(template.Content)
		fmt.Printf("\n%s Enter new template content:\n", infoStyle.Render("ℹ"))
		fmt.Println("(Type your template content. Press Ctrl+D when finished)")
		fmt.Println()

		newContent, err = tui.ReadMultilineInput()
		if err != nil {
			return fmt.Errorf("failed to read template content: %w", err)
		}
	}

	// Update template
	updatedTemplate := config.Template{
		Name:        newName,
		Description: newDescription,
		Content:     newContent,
	}

	if err := tm.SaveTemplate(key, updatedTemplate); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	fmt.Printf("\n%s Template '%s' updated successfully!\n", successStyle.Render("✅"), newName)

	return nil
}

// runTmplDelete deletes a template
func runTmplDelete(name string, force bool) error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	templates, err := tm.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Printf("%s No templates available to delete\n", infoStyle.Render("ℹ"))
		return nil
	}

	// If no name provided, show selection list
	var key string
	if name == "" {
		key, err = selectTemplate(templates, "Select template to delete:")
		if err != nil {
			return err
		}
	} else {
		// Convert name to key or find matching template
		key = findTemplateKey(templates, name)
		if key == "" {
			return fmt.Errorf("template '%s' not found", name)
		}
	}

	template := templates[key]

	// Confirm deletion unless force flag is used
	if !force {
		confirmed, err := tui.Confirm(fmt.Sprintf("Delete template '%s'?", template.Name))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Printf("%s Template deletion cancelled\n", infoStyle.Render("ℹ"))
			return nil
		}
	}

	// Delete template
	if err := tm.DeleteTemplate(key); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("%s Template '%s' deleted successfully!\n", successStyle.Render("✅"), template.Name)

	return nil
}

// runTmplShow shows template content
func runTmplShow(name string) error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	templates, err := tm.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Printf("%s No templates available to show\n", infoStyle.Render("ℹ"))
		return nil
	}

	// If no name provided, show selection list
	var key string
	if name == "" {
		key, err = selectTemplate(templates, "Select template to show:")
		if err != nil {
			return err
		}
	} else {
		// Convert name to key or find matching template
		key = findTemplateKey(templates, name)
		if key == "" {
			return fmt.Errorf("template '%s' not found", name)
		}
	}

	template := templates[key]

	fmt.Printf("%s Template: %s\n", infoStyle.Render("ℹ"), template.Name)
	if template.Description != "" {
		fmt.Printf("Description: %s\n", template.Description)
	}
	fmt.Printf("Key: %s\n\n", key)
	fmt.Println("Content:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(template.Content)
	fmt.Println(strings.Repeat("-", 50))

	return nil
}

// selectTemplate shows an interactive list for template selection
func selectTemplate(templates map[string]config.Template, title string) (string, error) {
	// Convert templates to list items
	var items []list.Item
	var keyMap = make(map[string]string) // map from list index to template key

	// Sort templates by name for consistent display
	var keys []string
	for key := range templates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		template := templates[key]
		item := templateListItem{
			key:         key,
			template:    template,
			description: template.Description,
		}
		items = append(items, item)
		keyMap[template.Name] = key
	}

	selectedItem, err := tui.ListSelection(title, items)
	if err != nil {
		return "", err
	}

	if templateItem, ok := selectedItem.(templateListItem); ok {
		return templateItem.key, nil
	}

	return "", fmt.Errorf("no template selected")
}

// findTemplateKey finds a template key by name (exact match or key match)
func findTemplateKey(templates map[string]config.Template, name string) string {
	// First try exact key match
	if template, exists := templates[name]; exists {
		_ = template
		return name
	}

	// Then try name match
	for key, template := range templates {
		if template.Name == name {
			return key
		}
	}

	// Finally try case-insensitive name match
	lowerName := strings.ToLower(name)
	for key, template := range templates {
		if strings.ToLower(template.Name) == lowerName {
			return key
		}
	}

	return ""
}

// createStructuredTemplate creates a template with structured, optional sections
func createStructuredTemplate(name, description string) (*config.Template, error) {
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

	// List of supported providers
	providers := []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}

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

	return &config.Template{
		Name:        name,
		Description: description,
		Content:     content,
	}, nil
}
