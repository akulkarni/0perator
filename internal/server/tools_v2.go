package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/operator"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerOperatorTool registers the minimal operator tool for backward compatibility
func (s *Server) registerOperatorTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "operator",
		Description: "Advanced tool for complex multi-step operations or discovering available actions. Use direct tools like create_web_app and setup_database instead when possible. Commands: 'list' (show all actions), 'discover' (search for actions), 'execute' (run action sequences).",
	}, s.handleOperator)
}

// OperatorInput represents the input for the operator tool
type OperatorInput struct {
	Command string                 `json:"command" jsonschema:"Command to execute: list, execute_tool, execute_recipe"`
	Tool    string                 `json:"tool,omitempty" jsonschema:"Tool name to execute"`
	Recipe  string                 `json:"recipe,omitempty" jsonschema:"Recipe name to execute"`
	Inputs  map[string]string      `json:"inputs,omitempty" jsonschema:"Inputs for the tool or recipe"`
}

// OperatorOutput represents the output from the operator tool
type OperatorOutput struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Tools   []string `json:"tools,omitempty"`
	Recipes []string `json:"recipes,omitempty"`
	Error   string   `json:"error,omitempty"`
}

func (s *Server) handleOperator(ctx context.Context, req *mcp.CallToolRequest, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	// Initialize operator if not already done
	if s.operator == nil {
		s.operator = operator.New()
	}

	switch input.Command {
	case "list":
		tools := s.operator.ListTools()
		recipes, err := s.operator.ListRecipes()
		if err != nil {
			return nil, OperatorOutput{}, fmt.Errorf("failed to list recipes: %w", err)
		}

		return &mcp.CallToolResult{
			IsError: false,
		}, OperatorOutput{
			Success: true,
			Message: fmt.Sprintf("Available: %d tools, %d recipes", len(tools), len(recipes)),
			Tools:   tools,
			Recipes: recipes,
		}, nil

	case "execute_tool":
		if input.Tool == "" {
			return nil, OperatorOutput{}, fmt.Errorf("tool name required for execute_tool command")
		}

		err := s.operator.ExecuteDirectTool(ctx, input.Tool, input.Inputs)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, OperatorOutput{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return &mcp.CallToolResult{
			IsError: false,
		}, OperatorOutput{
			Success: true,
			Message: fmt.Sprintf("Successfully executed tool: %s", input.Tool),
		}, nil

	case "execute_recipe":
		if input.Recipe == "" {
			return nil, OperatorOutput{}, fmt.Errorf("recipe name required for execute_recipe command")
		}

		err := s.operator.ExecuteRecipe(ctx, input.Recipe, input.Inputs)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, OperatorOutput{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return &mcp.CallToolResult{
			IsError: false,
		}, OperatorOutput{
			Success: true,
			Message: fmt.Sprintf("Successfully executed recipe: %s", input.Recipe),
		}, nil

	default:
		// For backward compatibility, try to execute as a tool
		err := s.operator.ExecuteDirectTool(ctx, input.Command, input.Inputs)
		if err != nil {
			return nil, OperatorOutput{}, fmt.Errorf("unknown command: %s", input.Command)
		}

		return &mcp.CallToolResult{
			IsError: false,
		}, OperatorOutput{
			Success: true,
			Message: fmt.Sprintf("Successfully executed: %s", input.Command),
		}, nil
	}
}