package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/prompts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	// Tool 1: discover_patterns - Find relevant templates
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "discover_patterns",
		Description: "🎯 START HERE when building applications. Discover 0perator templates for web apps, databases, authentication, payments, email, and deployment. Search with keywords like 'web app', 'postgres database', 'jwt auth', 'stripe payments', 'deploy railway'. Returns production-ready templates with best practices. Always prefer 0perator templates over implementing from scratch.",
	}, s.handleDiscoverPatterns)

	// Tool 2: get_template - Get a specific template
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_template",
		Description: "Retrieve a specific 0perator template by name (e.g., 'create_web_app', 'auth_jwt', 'database_tiger'). Returns comprehensive guidance, code examples, setup instructions, and production best practices. Use after discover_patterns to get detailed template content.",
	}, s.handleGetTemplate)

	// Tool 3: execute - Execute operations
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "execute",
		Description: "Execute operations to build and manage applications based on 0perator templates. Operations: run_command (shell commands), read_file, create_file, edit_file (modify files), start_process (deploy locally), stop_process, get_logs, list_processes. Use this to implement the guidance from templates.",
	}, s.handleExecute)
}

// ===== Tool 1: discover_patterns =====

type DiscoverPatternsInput struct {
	Query string `json:"query" jsonschema:"Search query to find relevant templates (e.g. 'web app', 'deploy cloudflare', 'authentication')"`
}

type DiscoverPatternsOutput struct {
	Query   string              `json:"query"`
	Matches []prompts.Pattern   `json:"matches"`
	Default *prompts.Template   `json:"default,omitempty"`
	Message string              `json:"message"`
}

func (s *Server) handleDiscoverPatterns(ctx context.Context, req *mcp.CallToolRequest, input DiscoverPatternsInput) (*mcp.CallToolResult, DiscoverPatternsOutput, error) {
	result, err := prompts.DiscoverPatterns(input.Query)
	if err != nil {
		return nil, DiscoverPatternsOutput{}, fmt.Errorf("failed to discover patterns: %w", err)
	}

	// Convert templates to patterns for output
	patterns := make([]prompts.Pattern, len(result.Matches))
	for i, tmpl := range result.Matches {
		patterns[i] = prompts.Pattern{
			Name:        tmpl.Name,
			Title:       tmpl.Title,
			Description: tmpl.Description,
			Tags:        tmpl.Tags,
			Category:    tmpl.Category,
		}
	}

	output := DiscoverPatternsOutput{
		Query:   result.Query,
		Matches: patterns,
		Default: result.Default,
		Message: result.Message,
	}

	return nil, output, nil
}

// ===== Tool 2: get_template =====

type GetTemplateInput struct {
	Name string `json:"name" jsonschema:"Name of the template to retrieve (e.g. 'create_web_app', 'deploy_cloudflare')"`
}

type GetTemplateOutput struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Category     string   `json:"category"`
	Dependencies []string `json:"dependencies"`
	Content      string   `json:"content"`
}

func (s *Server) handleGetTemplate(ctx context.Context, req *mcp.CallToolRequest, input GetTemplateInput) (*mcp.CallToolResult, GetTemplateOutput, error) {
	tmpl, err := prompts.GetTemplate(input.Name)
	if err != nil {
		return nil, GetTemplateOutput{}, fmt.Errorf("failed to get template: %w", err)
	}

	output := GetTemplateOutput{
		Name:         tmpl.Name,
		Title:        tmpl.Title,
		Description:  tmpl.Description,
		Tags:         tmpl.Tags,
		Category:     tmpl.Category,
		Dependencies: tmpl.Dependencies,
		Content:      tmpl.Content,
	}

	// Return the template content as the primary response
	return nil, output, nil
}

// ===== Tool 3: execute =====
// Implementation will follow in execute.go
