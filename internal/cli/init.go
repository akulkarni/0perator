package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/akulkarni/0perator/internal/mcp"
)

// Init initializes 0perator by installing tiger-cli and configuring MCP servers
func Init() error {
	startTime := time.Now()

	printBanner()

	// Check and install tiger-cli
	if err := ensureTigerCLI(); err != nil {
		return err
	}

	// Authenticate with Tiger Cloud
	if err := ensureTigerAuth(); err != nil {
		return err
	}

	// Select IDE(s) to configure
	selectedIDEs, err := selectIDEs()
	if err != nil {
		return err
	}

	// Install MCP servers for each selected IDE
	if err := installMCPServers(selectedIDEs); err != nil {
		return err
	}

	totalTime := time.Since(startTime)
	printSuccess(selectedIDEs, totalTime)
	return nil
}

func printBanner() {
	// ASCII art with "0" accented
	fmt.Println()
	fmt.Println("     ___                       _")
	fmt.Println("    / " + accent("_") + " \\ _ __   ___ _ __ __ _| |_ ___  _ __")
	fmt.Println("   | " + accent("|") + " | | '_ \\ / _ \\ '__/ _` | __/ _ \\| '__|")
	fmt.Println("   | " + accent("|_|") + " | |_) |  __/ | | (_| | || (_) | |")
	fmt.Println("    \\" + accent("___") + "/| .__/ \\___|_|  \\__,_|\\__\\___/|_|")
	fmt.Println("         |_|")
	fmt.Println()
	fmt.Println("  " + accent("Infrastructure for AI agents"))
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────")
}

func ensureTigerCLI() error {
	fmt.Println()
	fmt.Println(accent("[1/4]") + " Checking dependencies...")

	// Check if tiger is installed
	if _, err := exec.LookPath("tiger"); err == nil {
		fmt.Println("      ✓ tiger-cli found")
		return nil
	}

	fmt.Println("      tiger-cli not found " + accent("→") + " installing now")
	fmt.Print("      ")

	start := time.Now()

	// Install tiger-cli
	cmd := exec.Command("sh", "-c", "curl -fsSL https://cli.tiger.build | sh")
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Show progress animation while installing
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	// Simulate progress bar
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	progress := 0.0
	var installErr error

progressLoop:
	for {
		select {
		case installErr = <-done:
			// Show full progress bar
			fmt.Print("\r      " + fullProgressBar(32))
			elapsed := time.Since(start)
			fmt.Printf(" %.1fs\n", elapsed.Seconds())
			if installErr != nil {
				return fmt.Errorf("failed to install tiger-cli: %w", installErr)
			}
			fmt.Println("      ✓ tiger-cli installed")
			break progressLoop
		case <-ticker.C:
			// Increment progress (slower at end to feel realistic)
			if progress < 0.7 {
				progress += 0.05
			} else if progress < 0.9 {
				progress += 0.02
			} else if progress < 0.95 {
				progress += 0.01
			}
			fmt.Print("\r      " + progressBar(32, progress))
		}
	}

	return nil
}

