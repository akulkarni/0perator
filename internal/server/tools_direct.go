package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/tools"
	"github.com/akulkarni/0perator/skills"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerDirectTools registers individual MCP tools for direct access
func (s *Server) registerDirectTools() error {
	// Universal web app tool (handles all frameworks)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_web_app",
		Description: "🚀 Create any web application - Build an opinionated next.js app",
	}, s.handleCreateWebApp)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_database",
		Description: "🗄️ Set up any database - PostgreSQL on Tiger Cloud (default, FREE). Auto-configures with schema, migrations, and connection handling. Use for any database request.",
	}, s.handleCreateDatabase)

	// Open app in browser - call this after all setup is complete
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "open_app",
		Description: "🌐 Open the app in browser. Call this AFTER all setup (database, auth, UI) is complete to show the user their running app.",
	}, s.handleOpenApp)

	// View skill instructions
	skillsList, err := skills.ListSkills()
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "view_skill",
		Description: "📖 View instructions for a specific skill by name.\n\nAvailable skills:\n" + skillsList,
	}, s.handleViewSkill)

	return nil
}

// Input/Output types for each tool

type CreateWebAppInput struct {
	Name        string `json:"name" jsonschema:"Application name"`
	DBServiceID string `json:"db_service_id,omitempty" jsonschema:"Database service ID to connect to"`
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

	err := tools.CreateNextJSApp(ctx, input.Name, input.DBServiceID)

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

type CreateDatabaseInput struct {
	DatabaseName string `json:"name,omitempty" jsonschema:"Database name (default: app-db)"`
}

type CreateDatabaseOutput struct {
	Success   bool   `json:"success"`
	ServiceID string `json:"service_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *Server) handleCreateDatabase(ctx context.Context, req *mcp.CallToolRequest, input CreateDatabaseInput) (*mcp.CallToolResult, CreateDatabaseOutput, error) {
	// Set defaults
	if input.DatabaseName == "" {
		input.DatabaseName = "app-db"
	}

	serviceID, err := tools.CreateDatabase(ctx, input.DatabaseName)

	if err != nil {
		return nil, CreateDatabaseOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return nil, CreateDatabaseOutput{
		Success:   true,
		ServiceID: serviceID,
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

type ViewSkillInput struct {
	Name string `json:"name" jsonschema:"Skill name (directory name)"`
}

type ViewSkillOutput struct {
	Success     bool   `json:"success"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Body        string `json:"body,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (s *Server) handleViewSkill(ctx context.Context, req *mcp.CallToolRequest, input ViewSkillInput) (*mcp.CallToolResult, ViewSkillOutput, error) {
	if input.Name == "" {
		return nil, ViewSkillOutput{
			Success: false,
			Error:   "skill name is required",
		}, nil
	}

	skill, err := skills.GetSkill(input.Name)
	if err != nil {
		return nil, ViewSkillOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return nil, ViewSkillOutput{
		Success:     true,
		Name:        skill.Name,
		Description: skill.Description,
		Body:        skill.Body,
	}, nil
}
