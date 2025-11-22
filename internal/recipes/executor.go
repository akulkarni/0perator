package recipes

import (
	"context"
	"fmt"
	"strings"
)

// ToolFunc is a function that implements a tool
type ToolFunc func(ctx context.Context, args map[string]string) error

// Executor executes recipes by calling registered tools
type Executor struct {
	tools   map[string]ToolFunc
	verbose bool
}

// NewExecutor creates a new recipe executor
func NewExecutor() *Executor {
	return &Executor{
		tools:   make(map[string]ToolFunc),
		verbose: true,
	}
}

// RegisterTool registers a tool implementation
func (e *Executor) RegisterTool(name string, fn ToolFunc) {
	e.tools[name] = fn
}

// Execute executes a recipe with user inputs (simplified interface)
func (e *Executor) Execute(ctx context.Context, recipe *Recipe, userInputs map[string]string) error {
	// Convert string inputs to interface{} for compatibility
	inputs := make(map[string]interface{})
	for k, v := range userInputs {
		inputs[k] = v
	}

	results, err := e.ExecuteRecipe(ctx, recipe, inputs)

	// Print results
	for _, result := range results {
		fmt.Println(result)
	}

	return err
}

// ExecuteRecipe executes a recipe with user inputs
func (e *Executor) ExecuteRecipe(ctx context.Context, recipe *Recipe, userInputs map[string]interface{}) ([]string, error) {
	results := []string{}

	// Validate and prepare inputs
	inputs, err := recipe.ValidateUserInputs(userInputs)
	if err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Add header to results
	results = append(results, fmt.Sprintf("🚀 Executing recipe: %s", recipe.Name))
	results = append(results, fmt.Sprintf("📝 %s", recipe.Desc))
	results = append(results, "")

	// Execute each step
	for i, step := range recipe.Steps {
		// Replace variables in step
		step = ReplaceVariables(step, inputs)

		// Parse step into tool and args
		toolName, args := ParseStep(step)

		if e.verbose {
			results = append(results, fmt.Sprintf("Step %d: %s", i+1, step))
		}

		// Find and execute tool
		tool, exists := e.tools[toolName]
		if !exists {
			// Try to find tool with dynamic name (e.g., add_{{auth}}_auth)
			dynamicName := ReplaceVariables(toolName, inputs)
			tool, exists = e.tools[dynamicName]
			if !exists {
				return results, fmt.Errorf("unknown tool: %s", toolName)
			}
			toolName = dynamicName
		}

		// Execute the tool
		if err := tool(ctx, args); err != nil {
			results = append(results, fmt.Sprintf("  ❌ Failed: %v", err))
			return results, fmt.Errorf("step %d failed (%s): %w", i+1, toolName, err)
		}

		results = append(results, fmt.Sprintf("  ✅ %s completed", toolName))
	}

	results = append(results, "")
	results = append(results, fmt.Sprintf("✨ Recipe '%s' completed successfully!", recipe.Name))

	return results, nil
}

// ExecuteStep executes a single step from a recipe
func (e *Executor) ExecuteStep(ctx context.Context, step string, inputs map[string]string) error {
	// Replace variables
	step = ReplaceVariables(step, inputs)

	// Parse step
	toolName, args := ParseStep(step)

	// Find tool
	tool, exists := e.tools[toolName]
	if !exists {
		// Try dynamic name
		dynamicName := ReplaceVariables(toolName, inputs)
		tool, exists = e.tools[dynamicName]
		if !exists {
			return fmt.Errorf("unknown tool: %s", toolName)
		}
	}

	// Execute
	return tool(ctx, args)
}

// ListTools returns all registered tool names
func (e *Executor) ListTools() []string {
	tools := []string{}
	for name := range e.tools {
		tools = append(tools, name)
	}
	return tools
}

// HasTool checks if a tool is registered
func (e *Executor) HasTool(name string) bool {
	_, exists := e.tools[name]
	return exists
}

// SetVerbose sets whether to output detailed execution logs
func (e *Executor) SetVerbose(verbose bool) {
	e.verbose = verbose
}

// DryRun simulates recipe execution without actually running tools
func (e *Executor) DryRun(recipe *Recipe, userInputs map[string]interface{}) ([]string, error) {
	results := []string{}

	// Validate inputs
	inputs, err := recipe.ValidateUserInputs(userInputs)
	if err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	results = append(results, fmt.Sprintf("DRY RUN: %s", recipe.Name))
	results = append(results, fmt.Sprintf("Description: %s", recipe.Desc))
	results = append(results, "")
	results = append(results, "Would execute:")

	for i, step := range recipe.Steps {
		// Replace variables
		step = ReplaceVariables(step, inputs)
		toolName, args := ParseStep(step)

		// Check if tool exists
		if !e.HasTool(toolName) {
			dynamicName := ReplaceVariables(toolName, inputs)
			if !e.HasTool(dynamicName) {
				results = append(results, fmt.Sprintf("  %d. ⚠️  %s (TOOL NOT FOUND)", i+1, step))
				continue
			}
			toolName = dynamicName
		}

		// Format args for display
		argsList := []string{}
		for k, v := range args {
			argsList = append(argsList, fmt.Sprintf("%s=%s", k, v))
		}

		results = append(results, fmt.Sprintf("  %d. %s", i+1, toolName))
		if len(argsList) > 0 {
			results = append(results, fmt.Sprintf("     Args: %s", strings.Join(argsList, ", ")))
		}
	}

	return results, nil
}