package operator

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/recipes"
	"github.com/akulkarni/0perator/internal/tools"
)

// Operator is a minimal wrapper for backward compatibility
// All new code should use direct tools instead
type Operator struct{}

// New creates a new Operator instance
func New() *Operator {
	return &Operator{}
}

// ExecuteDirectTool executes a tool directly by name
// This is mainly for backward compatibility with the operator pattern
func (o *Operator) ExecuteDirectTool(ctx context.Context, toolName string, args map[string]string) error {
	switch toolName {
	case "create_web_app":
		// Handle framework selection
		framework := args["framework"]
		if framework == "" {
			framework = "nextjs"
		}
		switch framework {
		case "nextjs":
			return tools.CreateNextJSApp(ctx, args)
		case "react":
			return tools.CreateReactApp(ctx, args)
		case "express":
			return tools.CreateExpressAPI(ctx, args)
		default:
			return fmt.Errorf("unsupported framework: %s", framework)
		}
	case "setup_database":
		// Handle database type selection
		dbType := args["type"]
		if dbType == "" {
			dbType = "postgres"
		}
		switch dbType {
		case "postgres":
			return tools.SetupPostgresFree(ctx, args)
		case "sqlite":
			return tools.SetupSQLite(ctx, args)
		default:
			return fmt.Errorf("unsupported database type: %s", dbType)
		}
	case "add_ui_theme":
		// Handle theme selection
		theme := args["theme"]
		if theme == "" {
			theme = "brutalist"
		}
		switch theme {
		case "brutalist":
			return tools.AddBrutalistUI(ctx, args)
		default:
			return fmt.Errorf("theme %s not yet implemented", theme)
		}
	case "add_auth":
		// Handle auth type selection
		authType := args["type"]
		if authType == "" {
			authType = "jwt"
		}
		switch authType {
		case "jwt":
			return tools.AddJWTAuth(ctx, args)
		default:
			return fmt.Errorf("auth type %s not yet implemented", authType)
		}
	// Backward compatibility
	case "add_brutalist_ui":
		return tools.AddBrutalistUI(ctx, args)
	case "add_jwt_auth":
		return tools.AddJWTAuth(ctx, args)
	// Keep old names for backward compatibility
	case "create_nextjs_app":
		return tools.CreateNextJSApp(ctx, args)
	case "create_react_app":
		return tools.CreateReactApp(ctx, args)
	case "create_express_api":
		return tools.CreateExpressAPI(ctx, args)
	case "setup_postgres_free":
		return tools.SetupPostgresFree(ctx, args)
	case "setup_sqlite":
		return tools.SetupSQLite(ctx, args)
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ExecuteRecipe executes a recipe by name
func (o *Operator) ExecuteRecipe(ctx context.Context, recipeName string, inputs map[string]string) error {
	recipe, err := recipes.Load(recipeName)
	if err != nil {
		return fmt.Errorf("failed to load recipe: %w", err)
	}

	executor := recipes.NewExecutor()
	return executor.Execute(ctx, recipe, inputs)
}

// ListTools returns available direct tools
func (o *Operator) ListTools() []string {
	return []string{
		"create_web_app",  // Handles all frameworks
		"setup_database",  // Handles postgres and sqlite
		"add_auth",        // Handles jwt, oauth, magic-link, session, passkey
		"add_ui_theme",    // Handles brutalist, shadcn, material, etc.
	}
}

// ListRecipes returns available recipes
func (o *Operator) ListRecipes() ([]string, error) {
	return recipes.List()
}