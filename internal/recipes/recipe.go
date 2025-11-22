package recipes

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Recipe represents a recipe that composes multiple tools
type Recipe struct {
	Name   string            `yaml:"name"`
	Desc   string            `yaml:"desc"`
	Inputs map[string]string `yaml:"inputs"` // "param: type|options = default"
	Steps  []string          `yaml:"steps"`
}

// ParsedInput represents a parsed input definition
type ParsedInput struct {
	Name    string
	Type    string   // "string", "number", "bool", or "enum"
	Options []string // For enums like ["jwt", "clerk", "auth0"]
	Default string   // Default value if not provided
}

// ParseInputDef parses an input definition like "jwt|clerk|auth0 = jwt" or "string = my-app"
func ParseInputDef(name, def string) ParsedInput {
	parts := strings.SplitN(def, "=", 2)

	// Get default if exists
	defaultVal := ""
	if len(parts) == 2 {
		defaultVal = strings.TrimSpace(parts[1])
	}

	// Parse type/options
	typeDef := strings.TrimSpace(parts[0])
	if strings.Contains(typeDef, "|") {
		// It's an enum
		options := []string{}
		for _, opt := range strings.Split(typeDef, "|") {
			options = append(options, strings.TrimSpace(opt))
		}
		return ParsedInput{
			Name:    name,
			Type:    "enum",
			Options: options,
			Default: defaultVal,
		}
	}

	// Simple type (string, number, bool)
	return ParsedInput{
		Name:    name,
		Type:    typeDef,
		Default: defaultVal,
	}
}

// LoadRecipe loads a recipe from a YAML file
func LoadRecipe(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe file: %w", err)
	}

	var recipe Recipe
	if err := yaml.Unmarshal(data, &recipe); err != nil {
		return nil, fmt.Errorf("failed to parse recipe YAML: %w", err)
	}

	// Validate recipe
	if recipe.Name == "" {
		return nil, fmt.Errorf("recipe missing 'name' field")
	}
	if recipe.Desc == "" {
		return nil, fmt.Errorf("recipe missing 'desc' field")
	}
	if len(recipe.Steps) == 0 {
		return nil, fmt.Errorf("recipe has no steps")
	}

	return &recipe, nil
}

// GetParsedInputs returns all inputs with their parsed definitions
func (r *Recipe) GetParsedInputs() []ParsedInput {
	inputs := []ParsedInput{}
	for name, def := range r.Inputs {
		inputs = append(inputs, ParseInputDef(name, def))
	}
	return inputs
}

// ValidateUserInputs validates user inputs against recipe requirements
func (r *Recipe) ValidateUserInputs(userInputs map[string]interface{}) (map[string]string, error) {
	validated := make(map[string]string)

	for name, def := range r.Inputs {
		parsed := ParseInputDef(name, def)

		// Get value from user or use default
		var value string
		if val, exists := userInputs[name]; exists {
			value = fmt.Sprint(val)
		} else if parsed.Default != "" {
			value = parsed.Default
		} else {
			// No value and no default - only error if it's actually used in steps
			if r.isInputUsed(name) {
				return nil, fmt.Errorf("required input '%s' not provided", name)
			}
			continue
		}

		// Validate enum values
		if parsed.Type == "enum" && len(parsed.Options) > 0 {
			valid := false
			for _, opt := range parsed.Options {
				if value == opt {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("input '%s' must be one of: %s", name, strings.Join(parsed.Options, ", "))
			}
		}

		validated[name] = value
	}

	return validated, nil
}

// isInputUsed checks if an input is actually used in any step
func (r *Recipe) isInputUsed(inputName string) bool {
	searchStr := "{{" + inputName + "}}"
	for _, step := range r.Steps {
		if strings.Contains(step, searchStr) {
			return true
		}
	}
	return false
}

// ReplaceVariables replaces {{var}} with actual values in text
func ReplaceVariables(text string, vars map[string]string) string {
	result := text
	for key, val := range vars {
		result = strings.ReplaceAll(result, "{{"+key+"}}", val)
	}
	return result
}

// ParseStep parses "tool_name arg1=val arg2={{var}}" into tool and args
func ParseStep(step string) (tool string, args map[string]string) {
	parts := strings.Fields(step)
	if len(parts) == 0 {
		return "", nil
	}

	tool = parts[0]
	args = make(map[string]string)

	for _, part := range parts[1:] {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			args[kv[0]] = kv[1]
		}
	}

	return tool, args
}