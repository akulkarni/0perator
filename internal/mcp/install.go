package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IDEClient represents supported IDE clients
type IDEClient string

const (
	ClaudeCode IDEClient = "claude-code"
	Cursor     IDEClient = "cursor"
	Windsurf   IDEClient = "windsurf"
)

// ideConfigPaths maps IDE clients to their MCP config file paths
// Paths with ~ will be expanded to user's home directory
var ideConfigPaths = map[IDEClient]string{
	ClaudeCode: "~/.config/claude/mcp.json",
	Cursor:     "~/.cursor/mcp.json",
	Windsurf:   "~/.windsurf/mcp.json",
}

// MCPConfig represents the MCP configuration file structure
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// getConfigPath returns the config file path for a given IDE client
func getConfigPath(client IDEClient) (string, error) {
	path, ok := ideConfigPaths[client]
	if !ok {
		return "", fmt.Errorf("unsupported IDE client: %s", client)
	}
	return expandPath(path)
}

// InstallTigerMCP installs Tiger MCP for the given IDE client
func InstallTigerMCP(client IDEClient) error {
	fmt.Printf("Installing Tiger MCP for %s...\n", client)

	cmd := exec.Command("tiger", "mcp", "install", string(client), "--no-backup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Tiger MCP: %w", err)
	}

	fmt.Printf("✓ Tiger MCP installed for %s\n", client)
	return nil
}

// Install0peratorMCP adds 0perator MCP server to the IDE's config file
func Install0peratorMCP(client IDEClient) error {
	fmt.Printf("Installing 0perator MCP for %s...\n", client)

	configPath, err := getConfigPath(client)
	if err != nil {
		return err
	}

	// Read existing config (tiger should have created this)
	var config MCPConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file (did tiger mcp install run?): %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Add 0perator MCP server
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	config.MCPServers["0perator"] = MCPServer{
		Command: "0perator",
		Args:    []string{"mcp", "start"},
		Env: map[string]string{
			"OPERATOR_CONFIG": filepath.Join(os.Getenv("HOME"), ".config", "0perator", "config.json"),
		},
	}

	// Write back to file with pretty formatting
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ 0perator MCP installed for %s\n", client)
	return nil
}

// InstallBoth installs both Tiger and 0perator MCP servers for the given IDE
func InstallBoth(client IDEClient) error {
	fmt.Printf("\n━━━ Configuring %s ━━━\n", client)

	// First install Tiger MCP (this creates/updates the config file)
	if err := InstallTigerMCP(client); err != nil {
		return err
	}

	// Then add 0perator to the same config file that tiger just created
	if err := Install0peratorMCP(client); err != nil {
		return err
	}

	return nil
}

// GetSupportedIDEs returns a list of supported IDE clients
func GetSupportedIDEs() []IDEClient {
	ides := make([]IDEClient, 0, len(ideConfigPaths))
	for ide := range ideConfigPaths {
		ides = append(ides, ide)
	}
	return ides
}
