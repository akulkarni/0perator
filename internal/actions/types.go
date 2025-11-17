package actions

import (
	"context"
	"time"
)

// Tier represents the performance tier of an action
type Tier string

const (
	TierFast     Tier = "fast"     // < 30 seconds, optimized implementation
	TierFlexible Tier = "flexible" // 30s - 5min, may use LLM interpretation
)

// Category represents the type of action
type Category string

const (
	CategoryCreate Category = "create" // Create new projects/files
	CategorySetup  Category = "setup"  // Setup infrastructure/services
	CategoryAdd    Category = "add"    // Add features to existing projects
	CategoryDeploy Category = "deploy" // Deploy applications
	CategoryUtil   Category = "util"   // Utility actions
)

// InputType represents the type of an input parameter
type InputType string

const (
	InputTypeString  InputType = "string"
	InputTypeBool    InputType = "bool"
	InputTypeInt     InputType = "int"
	InputTypeFloat   InputType = "float"
	InputTypeArray   InputType = "array"
	InputTypeObject  InputType = "object"
)

// Input defines an input parameter for an action
type Input struct {
	Name        string      `json:"name" yaml:"name"`
	Type        InputType   `json:"type" yaml:"type"`
	Description string      `json:"description" yaml:"description"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []string    `json:"options,omitempty" yaml:"options,omitempty"` // For string type with predefined options
}

// Output defines an output value from an action
type Output struct {
	Name        string    `json:"name" yaml:"name"`
	Type        InputType `json:"type" yaml:"type"`
	Description string    `json:"description" yaml:"description"`
}

// ActionFunc is the function signature for action implementations
type ActionFunc func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)

// Action represents a single atomic operation
type Action struct {
	// Metadata
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description" yaml:"description"`
	Category     Category      `json:"category" yaml:"category"`
	Tags         []string      `json:"tags" yaml:"tags"`
	Tier         Tier          `json:"tier" yaml:"tier"`
	EstimatedTime time.Duration `json:"estimated_time" yaml:"estimated_time"`

	// Interface
	Inputs  []Input  `json:"inputs" yaml:"inputs"`
	Outputs []Output `json:"outputs" yaml:"outputs"`

	// Relationships
	Dependencies []string `json:"dependencies" yaml:"dependencies"` // Actions that must run before this one
	Conflicts    []string `json:"conflicts" yaml:"conflicts"`       // Actions that cannot coexist with this one

	// Implementation
	Implementation ActionFunc `json:"-" yaml:"-"` // The actual function to execute
}

// ActionCall represents a request to execute an action with specific inputs
type ActionCall struct {
	Action string                 `json:"action"`
	Inputs map[string]interface{} `json:"inputs"`
}

// ActionResult represents the result of executing an action
type ActionResult struct {
	Action   string                 `json:"action"`
	Success  bool                   `json:"success"`
	Outputs  map[string]interface{} `json:"outputs,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
}

// ExecutionResult represents the result of executing a sequence of actions
type ExecutionResult struct {
	Success       bool            `json:"success"`
	Actions       []ActionResult  `json:"actions"`
	TotalDuration time.Duration   `json:"total_duration"`
	Outputs       map[string]interface{} `json:"outputs"` // Combined outputs from all actions
}

// ActionMetadata provides a lightweight description of an action for discovery
type ActionMetadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     Category `json:"category"`
	Tags         []string `json:"tags"`
	Tier         Tier     `json:"tier"`
}

// Validate checks if the action definition is valid
func (a *Action) Validate() error {
	if a.Name == "" {
		return ErrInvalidAction("action name is required")
	}
	if a.Description == "" {
		return ErrInvalidAction("action description is required")
	}
	if a.Category == "" {
		return ErrInvalidAction("action category is required")
	}
	if a.Implementation == nil {
		return ErrInvalidAction("action implementation is required")
	}
	return nil
}

// ValidateInputs checks if the provided inputs match the action's requirements
func (a *Action) ValidateInputs(inputs map[string]interface{}) error {
	for _, input := range a.Inputs {
		value, exists := inputs[input.Name]
		if input.Required && !exists {
			return ErrMissingInput(input.Name)
		}
		if exists && input.Options != nil && len(input.Options) > 0 {
			// Validate that string value is one of the allowed options
			if strValue, ok := value.(string); ok {
				valid := false
				for _, option := range input.Options {
					if strValue == option {
						valid = true
						break
					}
				}
				if !valid {
					return ErrInvalidInput(input.Name, "must be one of: "+joinStrings(input.Options))
				}
			}
		}
	}
	return nil
}

// ApplyDefaults fills in default values for missing optional inputs
func (a *Action) ApplyDefaults(inputs map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing inputs
	for k, v := range inputs {
		result[k] = v
	}

	// Apply defaults for missing inputs
	for _, input := range a.Inputs {
		if _, exists := result[input.Name]; !exists && input.Default != nil {
			result[input.Name] = input.Default
		}
	}

	return result
}