package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Process represents a running local application
type Process struct {
	ID      string
	Path    string
	Port    int
	Command *exec.Cmd
	LogFile *os.File
	mu      sync.Mutex
	cancel  context.CancelFunc
}

// ProcessManager manages local application processes
type ProcessManager struct {
	processes map[string]*Process
	mu        sync.RWMutex
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

// Deploy starts a local application
func (pm *ProcessManager) Deploy(ctx context.Context, appPath string, port int) (*Process, error) {
	// Validate path exists
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("application path does not exist: %s", absPath)
	}

	// Detect app type
	appType, err := detectAppType(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect app type: %w", err)
	}

	// Allocate port if not specified
	if port == 0 {
		port, err = findFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to allocate port: %w", err)
		}
	} else {
		// Validate requested port is free
		if !isPortFree(port) {
			return nil, fmt.Errorf("port %d is already in use", port)
		}
	}

	// Install dependencies if needed
	if err := installDependencies(absPath, appType); err != nil {
		return nil, fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Create process context
	processCtx, cancel := context.WithCancel(ctx)

	// Create log file
	logDir := filepath.Join(os.TempDir(), "0perator-logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	processID := fmt.Sprintf("%s-%d", filepath.Base(absPath), time.Now().Unix())
	logFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("%s.log", processID)))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create command based on app type
	cmd, err := createStartCommand(absPath, appType, port)
	if err != nil {
		cancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to create start command: %w", err)
	}

	cmd.Dir = absPath

	// Set up logging
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create process
	process := &Process{
		ID:      processID,
		Path:    absPath,
		Port:    port,
		Command: cmd,
		LogFile: logFile,
		cancel:  cancel,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	// Stream logs in background
	go streamLogs(processCtx, stdout, logFile, "stdout")
	go streamLogs(processCtx, stderr, logFile, "stderr")

	// Monitor process
	go func() {
		cmd.Wait()
		cancel()
	}()

	// Register process
	pm.mu.Lock()
	pm.processes[processID] = process
	pm.mu.Unlock()

	// Wait for process to be healthy
	if err := waitForHealth(processCtx, port, 30*time.Second); err != nil {
		pm.Stop(processID)
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return process, nil
}

// Stop stops a running process
func (pm *ProcessManager) Stop(processID string) error {
	pm.mu.Lock()
	process, exists := pm.processes[processID]
	if !exists {
		pm.mu.Unlock()
		return fmt.Errorf("process not found: %s", processID)
	}
	delete(pm.processes, processID)
	pm.mu.Unlock()

	process.mu.Lock()
	defer process.mu.Unlock()

	// Cancel context
	process.cancel()

	// Kill process if still running
	if process.Command.Process != nil {
		if err := process.Command.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	// Close log file
	if process.LogFile != nil {
		process.LogFile.Close()
	}

	return nil
}

// GetProcess returns a process by ID
func (pm *ProcessManager) GetProcess(processID string) (*Process, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	process, exists := pm.processes[processID]
	return process, exists
}

// ListProcesses returns all running processes
func (pm *ProcessManager) ListProcesses() []*Process {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	processes := make([]*Process, 0, len(pm.processes))
	for _, p := range pm.processes {
		processes = append(processes, p)
	}
	return processes
}

// detectAppType detects the application type
func detectAppType(path string) (string, error) {
	// Check for Node.js
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return "nodejs", nil
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		return "python", nil
	}

	// Check for Go
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return "go", nil
	}

	return "", fmt.Errorf("unsupported application type (no package.json, requirements.txt, or go.mod found)")
}

// installDependencies installs application dependencies
func installDependencies(path, appType string) error {
	switch appType {
	case "nodejs":
		// Check if node_modules exists
		if _, err := os.Stat(filepath.Join(path, "node_modules")); err == nil {
			// Already installed
			return nil
		}

		// Run npm install
		cmd := exec.Command("npm", "install")
		cmd.Dir = path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case "python":
		// TODO: Implement Python dependency installation
		return nil

	case "go":
		// Go modules are typically handled by go run/build
		return nil

	default:
		return fmt.Errorf("unsupported app type: %s", appType)
	}
}

// createStartCommand creates the command to start the application
func createStartCommand(path, appType string, port int) (*exec.Cmd, error) {
	switch appType {
	case "nodejs":
		// Check package.json for dev script
		cmd := exec.Command("npm", "run", "dev")
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
		return cmd, nil

	case "python":
		// TODO: Implement Python start command
		return nil, fmt.Errorf("Python apps not yet supported")

	case "go":
		// TODO: Implement Go start command
		return nil, fmt.Errorf("Go apps not yet supported")

	default:
		return nil, fmt.Errorf("unsupported app type: %s", appType)
	}
}

// streamLogs streams process logs to a file
func streamLogs(ctx context.Context, reader io.Reader, logFile *os.File, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, prefix, line)
			logFile.WriteString(logLine)
		}
	}
}

// findFreePort finds an available port
func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

// isPortFree checks if a port is available
func isPortFree(port int) bool {
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// waitForHealth waits for the application to respond to health checks
func waitForHealth(ctx context.Context, port int, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://localhost:%d", port)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("application did not become healthy within %v", timeout)
}

// GetLogs returns the last N lines of logs for a process
func GetLogs(logFilePath string, numLines int) ([]string, error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last N lines
	start := 0
	if len(lines) > numLines {
		start = len(lines) - numLines
	}

	return lines[start:], nil
}

// TailLogs returns a formatted string of recent logs
func TailLogs(logFilePath string, numLines int) string {
	lines, err := GetLogs(logFilePath, numLines)
	if err != nil {
		return fmt.Sprintf("Error reading logs: %v", err)
	}

	if len(lines) == 0 {
		return "No logs available yet"
	}

	return strings.Join(lines, "\n")
}
