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
	ClaudeCode: "~/.claude.json",
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
	cmd := exec.Command("tiger", "mcp", "install", string(client), "--no-backup")
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
func Install0peratorMCP(client IDEClient, options Install0peratorMCPOptions) error {

	configPath, err := getConfigPath(client)
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config or create new one
	var config MCPConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// File doesn't exist, create new config
		config = MCPConfig{
			MCPServers: make(map[string]MCPServer),
		}
	} else {
		// File exists, parse it
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Add 0perator MCP server
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	var server MCPServer
	if options.DevMode {
		// In dev mode, use 'go run' with the project directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		mainPath := filepath.Join(cwd, "cmd", "0perator-mcp", "main.go")
		if _, err := os.Stat(mainPath); err != nil {
			return fmt.Errorf("dev mode requires running from the 0perator repository root (expected %s)", mainPath)
		}
		server = MCPServer{
			Command: "go",
			Args:    []string{"run", mainPath},
		}
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
		server = MCPServer{
			Command: operatorPath,
			Args:    []string{},
		}
	}

	config.MCPServers["0perator"] = server

	// Write back to file with pretty formatting
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InstallBoth installs both Tiger and 0perator MCP servers for the given IDE
func InstallBoth(client IDEClient, options Install0peratorMCPOptions) error {
	// Check if Tiger MCP config already exists
	configPath, err := getConfigPath(client)
	if err != nil {
		return err
	}

	tigerExists := false
	if data, err := os.ReadFile(configPath); err == nil {
		var config MCPConfig
		if json.Unmarshal(data, &config) == nil {
			if _, exists := config.MCPServers["tiger"]; exists {
				tigerExists = true
			}
		}
	}

	// Only install Tiger if it doesn't exist
	if !tigerExists {
		if err := InstallTigerMCP(client); err != nil {
			return err
		}
	}

	// Add 0perator to the config file
	if err := Install0peratorMCP(client, options); err != nil {
		return err
	}

	return nil
}

// GetSupportedIDEs returns a list of supported IDE clients in a consistent order
func GetSupportedIDEs() []IDEClient {
	// Return in a fixed order for consistent UI
	return []IDEClient{
		ClaudeCode,
		Cursor,
		Windsurf,
	}
}
