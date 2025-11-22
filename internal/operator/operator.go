package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
	"github.com/akulkarni/0perator/internal/actions/implementations"
)

// Operator is the main interface for executing actions and recipes
type Operator struct {
	registry *actions.Registry
	executor *Executor
	cache    *Cache
	mu       sync.RWMutex
}

// New creates a new Operator instance
func New() *Operator {
	registry := actions.NewRegistry()

	// Register built-in actions
	registerBuiltinActions(registry)

	return &Operator{
		registry: registry,
		executor: NewExecutor(registry),
		cache:    NewCache(),
	}
}

// registerBuiltinActions registers all built-in actions
func registerBuiltinActions(registry *actions.Registry) {
	// Register core actions - these wrap our new direct tools
	registry.Register(implementations.CreateWebAppAction())
	registry.Register(implementations.SetupPostgresAction()) // Now uses Tiger CLI directly
	registry.Register(implementations.SetupSQLiteAction())   // Local SQLite database

	// TODO: Add more actions as they are implemented
	// For now, these allow us to keep the operator interface while
	// transitioning to direct tools
	// registry.Register(implementations.AddJWTAuthAction())
	// registry.Register(implementations.AddStripePaymentsAction())
	// registry.Register(implementations.DeployLocalAction())
}

// DiscoverActions finds actions matching a query
func (o *Operator) DiscoverActions(query string) []actions.ActionMetadata {
	if query == "" {
		return o.registry.GetMetadata()
	}

	actionList := o.registry.Search(query)
	metadata := make([]actions.ActionMetadata, len(actionList))
	for i, action := range actionList {
		metadata[i] = actions.ActionMetadata{
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			Tags:        action.Tags,
			Tier:        action.Tier,
		}
	}
	return metadata
}

// GetAction returns detailed information about a specific action
func (o *Operator) GetAction(name string) (*actions.Action, error) {
	return o.registry.Get(name)
}

// ExecuteAction executes a single action
func (o *Operator) ExecuteAction(ctx context.Context, name string, inputs map[string]interface{}) (*actions.ActionResult, error) {
	call := actions.ActionCall{
		Action: name,
		Inputs: inputs,
	}

	return o.executor.ExecuteSingle(ctx, call)
}

// ExecuteSequence executes a sequence of actions
func (o *Operator) ExecuteSequence(ctx context.Context, calls []actions.ActionCall) (*actions.ExecutionResult, error) {
	// Validate sequence first
	if err := o.ValidateSequence(calls); err != nil {
		return nil, fmt.Errorf("invalid sequence: %w", err)
	}

	// Get optimal execution order
	orderedCalls, err := o.registry.GetExecutionOrder(calls)
	if err != nil {
		return nil, fmt.Errorf("failed to determine execution order: %w", err)
	}

	// Execute with parallelization where possible
	return o.executor.ExecuteParallel(ctx, orderedCalls)
}

// ValidateSequence validates that a sequence of actions can be executed
func (o *Operator) ValidateSequence(calls []actions.ActionCall) error {
	return o.registry.ValidateSequence(calls)
}

// GetAvailableActions returns all available actions, optionally filtered by category
func (o *Operator) GetAvailableActions(category string) []actions.ActionMetadata {
	var actionList []*actions.Action

	if category != "" {
		actionList = o.registry.ListByCategory(actions.Category(category))
	} else {
		actionList = o.registry.List()
	}

	metadata := make([]actions.ActionMetadata, len(actionList))
	for i, action := range actionList {
		metadata[i] = actions.ActionMetadata{
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			Tags:        action.Tags,
			Tier:        action.Tier,
		}
	}

	return metadata
}

// Executor handles the actual execution of actions
type Executor struct {
	registry *actions.Registry
}

// NewExecutor creates a new executor
func NewExecutor(registry *actions.Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// ExecuteSingle executes a single action
func (e *Executor) ExecuteSingle(ctx context.Context, call actions.ActionCall) (*actions.ActionResult, error) {
	return e.registry.Execute(ctx, call)
}

// ExecuteParallel executes actions in parallel where possible
func (e *Executor) ExecuteParallel(ctx context.Context, calls []actions.ActionCall) (*actions.ExecutionResult, error) {
	startTime := time.Now()
	result := &actions.ExecutionResult{
		Success: true,
		Actions: make([]actions.ActionResult, 0, len(calls)),
		Outputs: make(map[string]interface{}),
	}

	// Get parallel execution groups
	actionNames := make([]string, len(calls))
	callMap := make(map[string]actions.ActionCall)
	for i, call := range calls {
		actionNames[i] = call.Action
		callMap[call.Action] = call
	}

	graph := actions.NewDependencyGraph()
	for _, call := range calls {
		graph.AddNode(call.Action)
		action, _ := e.registry.Get(call.Action)
		if action != nil {
			for _, dep := range action.Dependencies {
				// Only add edge if dependency is in our execution set
				if _, exists := callMap[dep]; exists {
					graph.AddEdge(dep, call.Action)
				}
			}
		}
	}

	groups, err := graph.GetParallelGroups(actionNames)
	if err != nil {
		return nil, fmt.Errorf("failed to determine parallel groups: %w", err)
	}

	// Execute each group
	for _, group := range groups {
		// Execute actions in this group in parallel
		var wg sync.WaitGroup
		groupResults := make([]*actions.ActionResult, len(group))
		groupErrors := make([]error, len(group))

		for i, actionName := range group {
			wg.Add(1)
			go func(idx int, name string) {
				defer wg.Done()

				call := callMap[name]
				// Add outputs from previous actions to inputs
				enrichedInputs := make(map[string]interface{})
				for k, v := range call.Inputs {
					enrichedInputs[k] = v
				}
				// Add available outputs from previous actions
				for k, v := range result.Outputs {
					if _, exists := enrichedInputs[k]; !exists {
						enrichedInputs[k] = v
					}
				}
				call.Inputs = enrichedInputs

				actionResult, err := e.registry.Execute(ctx, call)
				if err != nil {
					groupErrors[idx] = err
				}
				groupResults[idx] = actionResult
			}(i, actionName)
		}

		wg.Wait()

		// Check for errors and collect results
		for i, err := range groupErrors {
			if err != nil {
				result.Success = false
				// Add failed action result
				result.Actions = append(result.Actions, actions.ActionResult{
					Action:  group[i],
					Success: false,
					Error:   err.Error(),
				})
				// Stop execution on error
				result.TotalDuration = time.Since(startTime)
				return result, nil
			}
		}

		// Add successful results
		for _, actionResult := range groupResults {
			if actionResult != nil {
				result.Actions = append(result.Actions, *actionResult)
				// Merge outputs
				for k, v := range actionResult.Outputs {
					result.Outputs[k] = v
				}
			}
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}