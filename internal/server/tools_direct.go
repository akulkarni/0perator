package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerDirectTools registers individual MCP tools for direct access
func (s *Server) registerDirectTools() {
	// Universal web app tool (handles all frameworks)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_web_app",
		Description: "🚀 Create any web application - Build an opinionated next.js app",
	}, s.handleCreateWebApp)

	// Universal database tool (handles all database types)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "setup_database",
		Description: "🗄️ Set up any database - PostgreSQL on Tiger Cloud (default, FREE). Auto-configures with schema, migrations, and connection handling. Use for any database request.",
	}, s.handleSetupDatabase)

	// Open app in browser - call this after all setup is complete
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "open_app",
		Description: "🌐 Open the app in browser. Call this AFTER all setup (database, auth, UI) is complete to show the user their running app.",
	}, s.handleOpenApp)
}

// Input/Output types for each tool

type CreateWebAppInput struct {
	Name string `json:"name" jsonschema:"Application name"`
}

type CreateWebAppOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (s *Server) handleCreateWebApp(ctx context.Context, req *mcp.CallToolRequest, input CreateWebAppInput) (*mcp.CallToolResult, CreateWebAppOutput, error) {
	// Set defaults
	if input.Name == "" {
		input.Name = "my-app"
	}

	err := tools.CreateNextJSApp(ctx, input.Name)

	if err != nil {
		return nil, CreateWebAppOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return nil, CreateWebAppOutput{
		Success: true,
		Message: fmt.Sprintf("Created app '%s'", input.Name),
		Path:    input.Name,
	}, nil
}

type SetupDatabaseInput struct {
	Name string `json:"name,omitempty" jsonschema:"Database name (default: app-db)"`
	Type string `json:"type,omitempty" jsonschema:"Database type: postgres (default)"`
}

type SetupDatabaseOutput struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	Type           string `json:"type"`
	ConnectionInfo string `json:"connection_info,omitempty"`
}

func (s *Server) handleSetupDatabase(ctx context.Context, req *mcp.CallToolRequest, input SetupDatabaseInput) (*mcp.CallToolResult, SetupDatabaseOutput, error) {
	// Set defaults
	if input.Name == "" {
		input.Name = "app-db"
	}
	if input.Type == "" {
		input.Type = "postgres" // Default to PostgreSQL for production readiness
	}

	switch input.Type {
	case "postgres":
		// Call the real implementation directly
		err := tools.SetupPostgresWithSchema(ctx, map[string]string{
			"name": input.Name,
		})

		if err != nil {
			return nil, SetupDatabaseOutput{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		return nil, SetupDatabaseOutput{
			Success:        true,
			Message:        fmt.Sprintf("Created PostgreSQL database '%s' on Tiger Cloud (free tier) with auto-schema", input.Name),
			Type:           "postgres",
			ConnectionInfo: "", // Connection info will be printed by the tool
		}, nil

	default:
		return nil, SetupDatabaseOutput{
			Success: false,
			Message: fmt.Sprintf("Database type '%s' not supported. Use 'postgres' (default).", input.Type),
		}, nil
	}
}

type SetupPostgresFreeInput struct {
	Name string `json:"name,omitempty" jsonschema:"Database name (default: app-db)"`
}

type SetupPostgresFreeOutput struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ConnectionInfo string `json:"connection_info,omitempty"`
}

func (s *Server) handleSetupPostgresFree(ctx context.Context, req *mcp.CallToolRequest, input SetupPostgresFreeInput) (*mcp.CallToolResult, SetupPostgresFreeOutput, error) {
	if input.Name == "" {
		input.Name = "app-db"
	}

	// Call the real implementation directly
	err := tools.SetupPostgresWithSchema(ctx, map[string]string{
		"name": input.Name,
	})

	if err != nil {
		return nil, SetupPostgresFreeOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Connection info will be printed by the tool
	connInfo := ""

	return nil, SetupPostgresFreeOutput{
		Success:        true,
		Message:        fmt.Sprintf("Created PostgreSQL database '%s' on Tiger Cloud (free tier) with auto-schema", input.Name),
		ConnectionInfo: connInfo,
	}, nil
}

type OpenAppInput struct {
	URL string `json:"url,omitempty" jsonschema:"URL to open (default: http://localhost:3000)"`
}

type OpenAppOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func (s *Server) handleOpenApp(ctx context.Context, req *mcp.CallToolRequest, input OpenAppInput) (*mcp.CallToolResult, OpenAppOutput, error) {
	url := input.URL
	if url == "" {
		url = "http://localhost:3000"
	}

	err := tools.OpenBrowser(url)
	if err != nil {
		return nil, OpenAppOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to open browser: %v", err),
			URL:     url,
		}, nil
	}

	return nil, OpenAppOutput{
		Success: true,
		Message: fmt.Sprintf("Opened %s in browser", url),
		URL:     url,
	}, nil
}

// Additional handlers would go here
