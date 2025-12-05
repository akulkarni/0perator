package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/timescale/tiger-cli/pkg/mcpinstall"
)

// ClientInfo re-exports mcpinstall.ClientInfo for convenience
type ClientInfo = mcpinstall.ClientInfo

// GetSupportedClients returns information about all supported MCP clients
func GetSupportedClients() []ClientInfo {
	return mcpinstall.SupportedClients()
}

// InstallTigerMCP installs Tiger MCP for the given IDE client
func InstallTigerMCP(clientName string) error {
	cmd := exec.Command("tiger", "mcp", "install", clientName, "--no-backup")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's just because tiger is already installed
		if strings.Contains(string(output), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to install Tiger MCP: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Install0peratorMCPOptions contains options for Install0peratorMCP
type Install0peratorMCPOptions struct {
	DevMode bool // Use 'go run' instead of compiled binary
}

// Install0peratorMCP adds 0perator MCP server to the IDE's config file
func Install0peratorMCP(clientName string, options Install0peratorMCPOptions) error {
	var command string
	var args []string

	if options.DevMode {
		// In dev mode, use 'go run' with the project directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		scriptPath := filepath.Join(cwd, "scripts", "run-source.sh")
		if _, err := os.Stat(scriptPath); err != nil {
			return fmt.Errorf("dev mode requires running from the 0perator repository root (expected %s)", scriptPath)
		}
		command = "sh"
		args = []string{scriptPath}
	} else {
		// Get the full path to the 0perator binary
		operatorPath, err := exec.LookPath("0perator")
		if err != nil {
			// If not in PATH, use the current executable path
			operatorPath, err = os.Executable()
			if err != nil {
				return fmt.Errorf("failed to determine 0perator binary path: %w", err)
			}
		}
		command = operatorPath
		args = []string{}
	}

	return mcpinstall.Install(mcpinstall.Options{
		ClientName: clientName,
		ServerName: "0perator",
		Command:    command,
		Args:       args,
	})
}

// InstallBoth installs both Tiger and 0perator MCP servers for the given IDE
func InstallBoth(clientName string, options Install0peratorMCPOptions) error {
	if err := InstallTigerMCP(clientName); err != nil {
		return err
	}

	if err := Install0peratorMCP(clientName, options); err != nil {
		return err
	}

	return nil
}
