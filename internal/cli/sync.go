package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rizome-dev/rizome/internal/config"
	"github.com/rizome-dev/rizome/internal/sync"
	"github.com/rizome-dev/rizome/internal/tui"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0EA5E9")).
			Bold(true)
)

// SyncCmd creates the sync command
func SyncCmd() *cobra.Command {
	var (
		dryRun         bool
		force          bool
		nonInteractive bool
		providers      string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Interactive provider configuration synchronization",
		Long: `The sync command reads RIZOME.md from the current directory and synchronizes
its content with provider-specific configuration files (CLAUDE.md, QWEN.md, 
CURSOR.md, GEMINI.md, WINDSURF.md).

Interactive Mode (default):
  Select which providers to sync using an interactive checkbox interface.
  Providers enabled in the registry are pre-selected by default.

Non-Interactive Mode:
  Use --non-interactive flag to sync enabled providers or specify providers with --providers.

RIZOME.md format:
  # RIZOME.md
  
  ## Common Instructions
  Instructions that apply to all AI providers
  
  ## Provider Overrides
  ### CLAUDE
  Claude-specific instructions
  
  ### QWEN
  Qwen-specific instructions

The sync command will create or update individual provider files with:
1. Common instructions section
2. Provider-specific overrides (if any)

Provider Registry:
  Use 'rizome init' to configure which providers are enabled by default.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncInteractive(dryRun, force, nonInteractive, providers)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files without prompting")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode (sync enabled providers)")
	cmd.Flags().StringVar(&providers, "providers", "", "Comma-separated list of providers to sync (requires --non-interactive)")

	return cmd
}

// runSyncInteractive runs the interactive sync command
func runSyncInteractive(dryRun, force, nonInteractive bool, providersFlag string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize sync manager to get available providers
	syncManager, err := sync.New(cwd)
	if err != nil {
		return fmt.Errorf("%s %w", errorStyle.Render("✗"), err)
	}

	// Get available providers from the sync manager
	config := syncManager.GetConfig()
	availableProviders := config.Providers

	var selectedProviders []string

	if nonInteractive {
		// Non-interactive mode
		if providersFlag != "" {
			// Use specified providers
			specifiedProviders := strings.Split(providersFlag, ",")
			for i, p := range specifiedProviders {
				specifiedProviders[i] = strings.ToUpper(strings.TrimSpace(p))
			}

			// Validate specified providers
			for _, provider := range specifiedProviders {
				found := false
				for _, available := range availableProviders {
					if provider == available {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("unknown provider '%s'. Available providers: %s",
						provider, strings.Join(availableProviders, ", "))
				}
			}

			selectedProviders = specifiedProviders
		} else {
			// Use enabled providers from registry as default
			selectedProviders = sync.GetEnabledProviders()
		}
	} else {
		// Interactive mode - show provider selection
		selected, err := selectProvidersForSync(availableProviders, config)
		if err != nil {
			return err
		}
		selectedProviders = selected
	}

	if len(selectedProviders) == 0 {
		fmt.Printf("%s No providers selected for sync\n", infoStyle.Render("ℹ"))
		return nil
	}

	fmt.Printf("%s Starting sync in %s\n", infoStyle.Render("ℹ"), cwd)
	if len(selectedProviders) < len(availableProviders) {
		fmt.Printf("  Providers: %s\n", strings.Join(selectedProviders, ", "))
	}

	if dryRun {
		fmt.Printf("%s Running in dry-run mode\n", infoStyle.Render("ℹ"))
	}

	// Perform sync with selected providers
	results, err := syncManager.SyncProviders(selectedProviders, dryRun, force)
	if err != nil {
		return fmt.Errorf("%s %w", errorStyle.Render("✗"), err)
	}

	// Display results
	fmt.Println()
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("%s %s.md: %v\n", errorStyle.Render("✗"), result.Provider, result.Error)
		} else {
			action := "synced"
			if result.Created {
				action = "created"
			} else if result.Updated {
				action = "updated"
			}

			if dryRun {
				action = "would be " + action
			}

			fmt.Printf("%s %s.md %s\n", successStyle.Render("✓"), result.Provider, action)
		}
	}

	if !dryRun {
		fmt.Printf("\n%s Sync completed successfully!\n", successStyle.Render("✅"))
	} else {
		fmt.Printf("\n%s Dry run completed. Run without --dry-run to apply changes.\n", infoStyle.Render("ℹ"))
	}

	return nil
}

// selectProvidersForSync shows an interactive provider selection interface
func selectProvidersForSync(availableProviders []string, syncConfig *sync.Config) ([]string, error) {
	// Show current RIZOME.md summary
	fmt.Printf("%s RIZOME.md Summary:\n", infoStyle.Render("ℹ"))

	if syncConfig.CommonInstructions != "" {
		lines := strings.Split(syncConfig.CommonInstructions, "\n")
		fmt.Printf("  Common Instructions: %d lines\n", len(lines))
	} else {
		fmt.Printf("  Common Instructions: none\n")
	}

	overrideCount := len(syncConfig.ProviderOverrides)
	if overrideCount > 0 {
		fmt.Printf("  Provider Overrides: %d (%s)\n", overrideCount, strings.Join(getOverrideKeys(syncConfig.ProviderOverrides), ", "))
	} else {
		fmt.Printf("  Provider Overrides: none\n")
	}

	fmt.Println()

	// Get enabled providers from registry for default selections
	enabledProviders := sync.GetEnabledProviders()
	enabledMap := make(map[string]bool)
	for _, provider := range enabledProviders {
		enabledMap[provider] = true
	}

	// Get provider registry for descriptions
	tm, err := config.NewTemplateManager()
	var registry *config.ProviderRegistry
	if err == nil {
		registry, err = tm.GetProviderRegistry()
	}

	// Create checkbox options for providers
	var options []tui.CheckboxOption
	for _, provider := range availableProviders {
		description := fmt.Sprintf("Generate %s.md file", provider)
		if _, hasOverride := syncConfig.ProviderOverrides[provider]; hasOverride {
			description += " (has specific overrides)"
		}

		// Add provider description from registry if available
		if registry != nil {
			if providerInfo, exists := registry.GetProvider(provider); exists {
				if providerInfo.Description != "" {
					description = fmt.Sprintf("%s - %s", description, providerInfo.Description)
				}
				if providerInfo.Category != "" {
					description = fmt.Sprintf("[%s] %s", providerInfo.Category, description)
				}
			}
		}

		// Check if provider is enabled by default in registry
		isEnabledByDefault := enabledMap[provider]

		option := tui.CheckboxOption{
			Label:       provider,
			Description: description,
			Value:       provider,
			Checked:     isEnabledByDefault, // Use registry default
		}
		options = append(options, option)
	}

	// Show provider selection
	selected, err := tui.CheckboxSelection("Select providers to sync (registry defaults pre-selected):", options)

	if err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		confirmed, err := tui.Confirm("No providers selected. Continue without syncing?")
		if err != nil {
			return nil, err
		}
		if !confirmed {
			return selectProvidersForSync(availableProviders, syncConfig) // Try again
		}
	}

	return selected, nil
}

// getOverrideKeys returns the keys from the provider overrides map
func getOverrideKeys(overrides map[string]string) []string {
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	return keys
}
