package cmd

import (
	"github.com/akulkarni/0perator/internal/cli"
	"github.com/spf13/cobra"
)

// buildMCPCmd creates the mcp command with subcommands
func buildMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  `Commands for managing the 0perator MCP (Model Context Protocol) server.`,
	}

	cmd.AddCommand(buildMCPStartCmd())

	return cmd
}

// buildMCPStartCmd creates the mcp start command
func buildMCPStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the MCP server",
		Long: `Start the 0perator MCP (Model Context Protocol) server.

The MCP server enables AI assistants to interact with 0perator's capabilities.`,
		Example: `  0perator mcp start`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return cli.MCPStart()
		},
	}
}
