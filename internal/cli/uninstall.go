package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akulkarni/0perator/internal/mcp"
)

// Uninstall removes 0perator from the system
func Uninstall() error {
	printBanner()

	fmt.Println()
	fmt.Println(accent("[1/3]") + " Confirm uninstallation")
	fmt.Println()

	// Confirm with user
	fmt.Print("  Are you sure you want to uninstall 0perator? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println()
		fmt.Println("  Uninstall cancelled")
		fmt.Println()
		return nil
	}

	// Remove MCP server configurations
	fmt.Println()
	fmt.Println(accent("[2/3]") + " Removing MCP server configurations...")
	fmt.Println()

	ides := mcp.GetSupportedIDEs()
	for _, ide := range ides {
		if err := removeMCPConfig(ide); err != nil {
			fmt.Printf("  %s Failed to remove %s config: %v\n", accent("!"), ide, err)
		} else {
			fmt.Printf("  %s Removed %s configuration\n", accent("✓"), ide)
		}
	}

	// Remove binary
	fmt.Println()
	fmt.Println(accent("[3/3]") + " Removing binary...")
	fmt.Println()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	// Check if binary is in a standard location
	if strings.Contains(binaryPath, ".local/bin") || strings.Contains(binaryPath, "/usr/local/bin") {
		if err := os.Remove(binaryPath); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
		fmt.Printf("  %s Removed binary: %s\n", accent("✓"), binaryPath)
	} else {
		fmt.Printf("  %s Binary location: %s\n", accent("!"), binaryPath)
		fmt.Println("  Please remove manually if needed")
	}

	// Remove config directory (future use)
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "0perator")
	if _, err := os.Stat(configDir); err == nil {
		if err := os.RemoveAll(configDir); err != nil {
			fmt.Printf("  %s Failed to remove config directory: %v\n", accent("!"), err)
		} else {
			fmt.Printf("  %s Removed config directory\n", accent("✓"))
		}
	}

	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
	fmt.Println("  " + accent("✨ Uninstall complete!"))
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  0perator has been removed from your system.")
	fmt.Println("  Please restart your IDE for changes to take effect.")
	fmt.Println()

	return nil
}

// removeMCPConfig removes 0perator MCP server from IDE config
func removeMCPConfig(client mcp.IDEClient) error {
	configPath, err := getIDEConfigPath(client)
	if err != nil {
		return err
	}

	// Check if config file exists
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Config doesn't exist, nothing to remove
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if mcpServers exists
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok || mcpServers == nil {
		return nil // No MCP servers configured
	}

	// Remove 0perator entry
	if _, exists := mcpServers["0perator"]; !exists {
		return nil // 0perator not configured
	}

	delete(mcpServers, "0perator")

	// Write back to file
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getIDEConfigPath returns the config file path for a given IDE
func getIDEConfigPath(client mcp.IDEClient) (string, error) {
	var path string
	switch client {
	case mcp.ClaudeCode:
		path = "~/.claude.json"
	case mcp.Cursor:
		path = "~/.cursor/mcp.json"
	case mcp.Windsurf:
		path = "~/.windsurf/mcp.json"
	default:
		return "", fmt.Errorf("unsupported IDE client: %s", client)
	}

	// Expand home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	return path, nil
}
