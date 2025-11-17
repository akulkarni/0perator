package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Registry manages all available actions
type Registry struct {
	mu      sync.RWMutex
	actions map[string]*Action
	graph   *DependencyGraph
}

// NewRegistry creates a new action registry
func NewRegistry() *Registry {
	return &Registry{
		actions: make(map[string]*Action),
		graph:   NewDependencyGraph(),
	}
}

// Register adds a new action to the registry
func (r *Registry) Register(action *Action) error {
	if err := action.Validate(); err != nil {
		return fmt.Errorf("invalid action: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[action.Name]; exists {
		return fmt.Errorf("action '%s' already registered", action.Name)
	}

	r.actions[action.Name] = action

	// Update dependency graph
	r.graph.AddNode(action.Name)
	for _, dep := range action.Dependencies {
		r.graph.AddEdge(dep, action.Name)
	}

	return nil
}

// Get retrieves an action by name
func (r *Registry) Get(name string) (*Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	if !exists {
		return nil, ErrActionNotFound(name)
	}
	return action, nil
}

// List returns all registered actions
func (r *Registry) List() []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions := make([]*Action, 0, len(r.actions))
	for _, action := range r.actions {
		actions = append(actions, action)
	}
	return actions
}

// ListByCategory returns actions filtered by category
func (r *Registry) ListByCategory(category Category) []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actions []*Action
	for _, action := range r.actions {
		if action.Category == category {
			actions = append(actions, action)
		}
	}
	return actions
}

// Search finds actions matching a query
func (r *Registry) Search(query string) []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []*Action

	for _, action := range r.actions {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(action.Name), query) ||
			strings.Contains(strings.ToLower(action.Description), query) {
			matches = append(matches, action)
			continue
		}

		// Search in tags
		for _, tag := range action.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, action)
				break
			}
		}
	}

	return matches
}

// ValidateSequence checks if a sequence of actions can be executed
func (r *Registry) ValidateSequence(calls []ActionCall) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check all actions exist
	actionNames := make([]string, len(calls))
	for i, call := range calls {
		if _, exists := r.actions[call.Action]; !exists {
			return ErrActionNotFound(call.Action)
		}
		actionNames[i] = call.Action
	}

	// Check for dependency cycles
	if err := r.graph.ValidateSequence(actionNames); err != nil {
		return err
	}

	// Check for conflicts
	executedActions := make(map[string]bool)
	for _, call := range calls {
		action := r.actions[call.Action]

		// Check conflicts with previously executed actions
		for _, conflict := range action.Conflicts {
			if executedActions[conflict] {
				return ErrConflict(call.Action, conflict)
			}
		}

		executedActions[call.Action] = true
	}

	return nil
}

// GetExecutionOrder returns the optimal execution order for a set of actions
func (r *Registry) GetExecutionOrder(calls []ActionCall) ([]ActionCall, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First validate the sequence
	if err := r.ValidateSequence(calls); err != nil {
		return nil, err
	}

	// Build a temporary graph for this execution
	tempGraph := NewDependencyGraph()
	actionMap := make(map[string]ActionCall)

	for _, call := range calls {
		tempGraph.AddNode(call.Action)
		actionMap[call.Action] = call

		action := r.actions[call.Action]
		for _, dep := range action.Dependencies {
			// Only add edge if dependency is in our execution set
			for _, c := range calls {
				if c.Action == dep {
					tempGraph.AddEdge(dep, call.Action)
					break
				}
			}
		}
	}

	// Get topological sort
	order, err := tempGraph.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Convert back to ActionCall slice
	result := make([]ActionCall, 0, len(order))
	for _, name := range order {
		if call, exists := actionMap[name]; exists {
			result = append(result, call)
		}
	}

	return result, nil
}

// GetMetadata returns lightweight metadata for all actions
func (r *Registry) GetMetadata() []ActionMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]ActionMetadata, 0, len(r.actions))
	for _, action := range r.actions {
		metadata = append(metadata, ActionMetadata{
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			Tags:        action.Tags,
			Tier:        action.Tier,
		})
	}
	return metadata
}

// Execute runs a single action
func (r *Registry) Execute(ctx context.Context, call ActionCall) (*ActionResult, error) {
	action, err := r.Get(call.Action)
	if err != nil {
		return nil, err
	}

	// Validate inputs
	if err := action.ValidateInputs(call.Inputs); err != nil {
		return nil, err
	}

	// Apply defaults
	inputs := action.ApplyDefaults(call.Inputs)

	// Execute the action
	start := time.Now()
	outputs, err := action.Implementation(ctx, inputs)
	duration := time.Since(start)

	if err != nil {
		return &ActionResult{
			Action:   call.Action,
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
		}, nil
	}

	return &ActionResult{
		Action:   call.Action,
		Success:  true,
		Outputs:  outputs,
		Duration: duration,
	}, nil
}