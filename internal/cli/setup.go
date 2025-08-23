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
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/spf13/cobra"

	"github.com/rizome-dev/rizome/internal/config"
	"github.com/rizome-dev/rizome/internal/tui"
)

// setupActionItem represents an action option in the setup menu
type setupActionItem struct {
	key         string
	title       string
	description string
}

func (s setupActionItem) FilterValue() string { return s.title }
func (s setupActionItem) Title() string       { return s.title }
func (s setupActionItem) Description() string { return s.description }

// categoryOptionItem represents a category option in the category selection
type categoryOptionItem struct {
	key         string
	title       string
	description string
}

func (c categoryOptionItem) FilterValue() string { return c.title }
func (c categoryOptionItem) Title() string       { return c.title }
func (c categoryOptionItem) Description() string { return c.description }

// providerOptionItem represents a provider option in the provider selection
type providerOptionItem struct {
	name        string
	title       string
	description string
}

func (p providerOptionItem) FilterValue() string { return p.title }
func (p providerOptionItem) Title() string       { return p.title }
func (p providerOptionItem) Description() string { return p.description }

// SetupCmd creates the setup command
func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Manage provider registry and default settings",
		Long: `Setup allows you to manage the provider registry and configure default settings.
		
You can enable/disable providers, add custom providers, and manage which providers
are selected by default in other commands like sync.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}

	return cmd
}

// runSetup executes the setup command with interactive provider management
func runSetup() error {
	tm, err := config.NewTemplateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	for {
		action, err := selectSetupAction()
		if err != nil {
			return err
		}

		switch action {
		case "manage":
			if err := manageProviders(tm); err != nil {
				return err
			}
		case "add":
			if err := addCustomProvider(tm); err != nil {
				return err
			}
		case "remove":
			if err := removeProvider(tm); err != nil {
				return err
			}
		case "status":
			if err := showProviderStatus(tm); err != nil {
				return err
			}
		case "exit":
			fmt.Println("Setup complete!")
			return nil
		}

		// Ask if user wants to continue
		if action != "status" {
			shouldContinue, err := tui.Confirm("Continue with setup?")
			if err != nil {
				return err
			}
			if !shouldContinue {
				fmt.Println("Setup complete!")
				return nil
			}
		}
	}
}

// selectSetupAction prompts user to select what they want to do
func selectSetupAction() (string, error) {
	items := []list.Item{
		setupActionItem{key: "manage", title: "Manage Provider Settings", description: "Enable/disable providers and view current settings"},
		setupActionItem{key: "add", title: "Add Custom Provider", description: "Add a new provider to the registry"},
		setupActionItem{key: "remove", title: "Remove Provider", description: "Remove a provider from the registry"},
		setupActionItem{key: "status", title: "Show Provider Status", description: "View current provider status and settings"},
		setupActionItem{key: "exit", title: "Exit Setup", description: "Finish configuration and exit"},
	}

	selected, err := tui.ListSelection("What would you like to do?", items)
	if err != nil {
		return "", err
	}

	if actionItem, ok := selected.(setupActionItem); ok {
		return actionItem.key, nil
	}

	return "", fmt.Errorf("unexpected selection type")
}

// manageProviders provides interactive provider enable/disable functionality
func manageProviders(tm *config.TemplateManager) error {
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

// addCustomProvider allows users to add a new provider
func addCustomProvider(tm *config.TemplateManager) error {
	fmt.Println("Add Custom Provider")

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
		categoryOptionItem{key: "Chat", title: "Chat", description: "AI chat/conversation models"},
		categoryOptionItem{key: "Code Editor", title: "Code Editor", description: "AI-powered code editors and IDEs"},
		categoryOptionItem{key: "API", title: "API", description: "AI API services"},
		categoryOptionItem{key: "Other", title: "Other", description: "Other AI providers"},
		categoryOptionItem{key: "custom", title: "Custom Category", description: "Enter a custom category"},
	}

	selectedCategory, err := tui.ListSelection("Select provider category:", categoryOptions)
	if err != nil {
		return err
	}

	categoryItem, ok := selectedCategory.(categoryOptionItem)
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

	fmt.Printf("Successfully added provider '%s'!\n", name)
	return nil
}

// removeProvider allows users to remove a provider
func removeProvider(tm *config.TemplateManager) error {
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

		options = append(options, providerOptionItem{
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

	providerItem, ok := selectedItem.(providerOptionItem)
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

	fmt.Printf("Successfully removed provider '%s'.\n", providerItem.name)
	return nil
}

// showProviderStatus displays current provider configuration
func showProviderStatus(tm *config.TemplateManager) error {
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