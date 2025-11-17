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
		Version: "0.2.0",
	}, nil)

	// Register ONLY the new operator tool (action-based system)
	// Old template-based tools are deprecated
	// s.registerTools() // REMOVED - using action-based system now
	s.registerOperatorTool()

	// Register prompts (keep for now, might be useful for other purposes)
	s.registerPrompts()

	return s
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
