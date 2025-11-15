package prompts

import "fmt"

// DefaultTemplates maps categories to their default template names
// NOTE: Update this map after creating the initial templates
var DefaultTemplates = map[string]string{
	// Will be populated after templates are created:
	// "deployment":     "deploy_cloudflare",
	// "database":       "database_tiger",
	// "authentication": "auth_jwt",
	// "payments":       "payments_stripe",
}

// GetDefaultTemplate returns the default template for a category
func GetDefaultTemplate(category string) (*Template, error) {
	templateName, ok := DefaultTemplates[category]
	if !ok {
		return nil, fmt.Errorf("no default template for category: %s", category)
	}

	return GetTemplate(templateName)
}

// ListDefaults returns all configured defaults
func ListDefaults() map[string]string {
	result := make(map[string]string)
	for k, v := range DefaultTemplates {
		result[k] = v
	}
	return result
}

// SetDefault sets the default template for a category (for user configuration in future)
func SetDefault(category, templateName string) error {
	// Validate that template exists
	_, err := GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("cannot set default: %w", err)
	}

	DefaultTemplates[category] = templateName
	return nil
}
