package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/akulkarni/0perator/internal/mcp"
)

// InitOptions contains options for the Init command
type InitOptions struct {
	DevMode bool     // Use 'go run' instead of compiled binary for 0perator MCP
	Clients []string // Client names to configure (skips interactive selection if provided)
}

// Init initializes 0perator by installing tiger-cli and configuring MCP servers
func Init(options InitOptions) error {
	startTime := time.Now()

	printBanner()

	if options.DevMode {
		fmt.Println(accent("  [DEV MODE]") + " Using 'go run' for 0perator MCP")
		fmt.Println()
	}

	// Check and install tiger-cli
	if err := ensureTigerCLI(); err != nil {
		return err
	}

	// Authenticate with Tiger Cloud
	if err := ensureTigerAuth(); err != nil {
		return err
	}

	// Select IDE(s) to configure
	var selectedIDEs []mcp.ClientInfo
	var err error
	if len(options.Clients) > 0 {
		// Use clients from command line
		selectedIDEs, err = resolveClients(options.Clients)
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(accent("[3/4]") + " IDE Configuration")
		fmt.Println()
		for _, client := range selectedIDEs {
			fmt.Printf("  %s %s\n", accent("✓"), client.Name)
		}
	} else {
		// Interactive selection
		selectedIDEs, err = selectIDEs()
		if err != nil {
			return err
		}
	}

	// Install MCP servers for each selected IDE
	if err := installMCPServers(selectedIDEs, options.DevMode); err != nil {
		return err
	}

	totalTime := time.Since(startTime)
	printSuccess(totalTime, options.DevMode)
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

// resolveClients converts client names to ClientInfo structs
func resolveClients(clientNames []string) ([]mcp.ClientInfo, error) {
	supportedClients := mcp.GetSupportedClients()

	// Build lookup map
	clientMap := make(map[string]mcp.ClientInfo)
	for _, c := range supportedClients {
		clientMap[c.ClientName] = c
	}

	var result []mcp.ClientInfo
	for _, name := range clientNames {
		client, ok := clientMap[name]
		if !ok {
			var validNames []string
			for _, c := range supportedClients {
				validNames = append(validNames, c.ClientName)
			}
			return nil, fmt.Errorf("unknown client %q, valid clients: %v", name, validNames)
		}
		result = append(result, client)
	}

	return result, nil
}

// clientSelectModel represents the Bubble Tea model for client selection
type clientSelectModel struct {
	clients      []mcp.ClientInfo
	cursor       int
	selected     map[int]bool
	numberBuffer string
	done         bool
	cancelled    bool
}

func (m clientSelectModel) Init() tea.Cmd {
	return nil
}

func (m clientSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			m.numberBuffer = ""
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.numberBuffer = ""
			if m.cursor < len(m.clients)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		case "a":
			// Select all
			for i := range m.clients {
				m.selected[i] = true
			}
		case "n":
			// Select none
			for i := range m.clients {
				m.selected[i] = false
			}
		case "backspace":
			if len(m.numberBuffer) > 0 {
				m.updateNumberBuffer(m.numberBuffer[:len(m.numberBuffer)-1])
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.updateNumberBuffer(m.numberBuffer + msg.String())
		}
	}
	return m, nil
}

func (m *clientSelectModel) updateNumberBuffer(newBuffer string) {
	if newBuffer == "" {
		m.numberBuffer = newBuffer
		return
	}

	num, err := strconv.Atoi(newBuffer)
	if err != nil {
		return
	}

	index := num - 1
	if index >= 0 && index < len(m.clients) {
		m.numberBuffer = newBuffer
		m.cursor = index
	}
}

func (m clientSelectModel) View() string {
	s := "  Select IDE(s) to configure:\n\n"

	for i, client := range m.clients {
		cursor := " "
		if m.cursor == i {
			cursor = accent(">")
		}

		checked := " "
		if m.selected[i] {
			checked = accent("✓")
		}

		s += fmt.Sprintf("  %s [%s] %d. %s\n", cursor, checked, i+1, client.Name)
	}

	if m.numberBuffer != "" {
		s += fmt.Sprintf("\n  Typing: %s", m.numberBuffer)
	}

	s += "\n  " + accent("↑/↓") + " navigate  " + accent("space") + " toggle  " + accent("a") + " all  " + accent("n") + " none  " + accent("enter") + " confirm"
	return s
}

func selectIDEs() ([]mcp.ClientInfo, error) {
	fmt.Println()
	fmt.Println(accent("[3/4]") + " IDE Configuration")
	fmt.Println()

	supportedClients := mcp.GetSupportedClients()

	model := clientSelectModel{
		clients:  supportedClients,
		cursor:   0,
		selected: make(map[int]bool),
	}

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run IDE selection: %w", err)
	}

	result := finalModel.(clientSelectModel)
	if result.cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}

	// Collect selected clients
	var selectedClients []mcp.ClientInfo
	for i, client := range supportedClients {
		if result.selected[i] {
			selectedClients = append(selectedClients, client)
		}
	}

	if len(selectedClients) == 0 {
		return nil, fmt.Errorf("no IDE selected")
	}

	fmt.Println()
	return selectedClients, nil
}

func installMCPServers(clients []mcp.ClientInfo, devMode bool) error {
	fmt.Println()
	fmt.Println(accent("[4/4]") + " Installing MCP servers...")
	fmt.Println()

	for _, client := range clients {
		fmt.Printf("  Configuring %s\n", accent(client.Name))

		// Install Tiger MCP
		tigerStart := time.Now()

		// Suppress tiger MCP output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		os.Stdout = nil
		os.Stderr = nil

		err := mcp.InstallTigerMCP(client.ClientName)

		os.Stdout = oldStdout
		os.Stderr = oldStderr

		tigerElapsed := time.Since(tigerStart)
		if err != nil {
			fmt.Printf("    %s Tiger MCP (%.1fs)\n", accent("✗"), tigerElapsed.Seconds())
			return fmt.Errorf("failed to install Tiger MCP for %s: %w", client.Name, err)
		}
		fmt.Printf("    %s Tiger MCP (%.1fs)\n", accent("✓"), tigerElapsed.Seconds())

		// Install 0perator MCP
		opStart := time.Now()

		opts := mcp.Install0peratorMCPOptions{DevMode: devMode}
		if err := mcp.Install0peratorMCP(client.ClientName, opts); err != nil {
			fmt.Printf("    %s 0perator MCP (%.1fs)\n", accent("✗"), time.Since(opStart).Seconds())
			return fmt.Errorf("failed to install 0perator MCP for %s: %w", client.Name, err)
		}

		opElapsed := time.Since(opStart)
		fmt.Printf("    %s 0perator MCP (%.1fs)\n", accent("✓"), opElapsed.Seconds())
		fmt.Println()
	}

	return nil
}

func printSuccess(totalTime time.Duration, devMode bool) {
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println("  " + accent(fmt.Sprintf("✨ All set! (%.1fs)", totalTime.Seconds())))
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  Each IDE now has:")
	fmt.Println("    • Tiger MCP (database, docs, best practices)")
	if devMode {
		fmt.Println("    • 0perator MCP (dev mode: using 'go run')")
	} else {
		fmt.Println("    • 0perator MCP (app scaffolding, deployment)")
	}
	fmt.Println()
	fmt.Println("  Next: Restart your IDE and try")
	fmt.Println("        \"Create a task management app\"")
	fmt.Println()
	fmt.Println("  Docs: " + accent("https://0p.dev/docs"))
	fmt.Println()
}
