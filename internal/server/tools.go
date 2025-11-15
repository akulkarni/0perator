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
		Description: "Discover relevant prompt templates by searching tags, descriptions, and categories. Use this to find templates for building applications, adding features, deploying, etc.",
	}, s.handleDiscoverPatterns)

	// Tool 2: get_template - Get a specific template
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_template",
		Description: "Retrieve a specific prompt template by name. Returns the full template content with guidance and best practices.",
	}, s.handleGetTemplate)

	// Tool 3: execute - Execute operations
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "execute",
		Description: "Execute operations for building and managing applications. Supports: run_command, read_file, create_file, edit_file, start_process, stop_process, get_logs, list_processes.",
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
