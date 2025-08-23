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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd returns the root command
func RootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "rizome",
		Short: "Agentic development environment workflow management",
		Long: `Rizome CLI manages workflows for agentic software development environments.
It synchronizes configuration across AI providers and manages development workflows.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(configFile)
		},
		// Override default help behavior to show our custom grouped commands
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.rizome/config.yaml)")

	// Set custom help template
	rootCmd.SetHelpTemplate(customHelpTemplate())

	// Disable the default help command and add our own
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	// Add custom help command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type rizome help [path to command] for full details.`,
		DisableFlagsInUseLine: true,
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var completions []string
			cmd, _, e := c.Root().Find(args)
			if e != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if cmd == nil {
				cmd = c.Root()
			}
			for _, subCmd := range cmd.Commands() {
				if subCmd.IsAvailableCommand() {
					completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(c *cobra.Command, args []string) {
			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				_ = c.Root().Usage()
			} else {
				_ = cmd.Help()
			}
		},
	})

	// Add main commands
	rootCmd.AddCommand(
		SyncCmd(),
		CompletionCmd(),
	)

	return rootCmd
}

// customHelpTemplate returns a custom help template with grouped commands
func customHelpTemplate() string {
	return `{{.Long}}

Usage:
  {{.UseLine}}

Main Commands:
  sync        Synchronize RIZOME.md with provider-specific configuration files

System Commands:
  completion  Generate shell completions

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`
}

// GetCustomHelp returns the formatted help text for display
func GetCustomHelp() string {
	return `Rizome CLI manages workflows for agentic software development environments.
It synchronizes configuration across AI providers and manages development workflows.

Usage:
  rizome [command]

Main Commands:
  sync        Synchronize RIZOME.md with provider-specific configuration files

System Commands:
  completion  Generate shell completions
  help        Help about any command

Flags:
  --config string   config file (default is $HOME/.rizome/config.yaml)
  -h, --help       help for rizome

Use "rizome [command] --help" for more information about a command.
`
}

// CompletionCmd generates shell completions
func CompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `To load completions:

Bash:
  $ source <(rizome completion bash)
  # To load completions for each session, execute once:
  $ rizome completion bash > /etc/bash_completion.d/rizome

Zsh:
  $ source <(rizome completion zsh)
  # To load completions for each session, execute once:
  $ rizome completion zsh > "${fpath[1]}/_rizome"

Fish:
  $ rizome completion fish | source
  # To load completions for each session, execute once:
  $ rizome completion fish > ~/.config/fish/completions/rizome.fish

PowerShell:
  PS> rizome completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> rizome completion powershell > rizome.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}

func initConfig(configFile string) error {
	if configFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(configFile)
	} else {
		// Search for config in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		rizomeDir := filepath.Join(home, ".rizome")

		// Create rizome directory if it doesn't exist
		if err := os.MkdirAll(rizomeDir, 0755); err != nil {
			return err
		}

		viper.AddConfigPath(rizomeDir)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// Read config if it exists
	_ = viper.ReadInConfig()

	return nil
}