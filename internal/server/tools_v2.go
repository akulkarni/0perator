package server

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/actions"
	"github.com/akulkarni/0perator/internal/operator"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OperatorCommand represents the different operator commands
type OperatorCommand string

const (
	OpCommandDiscover OperatorCommand = "discover"
	OpCommandExecute  OperatorCommand = "execute"
	OpCommandValidate OperatorCommand = "validate"
	OpCommandList     OperatorCommand = "list"
)

// OperatorType represents what to operate on
type OperatorType string

const (
	OpTypeAction   OperatorType = "action"
	OpTypeSequence OperatorType = "sequence"
	OpTypeRecipe   OperatorType = "recipe" // For future use
)

// registerOperatorTool registers the new unified operator tool
func (s *Server) registerOperatorTool() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "operator",
		Description: "Advanced tool for complex multi-step operations or discovering available actions. Use direct tools like create_web_app and setup_database instead when possible. Commands: 'list' (show all actions), 'discover' (search for actions), 'execute' (run action sequences).",
	}, s.handleOperator)
}

// OperatorInput represents the input for the operator tool
type OperatorInput struct {
	Command  OperatorCommand        `json:"command" jsonschema:"Command to execute: discover, execute, validate, or list"`
	Type     OperatorType           `json:"type,omitempty" jsonschema:"Type of operation: action or sequence"`
	Name     string                 `json:"name,omitempty" jsonschema:"Name of the action to execute"`
	Query    string                 `json:"query,omitempty" jsonschema:"Search query for discovering actions"`
	Category string                 `json:"category,omitempty" jsonschema:"Filter by category (create, setup, add, deploy)"`
	Inputs   map[string]interface{} `json:"inputs,omitempty" jsonschema:"Inputs for the action"`
	Actions  []ActionCall           `json:"actions,omitempty" jsonschema:"List of actions for sequence execution"`
}

// ActionCall represents a single action call in a sequence
type ActionCall struct {
	Action string                 `json:"action" jsonschema:"Name of the action"`
	Inputs map[string]interface{} `json:"inputs" jsonschema:"Inputs for the action"`
}

// OperatorOutput represents the output from the operator tool
type OperatorOutput struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Actions []ActionInfo           `json:"actions,omitempty"`
	Result  *actions.ActionResult  `json:"result,omitempty"`
	Results *actions.ExecutionResult `json:"results,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ActionInfo provides information about an action
type ActionInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    actions.Category `json:"category"`
	Tags        []string         `json:"tags"`
	Tier        actions.Tier     `json:"tier"`
	Inputs      []actions.Input  `json:"inputs,omitempty"`
	Outputs     []actions.Output `json:"outputs,omitempty"`
}

func (s *Server) handleOperator(ctx context.Context, req *mcp.CallToolRequest, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	// Initialize operator if not already done
	if s.operator == nil {
		s.operator = operator.New()
	}

	switch input.Command {
	case OpCommandList:
		return s.handleOperatorList(ctx, input)

	case OpCommandDiscover:
		return s.handleOperatorDiscover(ctx, input)

	case OpCommandExecute:
		return s.handleOperatorExecute(ctx, input)

	case OpCommandValidate:
		return s.handleOperatorValidate(ctx, input)

	default:
		return nil, OperatorOutput{}, fmt.Errorf("unknown command: %s", input.Command)
	}
}

func (s *Server) handleOperatorList(ctx context.Context, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	actionList := s.operator.GetAvailableActions(input.Category)

	infos := make([]ActionInfo, len(actionList))
	for i, action := range actionList {
		infos[i] = ActionInfo{
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			Tags:        action.Tags,
			Tier:        action.Tier,
		}
	}

	output := OperatorOutput{
		Success: true,
		Message: fmt.Sprintf("Found %d actions", len(infos)),
		Actions: infos,
	}

	return nil, output, nil
}

func (s *Server) handleOperatorDiscover(ctx context.Context, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	if input.Query == "" {
		return nil, OperatorOutput{}, fmt.Errorf("query is required for discover command")
	}

	actionList := s.operator.DiscoverActions(input.Query)

	infos := make([]ActionInfo, len(actionList))
	for i, action := range actionList {
		infos[i] = ActionInfo{
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			Tags:        action.Tags,
			Tier:        action.Tier,
		}
	}

	output := OperatorOutput{
		Success: true,
		Message: fmt.Sprintf("Found %d actions matching '%s'", len(infos), input.Query),
		Actions: infos,
	}

	return nil, output, nil
}

func (s *Server) handleOperatorExecute(ctx context.Context, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	// Default to "action" type if not specified and name is provided
	if input.Type == "" && input.Name != "" {
		input.Type = OpTypeAction
	}

	switch input.Type {
	case OpTypeAction:
		if input.Name == "" {
			return nil, OperatorOutput{}, fmt.Errorf("action name is required")
		}

		result, err := s.operator.ExecuteAction(ctx, input.Name, input.Inputs)
		if err != nil {
			return nil, OperatorOutput{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return nil, OperatorOutput{
			Success: true,
			Message: fmt.Sprintf("Action '%s' executed successfully", input.Name),
			Result:  result,
		}, nil

	case OpTypeSequence:
		if len(input.Actions) == 0 {
			return nil, OperatorOutput{}, fmt.Errorf("actions list is required for sequence execution")
		}

		// Convert to internal action calls
		calls := make([]actions.ActionCall, len(input.Actions))
		for i, a := range input.Actions {
			calls[i] = actions.ActionCall{
				Action: a.Action,
				Inputs: a.Inputs,
			}
		}

		results, err := s.operator.ExecuteSequence(ctx, calls)
		if err != nil {
			return nil, OperatorOutput{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return nil, OperatorOutput{
			Success: results.Success,
			Message: fmt.Sprintf("Executed %d actions in %v", len(results.Actions), results.TotalDuration),
			Results: results,
		}, nil

	default:
		return nil, OperatorOutput{}, fmt.Errorf("unknown type: %s", input.Type)
	}
}

func (s *Server) handleOperatorValidate(ctx context.Context, input OperatorInput) (*mcp.CallToolResult, OperatorOutput, error) {
	if len(input.Actions) == 0 {
		return nil, OperatorOutput{}, fmt.Errorf("actions list is required for validation")
	}

	// Convert to internal action calls
	calls := make([]actions.ActionCall, len(input.Actions))
	for i, a := range input.Actions {
		calls[i] = actions.ActionCall{
			Action: a.Action,
			Inputs: a.Inputs,
		}
	}

	err := s.operator.ValidateSequence(calls)
	if err != nil {
		return nil, OperatorOutput{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		}, nil
	}

	return nil, OperatorOutput{
		Success: true,
		Message: "Sequence is valid",
	}, nil
}