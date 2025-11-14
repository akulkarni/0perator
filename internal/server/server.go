package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akulkarni/0perator/internal/runtime"
	"github.com/akulkarni/0perator/internal/templates"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the 0perator MCP server
type Server struct {
	mcpServer      *mcp.Server
	processManager *runtime.ProcessManager
}

// New creates a new 0perator MCP server
func New() *Server {
	s := &Server{
		processManager: runtime.NewProcessManager(),
	}

	// Create MCP server with metadata
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "0perator",
		Version: "0.1.0",
	}, nil)

	// Register tools
	s.registerTools()

	return s
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	// Create app tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_app",
		Description: "Scaffold a new application from a template. For databases, use Tiger MCP's service_create tool.",
	}, s.handleCreateApp)

	// Deploy local tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "deploy_local",
		Description: "Deploy an application locally (bare process, no containers)",
	}, s.handleDeployLocal)

	// Stop local deployment
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "stop_local",
		Description: "Stop a locally deployed application",
	}, s.handleStopLocal)

	// List local deployments
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_local",
		Description: "List all locally deployed applications",
	}, s.handleListLocal)

	// Get logs for local deployment
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "logs_local",
		Description: "Get logs for a locally deployed application",
	}, s.handleLogsLocal)
}

// Input/Output types for create_app
type CreateAppInput struct {
	Name        string `json:"name" jsonschema:"Name of the application (used for directory name)"`
	Template    string `json:"template,omitempty" jsonschema:"Template to use: web-node (default), api-node, cli-node"`
	Description string `json:"description,omitempty" jsonschema:"Brief description of what the app does (helps customize scaffolding)"`
	DatabaseURL string `json:"database_url,omitempty" jsonschema:"Database connection string (use Tiger MCP's service_create to provision database first)"`
}

type CreateAppOutput struct {
	Message string `json:"message" jsonschema:"Success message with next steps"`
}

// handleCreateApp handles the create_app tool
func (s *Server) handleCreateApp(ctx context.Context, req *mcp.CallToolRequest, input CreateAppInput) (*mcp.CallToolResult, CreateAppOutput, error) {
	// Default to web-node if not specified
	templateName := input.Template
	if templateName == "" {
		templateName = "web-node"
	}

	// Get current working directory for output
	cwd, err := os.Getwd()
	if err != nil {
		return nil, CreateAppOutput{}, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Scaffold the application
	opts := templates.ScaffoldOptions{
		AppName:     input.Name,
		Description: input.Description,
		DatabaseURL: input.DatabaseURL,
		OutputDir:   cwd,
	}

	if err := templates.Scaffold(templateName, opts); err != nil {
		return nil, CreateAppOutput{}, fmt.Errorf("failed to scaffold application: %w", err)
	}

	// Build success message
	appPath := filepath.Join(cwd, input.Name)
	message := fmt.Sprintf(`✓ Application scaffolded successfully!

Name:        %s
Template:    %s
Location:    %s

Next steps:
  cd %s
  npm install
  npm run dev

The application will be available at http://localhost:3000

Template includes:
  • TypeScript with strict mode
  • Fastify web framework
  • Zod for validation
  • Vitest for testing
  • ESLint + Prettier
%s
Generated with 0perator`,
		input.Name,
		templateName,
		appPath,
		input.Name,
		func() string {
			if input.DatabaseURL != "" {
				return "  • PostgreSQL integration\n"
			}
			return ""
		}(),
	)

	return nil, CreateAppOutput{Message: message}, nil
}

// Input/Output types for deploy_local
type DeployLocalInput struct {
	Path string  `json:"path" jsonschema:"Path to the application directory"`
	Port float64 `json:"port,omitempty" jsonschema:"Port to run on (auto-assigned if not provided)"`
}

type DeployLocalOutput struct {
	Message   string `json:"message" jsonschema:"Success message with deployment details"`
	ProcessID string `json:"process_id" jsonschema:"Process ID for managing the deployment"`
	Port      int    `json:"port" jsonschema:"Port the application is running on"`
	URL       string `json:"url" jsonschema:"URL to access the application"`
}

// handleDeployLocal handles the deploy_local tool
func (s *Server) handleDeployLocal(ctx context.Context, req *mcp.CallToolRequest, input DeployLocalInput) (*mcp.CallToolResult, DeployLocalOutput, error) {
	// Convert port to int
	port := int(input.Port)

	// Deploy the application
	process, err := s.processManager.Deploy(ctx, input.Path, port)
	if err != nil {
		return nil, DeployLocalOutput{}, fmt.Errorf("deployment failed: %w", err)
	}

	// Get recent logs
	logs := runtime.TailLogs(process.LogFile.Name(), 20)

	// Build success message
	message := fmt.Sprintf(`✓ Application deployed successfully!

Process ID:  %s
Path:        %s
Port:        %d
URL:         http://localhost:%d
Log File:    %s

Recent logs:
%s

The application is now running. Use the URL above to access it.
To view more logs, check the log file.

Generated with 0perator`,
		process.ID,
		process.Path,
		process.Port,
		process.Port,
		process.LogFile.Name(),
		logs,
	)

	return nil, DeployLocalOutput{
		Message:   message,
		ProcessID: process.ID,
		Port:      process.Port,
		URL:       fmt.Sprintf("http://localhost:%d", process.Port),
	}, nil
}

// Input/Output types for stop_local
type StopLocalInput struct {
	ProcessID string `json:"process_id" jsonschema:"Process ID returned from deploy_local"`
}

type StopLocalOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
}

