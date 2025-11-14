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
	fmt.Println()
	fmt.Println(accent("     ██████╗ ██████╗ ███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗ "))
	fmt.Println(accent("    ██╔═████╗██╔══██╗██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗"))
	fmt.Println(accent("    ██║██╔██║██████╔╝█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝"))
	fmt.Println(accent("    ████╔╝██║██╔═══╝ ██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗"))
	fmt.Println(accent("    ╚██████╔╝██║     ███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║"))
	fmt.Println(accent("     ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝"))
	fmt.Println()
	fmt.Println(accent("               Infrastructure for AI native development"))
	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
}

func ensureTigerCLI() error {
	fmt.Println()
	fmt.Println(accent("[1/4]") + " Checking dependencies...")
	fmt.Println()

	// Check if tiger is installed
	if _, err := exec.LookPath("tiger"); err == nil {
		fmt.Println("  " + accent("✓") + " tiger-cli found")
		return nil
	}

	fmt.Println("  Installing tiger-cli...")

	start := time.Now()

	// Install tiger-cli
	cmd := exec.Command("sh", "-c", "curl -fsSL https://cli.tiger.build | sh")
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Show spinner while installing
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	var installErr error

progressLoop:
	for {
		select {
		case installErr = <-done:
			elapsed := time.Since(start)
			if installErr != nil {
				fmt.Printf("  %s Installation failed (%.1fs)\n", accent("✗"), elapsed.Seconds())
				return fmt.Errorf("failed to install tiger-cli: %w", installErr)
			}
			fmt.Printf("  %s tiger-cli installed (%.1fs)\n", accent("✓"), elapsed.Seconds())
			break progressLoop
		case <-ticker.C:
			fmt.Printf("\r  %s Installing...", accent(spinner[spinnerIdx]))
			spinnerIdx = (spinnerIdx + 1) % len(spinner)
		}
	}
	fmt.Print("\r") // Clear spinner line

	return nil
}

func ensureTigerAuth() error {
	fmt.Println()
	fmt.Println(accent("[2/4]") + " Authentication")
	fmt.Println()

	// Check if already authenticated
	statusCmd := exec.Command("tiger", "auth", "status")
	output, err := statusCmd.CombinedOutput()
	if err == nil && strings.Contains(string(output), "Logged in") {
		// Parse project ID if available
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Project ID") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					projectID := strings.TrimSpace(parts[1])
					fmt.Printf("  %s Authenticated (Project: %s)\n", accent("✓"), projectID)
					return nil
				}
			}
		}
		fmt.Println("  " + accent("✓") + " Authenticated with Tiger Cloud")
		return nil
	}

	// Not authenticated, prompt user
	fmt.Print("  Authenticate with Tiger Cloud (opens browser) [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" && response != "yes" {
		return fmt.Errorf("authentication required to continue")
	}

	fmt.Println()
	fmt.Println("  Opening browser...")
	start := time.Now()

	loginCmd := exec.Command("tiger", "auth", "login")
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	loginCmd.Stdin = os.Stdin

	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("  %s Authentication complete (%.1fs)\n", accent("✓"), elapsed.Seconds())

	return nil
}

func selectIDEs() ([]mcp.IDEClient, error) {
	fmt.Println()
	fmt.Println(accent("[3/4]") + " IDE Configuration")
	fmt.Println()
	supportedIDEs := mcp.GetSupportedIDEs()

	fmt.Println("  Select IDE(s) to configure:")
	for i, ide := range supportedIDEs {
		fmt.Printf("    " + accent(fmt.Sprintf("%d.", i+1)) + " %s\n", ide)
	}
	fmt.Println()
	fmt.Print("  Enter selections (e.g., 1,2) or press Enter for all: ")

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
		fmt.Printf("  Configuring %s\n", accent(string(ide)))

		// Install Tiger MCP
		tigerStart := time.Now()

		// Suppress tiger MCP output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		os.Stdout = nil
		os.Stderr = nil

		err := mcp.InstallTigerMCP(ide)

		os.Stdout = oldStdout
		os.Stderr = oldStderr

		tigerElapsed := time.Since(tigerStart)
		if err != nil {
			fmt.Printf("    %s Tiger MCP (%.1fs)\n", accent("✗"), tigerElapsed.Seconds())
			return fmt.Errorf("failed to install Tiger MCP for %s: %w", ide, err)
		}
		fmt.Printf("    %s Tiger MCP (%.1fs)\n", accent("✓"), tigerElapsed.Seconds())

		// Install 0perator MCP
		opStart := time.Now()

		if err := mcp.Install0peratorMCP(ide); err != nil {
			fmt.Printf("    %s 0perator MCP (%.1fs)\n", accent("✗"), time.Since(opStart).Seconds())
			return fmt.Errorf("failed to install 0perator MCP for %s: %w", ide, err)
		}

		opElapsed := time.Since(opStart)
		fmt.Printf("    %s 0perator MCP (%.1fs)\n", accent("✓"), opElapsed.Seconds())
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
