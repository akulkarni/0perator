package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Uninstall removes 0perator from the system
func Uninstall() error {
	printBanner()

	fmt.Println()
	fmt.Println(accent("[1/2]") + " Confirm uninstallation")
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

	// Remove binary
	fmt.Println()
	fmt.Println(accent("[2/2]") + " Removing binary...")
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
	fmt.Println("  Please manually remove '0perator' from your IDE's MCP configuration.")
	fmt.Println()
	fmt.Println("  Restart your IDE for changes to take effect.")
	fmt.Println()

	return nil
}
