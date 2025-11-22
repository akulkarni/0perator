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
	case "add_brutalist_ui":
		return tools.AddBrutalistUI(ctx, args)
	case "add_jwt_auth":
		return tools.AddJWTAuth(ctx, args)
	case "add_stripe_payments":
		// Stub - returns basic implementation
		return fmt.Errorf("Stripe payments integration is not yet implemented")
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
		"create_nextjs_app",
		// "create_react_app",     // Stub - commented out
		// "create_express_api",   // Stub - commented out
		"setup_postgres_free",
		"setup_sqlite",
		"add_brutalist_ui",
		"add_jwt_auth",
		// "add_stripe_payments",  // Stub - commented out
	}
}

// ListRecipes returns available recipes
func (o *Operator) ListRecipes() ([]string, error) {
	return recipes.List()
}