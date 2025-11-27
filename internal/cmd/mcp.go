package cmd

import (
	"github.com/akulkarni/0perator/internal/cli"
	"github.com/spf13/cobra"
)

// buildMCPCmd creates the mcp command
func buildMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server",
		Long: `Start the 0perator MCP (Model Context Protocol) server.

This is the default command when 0perator is run without any arguments.
The MCP server enables AI assistants to interact with 0perator's capabilities.`,
		Example: `  0perator mcp        # Start MCP server explicitly
  0perator            # Start MCP server (default)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No argument validation needed
			cmd.SilenceUsage = true

			return runMCP(cmd)
		},
	}
}

// runMCP starts the MCP server
func runMCP(cmd *cobra.Command) error {
	return cli.MCPStart()
}
