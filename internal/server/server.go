package server

import (
	"context"

	"github.com/akulkarni/0perator/internal/operator"
	"github.com/akulkarni/0perator/internal/runtime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the 0perator MCP server
type Server struct {
	mcpServer      *mcp.Server
	processManager *runtime.ProcessManager
	operator       *operator.Operator
}

// New creates a new 0perator MCP server
func New() *Server {
	s := &Server{
		processManager: runtime.NewProcessManager(),
		operator:       operator.New(),
	}

	// Create MCP server with metadata
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "0perator",
		Version: "2.0.0", // Version 2: Uses Tiger CLI directly
	}, nil)

	// Register direct tools - these are what Claude sees
	s.registerDirectTools()

	// Register the operator tool (minimal wrapper for backward compatibility)
	s.registerOperatorTool()

	// Debug: Log registered tools count
	// fmt.Fprintf(os.Stderr, "Debug: Registered tools in MCP server\n")

	return s
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
