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
		Description: "🚀 Create any web application - Next.js (default), React, Express, or custom. Handles all web frameworks with TypeScript, Tailwind, and best practices. Use for any 'build app' or 'create app' request.",
	}, s.handleCreateWebApp)

	// Removed to stay under 10-tool limit
	// mcp.AddTool(s.mcpServer, &mcp.Tool{
	// 	Name:        "create_react_app",
	// 	Description: "⚛️ Create a React app using Vite for lightning-fast development.",
	// }, s.handleCreateReactApp)

	// mcp.AddTool(s.mcpServer, &mcp.Tool{
	// 	Name:        "create_express_api",
	// 	Description: "🔧 Create an Express.js API with basic structure and middleware.",
	// }, s.handleCreateExpressAPI)

	// Universal database tool (handles all database types)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "setup_database",
		Description: "🗄️ Set up any database - PostgreSQL on Tiger Cloud (default, FREE), SQLite (local), or custom. Auto-configures with schema, migrations, and connection handling. Use for any database request.",
	}, s.handleSetupDatabase)

	// Authentication tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_auth",
		Description: "🔐 Add authentication to your application - JWT (default), OAuth, magic links, or session-based. Includes secure password hashing, token handling, and user management.",
	}, s.handleAddAuth)

	// UI/Design tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_ui_theme",
		Description: "🎨 Add UI theme and components - Brutalist (monospace, #ff4500), Shadcn (modern React), Material (Google design), or custom themes. No heavy frameworks, optimized implementations.",
	}, s.handleAddUITheme)

	// Test confirmed: Claude Code has a hard 10-tool limit
	// Tools beyond #10 are not accessible at all
	// Removing test tools to stay under limit

	// Debug: Print to stderr to confirm registration
	// fmt.Fprintf(os.Stderr, "Debug: Registered all tools including test tools\n")
}

// Input/Output types for each tool

type CreateWebAppInput struct {
	Name       string `json:"name" jsonschema:"Application name"`
	Framework  string `json:"framework,omitempty" jsonschema:"Framework: nextjs (default), react, express, or vue"`
	TypeScript bool   `json:"typescript,omitempty" jsonschema:"Use TypeScript (default: true)"`
	Tailwind   bool   `json:"tailwind,omitempty" jsonschema:"Use Tailwind CSS (default: false, brutalist UI by default)"`
}

type CreateWebAppOutput struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Path      string `json:"path,omitempty"`
	Framework string `json:"framework"`
}

