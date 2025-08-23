package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/rizome-dev/rizome/internal/sync"
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
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize RIZOME.md with provider-specific configuration files",
		Long: `The sync command reads RIZOME.md from the current directory and synchronizes
its content with provider-specific configuration files (CLAUDE.md, QWEN.md, 
CURSOR.md, GEMINI.md, etc.).

RIZOME.md format:
  # RIZOME.md
  
  ## Common Instructions
  Instructions that apply to all AI providers
  
  ## Provider Overrides
  ### Claude
  Claude-specific instructions
  
  ### Qwen
  Qwen-specific instructions

The sync command will create or update individual provider files with:
1. Common instructions section
2. Provider-specific overrides (if any)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(dryRun, force)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files without prompting")

	return cmd
}

func runSync(dryRun, force bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Printf("%s Starting sync in %s\n", infoStyle.Render("ℹ"), cwd)

	syncManager, err := sync.New(cwd)
	if err != nil {
		return fmt.Errorf("%s %w", errorStyle.Render("✗"), err)
	}

	if dryRun {
		fmt.Printf("%s Running in dry-run mode\n", infoStyle.Render("ℹ"))
	}

	results, err := syncManager.Sync(dryRun, force)
	if err != nil {
		return fmt.Errorf("%s %w", errorStyle.Render("✗"), err)
	}

	// Display results
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("%s %s: %v\n", errorStyle.Render("✗"), result.Provider, result.Error)
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
		fmt.Printf("\n%s Dry run completed. Use --force to apply changes.\n", infoStyle.Render("ℹ"))
	}

	return nil
}