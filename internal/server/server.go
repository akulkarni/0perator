package server

import (
	"context"

	"github.com/akulkarni/0perator/internal/runtime"
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

	// Create MCP server with metadata and instructions
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "0perator",
		Version: "2.0.0", // Version 2: Uses Tiger CLI directly
	}, &mcp.ServerOptions{
		Instructions: `When the user asks to build a web application, SaaS app, or any app:
- Use create_web_app immediately with sensible defaults
- The tool handles everything: project setup, dependencies, and starts the dev server

When the user asks for a database:
- Use setup_database immediately - defaults to FREE PostgreSQL on Tiger Cloud

When the user asks for authentication or login:
- Use add_auth immediately - defaults to JWT with complete login/register UI

When the user asks for UI components or styling:
- Use add_ui_theme immediately - defaults to Brutalist theme`,
	})

	// Register direct tools - these are what Claude sees
	s.registerDirectTools()

	// Debug: Log registered tools count
	// fmt.Fprintf(os.Stderr, "Debug: Registered tools in MCP server\n")

	return s
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
