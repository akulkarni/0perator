package actions

import "fmt"

// ActionError represents an error that occurred during action processing
type ActionError struct {
	Type    string
	Message string
	Action  string
	Input   string
}

func (e *ActionError) Error() string {
	if e.Action != "" {
		return fmt.Sprintf("%s [action: %s]: %s", e.Type, e.Action, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Common error constructors

func ErrActionNotFound(name string) error {
	return &ActionError{
		Type:    "ActionNotFound",
		Message: fmt.Sprintf("action '%s' not found in registry", name),
		Action:  name,
	}
}

func ErrInvalidAction(msg string) error {
	return &ActionError{
		Type:    "InvalidAction",
		Message: msg,
	}
}

func ErrMissingInput(input string) error {
	return &ActionError{
		Type:    "MissingInput",
		Message: fmt.Sprintf("required input '%s' not provided", input),
		Input:   input,
	}
}

func ErrInvalidInput(input, msg string) error {
	return &ActionError{
		Type:    "InvalidInput",
		Message: fmt.Sprintf("invalid input '%s': %s", input, msg),
		Input:   input,
	}
}

func ErrActionFailed(action, msg string) error {
	return &ActionError{
		Type:    "ActionFailed",
		Message: msg,
		Action:  action,
	}
}

func ErrDependencyFailed(action, dependency string) error {
	return &ActionError{
		Type:    "DependencyFailed",
		Message: fmt.Sprintf("dependency '%s' failed", dependency),
		Action:  action,
	}
}

func ErrConflict(action1, action2 string) error {
	return &ActionError{
		Type:    "ConflictDetected",
		Message: fmt.Sprintf("action '%s' conflicts with '%s'", action1, action2),
		Action:  action1,
	}
}

func ErrCyclicDependency(actions []string) error {
	return &ActionError{
		Type:    "CyclicDependency",
		Message: fmt.Sprintf("cyclic dependency detected: %v", actions),
	}
}

// Helper function
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}