package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akulkarni/0perator/internal/templates"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the 0perator MCP server
type Server struct {
	mcpServer *server.MCPServer
}

// New creates a new 0perator MCP server
func New() *Server {
	s := &Server{}

	// Create MCP server with metadata
	s.mcpServer = server.NewMCPServer(
		"0perator",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	s.registerTools()

	return s
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Serve()
}

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	// Create app tool
	s.mcpServer.AddTool(
		mcp.NewTool("create_app",
			mcp.WithDescription("Scaffold a new application from a template. For databases, use Tiger MCP's service_create tool."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the application (used for directory name)"),
			),
			mcp.WithString("template",
				mcp.DefaultValue("web-node"),
				mcp.Description("Template to use: web-node (default), api-node, cli-node"),
			),
			mcp.WithString("description",
				mcp.Optional(),
				mcp.Description("Brief description of what the app does (helps customize scaffolding)"),
			),
			mcp.WithString("database_url",
				mcp.Optional(),
				mcp.Description("Database connection string (use Tiger MCP's service_create to provision database first)"),
			),
		),
		s.handleCreateApp,
	)

	// Deploy local tool
	s.mcpServer.AddTool(
		mcp.NewTool("deploy_local",
			mcp.WithDescription("Deploy an application locally (bare process, no containers)"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the application directory"),
			),
			mcp.WithNumber("port",
				mcp.Optional(),
				mcp.Description("Port to run on (auto-assigned if not provided)"),
			),
		),
		s.handleDeployLocal,
	)
}

// handleCreateApp handles the create_app tool
func (s *Server) handleCreateApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments
	name, _ := request.Params.Arguments["name"].(string)
	templateName, _ := request.Params.Arguments["template"].(string)
	description, _ := request.Params.Arguments["description"].(string)
	databaseURL, _ := request.Params.Arguments["database_url"].(string)

	// Default to web-node if not specified
	if templateName == "" {
		templateName = "web-node"
	}

	// Get current working directory for output
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Scaffold the application
	opts := templates.ScaffoldOptions{
		AppName:     name,
		Description: description,
		DatabaseURL: databaseURL,
		OutputDir:   cwd,
	}

	if err := templates.Scaffold(templateName, opts); err != nil {
		return nil, fmt.Errorf("failed to scaffold application: %w", err)
	}

	// Build success message
	appPath := filepath.Join(cwd, name)
	result := fmt.Sprintf(`✓ Application scaffolded successfully!

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
		name,
		templateName,
		appPath,
		name,
		func() string {
			if databaseURL != "" {
				return "  • PostgreSQL integration\n"
			}
			return ""
		}(),
	)

	return mcp.NewToolResultText(result), nil
}

// handleDeployLocal handles the deploy_local tool
func (s *Server) handleDeployLocal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments
	path, _ := request.Params.Arguments["path"].(string)
	port, _ := request.Params.Arguments["port"].(float64)

	// TODO: Implement actual local deployment
	result := fmt.Sprintf("Deploying app from: %s\nPort: %.0f\n\nTODO: Implement local deployment",
		path, port)

	return mcp.NewToolResultText(result), nil
}
