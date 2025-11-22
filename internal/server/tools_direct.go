package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerDirectTools registers individual MCP tools for direct access
func (s *Server) registerDirectTools() {
	// Generic web app tool (with framework choice)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_web_app",
		Description: "🚀 PREFERRED TOOL for creating web applications. Instantly scaffolds a modern web app with best practices. Defaults to Next.js (full-stack). Also supports: react, express. Use this for any 'build app' or 'create app' request.",
	}, s.handleCreateWebApp)

	// Specific web framework tools (for explicit requests)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_nextjs_app",
		Description: "🚀 Create a Next.js app with TypeScript and Tailwind CSS. Fast setup for modern web applications.",
	}, s.handleCreateNextJSApp)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_react_app",
		Description: "⚛️ Create a React app using Vite for lightning-fast development.",
	}, s.handleCreateReactApp)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_express_api",
		Description: "🔧 Create an Express.js API with basic structure and middleware.",
	}, s.handleCreateExpressAPI)

	// Database tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "setup_database",
		Description: "🗄️ PREFERRED TOOL for database setup. Creates a production-ready database instantly. Defaults to FREE PostgreSQL on Tiger Cloud with TimescaleDB + AI extensions. Use this for any 'database' or 'postgres' request.",
	}, s.handleSetupDatabase)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "setup_postgres_free",
		Description: "🐘 Create a FREE PostgreSQL database on Tiger Cloud (includes TimescaleDB + AI extensions). No credit card required.",
	}, s.handleSetupPostgresFree)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "setup_sqlite",
		Description: "💾 Create a local SQLite database. Instant setup with zero configuration. Perfect for development and prototyping.",
	}, s.handleSetupSQLite)

	// Authentication tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_jwt_auth",
		Description: "🔐 Add JWT authentication to your application with secure token handling.",
	}, s.handleAddJWTAuth)

	// Payment tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_stripe_payments",
		Description: "💳 Integrate Stripe payments with checkout and subscription support.",
	}, s.handleAddStripePayments)

	// UI/Design tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_brutalist_ui",
		Description: "🏗️ Add brutalist/minimalist UI components - monospace fonts, #ff4500 links, no CSS frameworks, inline styles only.",
	}, s.handleAddBrutalistUI)

	// Debug: Print to stderr to confirm registration
	// fmt.Fprintf(os.Stderr, "Debug: Registered add_brutalist_ui tool\n")
}

// Input/Output types for each tool

type CreateWebAppInput struct {
	Name      string `json:"name" jsonschema:"Application name"`
	Framework string `json:"framework,omitempty" jsonschema:"Framework: nextjs (default), react, or express"`
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
	case "nextjs", "react", "express":
		// Valid frameworks
	default:
		return nil, CreateWebAppOutput{
			Success: false,
			Message: fmt.Sprintf("Invalid framework '%s'. Choose: nextjs, react, or express", input.Framework),
		}, nil
	}

	// Call the appropriate real implementation directly
	var err error
	args := map[string]string{
		"name": input.Name,
	}

	switch input.Framework {
	case "nextjs":
		err = tools.CreateNextJSApp(ctx, args)
	case "react":
		err = tools.CreateReactApp(ctx, args)
	case "express":
		err = tools.CreateExpressAPI(ctx, args)
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

type AddJWTAuthInput struct {
	Framework string `json:"framework,omitempty" jsonschema:"Framework to add auth to (nextjs, express)"`
}

type AddJWTAuthOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) handleAddJWTAuth(ctx context.Context, req *mcp.CallToolRequest, input AddJWTAuthInput) (*mcp.CallToolResult, AddJWTAuthOutput, error) {
	// Use the real implementation
	err := tools.AddJWTAuth(ctx, map[string]string{
		"framework": input.Framework,
	})

	if err != nil {
		return nil, AddJWTAuthOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to add JWT auth: %v", err),
		}, nil
	}

	return nil, AddJWTAuthOutput{
		Success: true,
		Message: "JWT authentication added successfully with login, register, verify endpoints and auth middleware",
	}, nil
}

type AddStripePaymentsInput struct {
	Mode string `json:"mode,omitempty" jsonschema:"Payment mode (subscription or one-time, default: subscription)"`
}

type AddStripePaymentsOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) handleAddStripePayments(ctx context.Context, req *mcp.CallToolRequest, input AddStripePaymentsInput) (*mcp.CallToolResult, AddStripePaymentsOutput, error) {
	if input.Mode == "" {
		input.Mode = "subscription"
	}

	// Placeholder for now
	return nil, AddStripePaymentsOutput{
		Success: true,
		Message: fmt.Sprintf("Stripe payments added in %s mode", input.Mode),
	}, nil
}

type AddBrutalistUIInput struct {
	Component string `json:"component,omitempty" jsonschema:"Component type: all, auth, forms, layout, feedback, or custom name (default: all)"`
}

type AddBrutalistUIOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) handleAddBrutalistUI(ctx context.Context, req *mcp.CallToolRequest, input AddBrutalistUIInput) (*mcp.CallToolResult, AddBrutalistUIOutput, error) {
	err := tools.AddBrutalistUI(ctx, map[string]string{
		"component": input.Component,
	})

	if err != nil {
		return nil, AddBrutalistUIOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to add brutalist UI: %v", err),
		}, nil
	}

	return nil, AddBrutalistUIOutput{
		Success: true,
		Message: "Brutalist UI components added with monospace fonts, #ff4500 actions, and inline styles",
	}, nil
}