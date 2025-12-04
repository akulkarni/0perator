package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the 0perator MCP server
type Server struct {
	mcpServer *mcp.Server
}

// New creates a new 0perator MCP server
func New() (*Server, error) {
	s := &Server{}

	// Create MCP server with metadata and instructions
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "0perator",
		Version: "2.0.0", // Version 2: Uses Tiger CLI directly
	}, &mcp.ServerOptions{
		Instructions: `When the user asks to build a web application, SaaS app, or any app: use the view_skill tool for the skill named create-app.`,
	})

	// Register direct tools - these are what Claude sees
	if err := s.registerDirectTools(); err != nil {
		return nil, err
	}

	return s, nil
}

// Start starts the MCP server (stdio mode)
func (s *Server) Start() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
