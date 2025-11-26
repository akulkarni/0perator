package main

import (
	"fmt"
	"os"

	"github.com/akulkarni/0perator/internal/cli"
)

const version = "2.0.1"

func main() {
	if len(os.Args) < 2 {
		// No args: start MCP server (default behavior)
		startMCP()
		return
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("0perator %s\n", version)
	case "init":
		if err := cli.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "mcp":
		startMCP()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func startMCP() {
	if err := cli.MCPStart(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`0perator - Infrastructure for AI native development

Usage:
  0perator [command]

Commands:
  init      Configure IDEs with MCP servers
  mcp       Start the MCP server (default if no command)
  version   Print version information
  help      Show this help message

Examples:
  0perator init       # Set up your IDE
  0perator            # Start MCP server (for IDE use)`)
}