func ensureTigerAuth() error {
	fmt.Println()
	fmt.Println(accent("[2/4]") + " Authentication")

	// Check if already authenticated
	cmd := exec.Command("tiger", "service", "list")
	cmd.Stderr = nil
	cmd.Stdout = nil
	if err := cmd.Run(); err == nil {
		// Try to get user email
		emailCmd := exec.Command("tiger", "auth", "whoami")
		if output, err := emailCmd.Output(); err == nil {
			email := strings.TrimSpace(string(output))
			if email != "" {
				fmt.Printf("      ✓ Already authenticated as %s\n", email)
				return nil
			}
		}
		fmt.Println("      ✓ Already authenticated with Tiger Cloud")
		return nil
	}

	// Not authenticated, prompt user
	fmt.Print("      ? Authenticate with Tiger Cloud (opens browser) [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" && response != "yes" {
		return fmt.Errorf("authentication required to continue")
	}

	fmt.Println("      " + accent("↓") + " Opening browser...")
	start := time.Now()

	cmd = exec.Command("tiger", "auth", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Print("      " + fullProgressBar(32))
	fmt.Printf(" %.1fs\n", elapsed.Seconds())

	// Try to get user email
	emailCmd := exec.Command("tiger", "auth", "whoami")
	if output, err := emailCmd.Output(); err == nil {
		email := strings.TrimSpace(string(output))
		if email != "" {
			fmt.Printf("      ✓ Authenticated as %s\n", email)
			return nil
		}
	}

	fmt.Println("      ✓ Authentication successful!")
	return nil
}

func selectIDEs() ([]mcp.IDEClient, error) {
	fmt.Println()
	fmt.Println(accent("[3/4]") + " IDE Configuration")
	supportedIDEs := mcp.GetSupportedIDEs()

	fmt.Println("      ? Select IDE(s) to configure:")
	for i, ide := range supportedIDEs {
		fmt.Printf("        " + accent(fmt.Sprintf("%d)", i+1)) + " %s\n", ide)
	}
	fmt.Println()
	fmt.Print("      Enter numbers separated by commas (e.g., 1,2) or press Enter for all: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(response)

	// If empty, select all
	if response == "" {
		return supportedIDEs, nil
	}

	// Parse comma-separated numbers
	var selected []mcp.IDEClient
	parts := strings.Split(response, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var idx int
		if _, err := fmt.Sscanf(part, "%d", &idx); err != nil {
			return nil, fmt.Errorf("invalid selection: %s", part)
		}
		if idx < 1 || idx > len(supportedIDEs) {
			return nil, fmt.Errorf("selection out of range: %d", idx)
		}
		selected = append(selected, supportedIDEs[idx-1])
	}

	return selected, nil
}

func installMCPServers(ides []mcp.IDEClient) error {
	fmt.Println()
	fmt.Println(accent("[4/4]") + " Installing MCP servers...")
	fmt.Println()

	for _, ide := range ides {
		fmt.Printf("      %s\n", ide)

		// Install Tiger MCP
		tigerStart := time.Now()
		fmt.Print("        Tiger MCP      ")

		// Redirect tiger output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		os.Stdout = nil
		os.Stderr = nil

		err := mcp.InstallTigerMCP(ide)

		os.Stdout = oldStdout
		os.Stderr = oldStderr

		if err != nil {
			fmt.Println()
			return fmt.Errorf("failed to install Tiger MCP for %s: %w", ide, err)
		}

		tigerElapsed := time.Since(tigerStart)
		fmt.Print(fullProgressBar(32))
		fmt.Printf(" %.1fs ✓\n", tigerElapsed.Seconds())

		// Install 0perator MCP
		opStart := time.Now()
		fmt.Print("        0perator MCP   ")

		if err := mcp.Install0peratorMCP(ide); err != nil {
			fmt.Println()
			return fmt.Errorf("failed to install 0perator MCP for %s: %w", ide, err)
		}

		opElapsed := time.Since(opStart)
		fmt.Print(fullProgressBar(32))
		fmt.Printf(" %.1fs ✓\n", opElapsed.Seconds())
		fmt.Println()
	}

	return nil
}

func printSuccess(ides []mcp.IDEClient, totalTime time.Duration) {
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println("  " + accent(fmt.Sprintf("✨ All set! (%.1fs)", totalTime.Seconds())))
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  Each IDE now has:")
	fmt.Println("    • Tiger MCP (database, docs, best practices)")
	fmt.Println("    • 0perator MCP (app scaffolding, deployment)")
	fmt.Println()
	fmt.Println("  Next: Restart your IDE and try")
	fmt.Println("        \"Create a task management app\"")
	fmt.Println()
	fmt.Println("  Docs: " + accent("https://0p.dev/docs"))
	fmt.Println()
}
