package cmd

import (
	"github.com/akulkarni/0perator/internal/cli"
	"github.com/spf13/cobra"
)

// buildInitCmd creates the init command
func buildInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure IDEs with MCP servers",
		Long: `Initialize 0perator by configuring your IDE(s) with MCP servers.

This interactive command will:
  1. Check and install tiger-cli if needed
  2. Authenticate with Tiger Cloud
  3. Let you select which IDE(s) to configure
  4. Install Tiger MCP and 0perator MCP servers`,
		Example: `  0perator init       # Set up your IDE interactively`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No argument validation needed - this command takes no args
			cmd.SilenceUsage = true

			return cli.Init()
		},
	}
}