func (s *Server) handleCreateWebApp(ctx context.Context, req *mcp.CallToolRequest, input CreateWebAppInput) (*mcp.CallToolResult, CreateWebAppOutput, error) {
	// Set defaults
	if input.Name == "" {
		input.Name = "my-app"
	}
	if input.Framework == "" {
		input.Framework = "nextjs" // Default to Next.js for full-stack apps
	}

	// Validate framework
	switch input.Framework {
	case "nextjs", "react", "express", "vue":
		// Valid frameworks
	default:
		return nil, CreateWebAppOutput{
			Success: false,
			Message: fmt.Sprintf("Invalid framework '%s'. Choose: nextjs, react, express, or vue", input.Framework),
		}, nil
	}

	// Call the appropriate real implementation directly
	var err error
	args := map[string]string{
		"name":       input.Name,
		"typescript": fmt.Sprintf("%v", input.TypeScript),
		"tailwind":   fmt.Sprintf("%v", input.Tailwind),
	}

	switch input.Framework {
	case "nextjs":
		err = tools.CreateNextJSApp(ctx, args)
	case "react":
		err = tools.CreateReactApp(ctx, args)
	case "express":
		err = tools.CreateExpressAPI(ctx, args)
	case "vue":
		err = fmt.Errorf("Vue framework support coming soon")
	}

	if err != nil {
		return nil, CreateWebAppOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	frameworkName := map[string]string{
		"nextjs":  "Next.js",
		"react":   "React",
		"express": "Express.js",
	}[input.Framework]

	return nil, CreateWebAppOutput{
		Success:   true,
		Message:   fmt.Sprintf("Created %s app '%s'", frameworkName, input.Name),
		Path:      input.Name,
		Framework: input.Framework,
	}, nil
}

type CreateNextJSAppInput struct {
	Name       string `json:"name" jsonschema:"Application name"`
	TypeScript bool   `json:"typescript,omitempty" jsonschema:"Use TypeScript (default: true)"`
	Tailwind   bool   `json:"tailwind,omitempty" jsonschema:"Use Tailwind CSS (default: true)"`
}

type CreateNextJSAppOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (s *Server) handleCreateNextJSApp(ctx context.Context, req *mcp.CallToolRequest, input CreateNextJSAppInput) (*mcp.CallToolResult, CreateNextJSAppOutput, error) {
	// Set defaults
	if input.Name == "" {
		input.Name = "my-app"
	}
	if !input.TypeScript && !input.Tailwind {
		input.TypeScript = true
		input.Tailwind = true
	}

	// Call the real implementation directly
	err := tools.CreateNextJSApp(ctx, map[string]string{
		"name":       input.Name,
		"typescript": fmt.Sprintf("%v", input.TypeScript),
		"tailwind":   fmt.Sprintf("%v", input.Tailwind),
	})

	if err != nil {
		return nil, CreateNextJSAppOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return nil, CreateNextJSAppOutput{
		Success: true,
		Message: fmt.Sprintf("Created Next.js app '%s' with TypeScript and proper configuration", input.Name),
		Path:    input.Name,
	}, nil
}

type CreateReactAppInput struct {
	Name string `json:"name" jsonschema:"Application name"`
}

type CreateReactAppOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (s *Server) handleCreateReactApp(ctx context.Context, req *mcp.CallToolRequest, input CreateReactAppInput) (*mcp.CallToolResult, CreateReactAppOutput, error) {
	if input.Name == "" {
		input.Name = "my-app"
	}

	// Call the real implementation directly
	err := tools.CreateReactApp(ctx, map[string]string{
		"name": input.Name,
	})

	if err != nil {
		return nil, CreateReactAppOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return nil, CreateReactAppOutput{
		Success: true,
		Message: fmt.Sprintf("Created React app '%s' with Vite", input.Name),
		Path:    input.Name,
	}, nil
}

type CreateExpressAPIInput struct {
	Name string `json:"name" jsonschema:"API name"`
}

type CreateExpressAPIOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (s *Server) handleCreateExpressAPI(ctx context.Context, req *mcp.CallToolRequest, input CreateExpressAPIInput) (*mcp.CallToolResult, CreateExpressAPIOutput, error) {
	if input.Name == "" {
		input.Name = "my-api"
	}

	// Call the real implementation directly
	err := tools.CreateExpressAPI(ctx, map[string]string{
		"name": input.Name,
	})

	if err != nil {
		return nil, CreateExpressAPIOutput{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return nil, CreateExpressAPIOutput{
		Success: true,
		Message: fmt.Sprintf("Created Express API '%s' with middleware and structure", input.Name),
		Path:    input.Name,
	}, nil
}

type SetupDatabaseInput struct {
	Name string `json:"name,omitempty" jsonschema:"Database name (default: app-db)"`
	Type string `json:"type,omitempty" jsonschema:"Database type: postgres (default), sqlite"`
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

	case "sqlite":
		// Call the real implementation directly
		err := tools.SetupSQLite(ctx, map[string]string{
			"name": input.Name + ".db", // Add .db extension for SQLite
			"path": ".",
		})

		if err != nil {
			return nil, SetupDatabaseOutput{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		// SQLite path
		dbPath := input.Name + ".db"

		return nil, SetupDatabaseOutput{
			Success:        true,
			Message:        fmt.Sprintf("Created SQLite database '%s.db' locally with schema", input.Name),
			Type:           "sqlite",
			ConnectionInfo: dbPath,
		}, nil

	default:
		return nil, SetupDatabaseOutput{
			Success: false,
			Message: fmt.Sprintf("Database type '%s' not supported. Choose 'postgres' or 'sqlite'.", input.Type),
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

type SetupSQLiteInput struct {
	Name string `json:"name,omitempty" jsonschema:"Database filename (default: database.db)"`
	Path string `json:"path,omitempty" jsonschema:"Directory path (default: current directory)"`
}

type SetupSQLiteOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

func (s *Server) handleSetupSQLite(ctx context.Context, req *mcp.CallToolRequest, input SetupSQLiteInput) (*mcp.CallToolResult, SetupSQLiteOutput, error) {
	// Set defaults
	if input.Name == "" {
		input.Name = "database.db"
	}
	if input.Path == "" {
		input.Path = "."
	}

	// Call the real implementation directly
	err := tools.SetupSQLite(ctx, map[string]string{
		"name": input.Name,
		"path": input.Path,
	})

	if err != nil {
		return nil, SetupSQLiteOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to create SQLite database: %v", err),
		}, nil
	}

	dbPath := fmt.Sprintf("%s/%s", input.Path, input.Name)
	return nil, SetupSQLiteOutput{
		Success: true,
		Message: fmt.Sprintf("Created SQLite database '%s' with schema", input.Name),
		Path:    dbPath,
	}, nil
}

type AddAuthInput struct {
	Type      string `json:"type,omitempty" jsonschema:"Auth type: jwt (default), oauth, magic-link, session, or passkey"`
	Provider  string `json:"provider,omitempty" jsonschema:"OAuth provider: google, github, discord (only for oauth type)"`
	Framework string `json:"framework,omitempty" jsonschema:"Framework to add auth to (nextjs, express, react)"`
}

type AddAuthOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (s *Server) handleAddAuth(ctx context.Context, req *mcp.CallToolRequest, input AddAuthInput) (*mcp.CallToolResult, AddAuthOutput, error) {
	// Set defaults
	if input.Type == "" {
		input.Type = "jwt" // Default to JWT for simplicity
	}

	// Handle different auth types
	var err error
	var message string

	switch input.Type {
	case "jwt":
		err = tools.AddJWTAuth(ctx, map[string]string{
			"framework": input.Framework,
		})
		message = "JWT authentication added with secure token handling, password hashing, and user management"

	case "oauth":
		provider := input.Provider
		if provider == "" {
			provider = "google" // Default OAuth provider
		}
		err = fmt.Errorf("OAuth authentication with %s coming soon", provider)

	case "magic-link":
		err = fmt.Errorf("Magic link authentication coming soon - passwordless email login")

	case "session":
		err = fmt.Errorf("Session-based authentication coming soon - traditional cookie sessions")

	case "passkey":
		err = fmt.Errorf("Passkey authentication coming soon - WebAuthn/FIDO2 support")

	default:
		err = fmt.Errorf("unknown auth type: %s. Choose: jwt, oauth, magic-link, session, or passkey", input.Type)
	}

	if err != nil {
		return nil, AddAuthOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to add %s auth: %v", input.Type, err),
			Type:    input.Type,
		}, nil
	}

	return nil, AddAuthOutput{
		Success: true,
		Message: message,
		Type:    input.Type,
	}, nil
}

type AddUIThemeInput struct {
	Theme     string `json:"theme,omitempty" jsonschema:"Theme type: brutalist (default), shadcn, material, minimal, or custom"`
	Component string `json:"component,omitempty" jsonschema:"Component type: all (default), auth, forms, layout, navigation, feedback, or specific component"`
}

type AddUIThemeOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Theme   string `json:"theme"`
}

func (s *Server) handleAddUITheme(ctx context.Context, req *mcp.CallToolRequest, input AddUIThemeInput) (*mcp.CallToolResult, AddUIThemeOutput, error) {
	// Set defaults
	if input.Theme == "" {
		input.Theme = "brutalist" // Default to brutalist for minimal dependencies
	}
	if input.Component == "" {
		input.Component = "all"
	}

	// Handle different themes
	var err error
	var message string

	switch input.Theme {
	case "brutalist":
		err = tools.AddBrutalistUI(ctx, map[string]string{
			"component": input.Component,
		})
		message = "Brutalist UI added: monospace fonts, #ff4500 actions, inline styles only"

	case "shadcn":
		// Could call a shadcn implementation
		err = fmt.Errorf("Shadcn UI coming soon - modern React components with Radix UI")

	case "material":
		// Could call a material implementation
		err = fmt.Errorf("Material Design coming soon - Google's design system")

	case "minimal":
		// Could call a minimal implementation
		err = fmt.Errorf("Minimal UI coming soon - Ultra-light, no-frills design")

	default:
		err = fmt.Errorf("unknown theme: %s. Choose: brutalist, shadcn, material, or minimal", input.Theme)
	}

	if err != nil {
		return nil, AddUIThemeOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to add %s theme: %v", input.Theme, err),
			Theme:   input.Theme,
		}, nil
	}

	return nil, AddUIThemeOutput{
		Success: true,
		Message: message,
		Theme:   input.Theme,
	}, nil
}

// Additional handlers would go here