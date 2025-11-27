package cmd

import (
	"github.com/akulkarni/0perator/internal/cli"
	"github.com/spf13/cobra"
)

// buildUninstallCmd creates the uninstall command
func buildUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove 0perator and MCP configurations",
		Long: `Remove 0perator from your system.

This command will:
  1. Prompt for confirmation
  2. Remove MCP server configurations from all configured IDEs
  3. Remove the 0perator binary (if in standard location)
  4. Remove the 0perator config directory`,
		Example: `  0perator uninstall  # Remove 0perator from your system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No argument validation needed - this command takes no args
			cmd.SilenceUsage = true

			return cli.Uninstall()
		},
	}
}