// handleStopLocal handles the stop_local tool
func (s *Server) handleStopLocal(ctx context.Context, req *mcp.CallToolRequest, input StopLocalInput) (*mcp.CallToolResult, StopLocalOutput, error) {
	// Stop the process
	if err := s.processManager.Stop(input.ProcessID); err != nil {
		return nil, StopLocalOutput{}, fmt.Errorf("failed to stop process: %w", err)
	}

	message := fmt.Sprintf(`✓ Application stopped successfully!

Process ID: %s

Generated with 0perator`, input.ProcessID)

	return nil, StopLocalOutput{Message: message}, nil
}

// Input/Output types for list_local
type ListLocalInput struct {
	// No input parameters needed
}

type ProcessInfo struct {
	ProcessID string `json:"process_id" jsonschema:"Process ID"`
	Path      string `json:"path" jsonschema:"Application path"`
	Port      int    `json:"port" jsonschema:"Port number"`
	URL       string `json:"url" jsonschema:"Application URL"`
	LogFile   string `json:"log_file" jsonschema:"Log file path"`
}

type ListLocalOutput struct {
	Message   string        `json:"message" jsonschema:"Formatted list of running applications"`
	Processes []ProcessInfo `json:"processes" jsonschema:"List of running processes"`
}

// handleListLocal handles the list_local tool
func (s *Server) handleListLocal(ctx context.Context, req *mcp.CallToolRequest, input ListLocalInput) (*mcp.CallToolResult, ListLocalOutput, error) {
	processes := s.processManager.ListProcesses()

	if len(processes) == 0 {
		return nil, ListLocalOutput{
			Message:   "No applications currently running locally.",
			Processes: []ProcessInfo{},
		}, nil
	}

	message := "Local Applications:\n\n"
	processInfos := make([]ProcessInfo, 0, len(processes))

	for i, proc := range processes {
		message += fmt.Sprintf("%d. Process ID: %s\n", i+1, proc.ID)
		message += fmt.Sprintf("   Path:       %s\n", proc.Path)
		message += fmt.Sprintf("   Port:       %d\n", proc.Port)
		message += fmt.Sprintf("   URL:        http://localhost:%d\n", proc.Port)
		message += fmt.Sprintf("   Log File:   %s\n\n", proc.LogFile.Name())

		processInfos = append(processInfos, ProcessInfo{
			ProcessID: proc.ID,
			Path:      proc.Path,
			Port:      proc.Port,
			URL:       fmt.Sprintf("http://localhost:%d", proc.Port),
			LogFile:   proc.LogFile.Name(),
		})
	}

	message += "Generated with 0perator"

	return nil, ListLocalOutput{
		Message:   message,
		Processes: processInfos,
	}, nil
}

// Input/Output types for logs_local
type LogsLocalInput struct {
	ProcessID string  `json:"process_id" jsonschema:"Process ID returned from deploy_local"`
	Lines     float64 `json:"lines,omitempty" jsonschema:"Number of log lines to retrieve (default: 50)"`
}

type LogsLocalOutput struct {
	Message string `json:"message" jsonschema:"Formatted log output"`
	Logs    string `json:"logs" jsonschema:"Raw log content"`
}

// handleLogsLocal handles the logs_local tool
func (s *Server) handleLogsLocal(ctx context.Context, req *mcp.CallToolRequest, input LogsLocalInput) (*mcp.CallToolResult, LogsLocalOutput, error) {
	lines := int(input.Lines)
	if lines <= 0 {
		lines = 50
	}

	// Get process
	process, exists := s.processManager.GetProcess(input.ProcessID)
	if !exists {
		return nil, LogsLocalOutput{}, fmt.Errorf("process not found: %s", input.ProcessID)
	}

	// Get logs
	logs := runtime.TailLogs(process.LogFile.Name(), lines)

	message := fmt.Sprintf(`Logs for Process: %s
Path: %s
Port: %d
Log File: %s

%s

Generated with 0perator`,
		process.ID,
		process.Path,
		process.Port,
		process.LogFile.Name(),
		logs,
	)

	return nil, LogsLocalOutput{
		Message: message,
		Logs:    logs,
	}, nil
}
