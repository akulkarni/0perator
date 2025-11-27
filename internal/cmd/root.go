package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "2.0.2"

// buildRootCmd creates and returns the root command with all subcommands attached.
// This is the main entry point for building the complete CLI tree.
func buildRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "0perator",
		Short: "Infrastructure for AI native development",
		Long: `0perator - Infrastructure for AI native development

0perator provides MCP (Model Context Protocol) servers that enable
AI assistants to help you build and deploy applications.

Running 0perator without any command starts the MCP server (default behavior).`,
		// When run without subcommands, start the MCP server
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is the default behavior when no subcommand is provided
			return runMCP(cmd)
		},
	}

	// Add all subcommands
	cmd.AddCommand(buildVersionCmd())
	cmd.AddCommand(buildInitCmd())
	cmd.AddCommand(buildUninstallCmd())
	cmd.AddCommand(buildMCPCmd())

	return cmd
}

// Execute builds and executes the root command.
// This is the main entry point called from main().
func Execute() {
	rootCmd := buildRootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
