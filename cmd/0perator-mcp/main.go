package main

import (
	"fmt"
	"os"

	"github.com/akulkarni/0perator/internal/cli"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := cli.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "mcp":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: mcp command requires a subcommand (start)\n")
			os.Exit(1)
		}
		subcommand := os.Args[2]
		if subcommand == "start" {
			if err := cli.MCPStart(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: unknown mcp subcommand: %s\n", subcommand)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("0perator %s\n", Version)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`0perator - Infrastructure for AI agents

Usage:
  0perator init              Initialize and configure MCP servers
  0perator mcp start         Start the MCP server (stdio mode)
  0perator version           Show version information
  0perator help              Show this help message

Examples:
  $ 0perator init            # Set up 0perator with your IDE
  $ 0perator mcp start       # Run MCP server (called by IDE)

Documentation: https://0p.dev/docs`)
}
