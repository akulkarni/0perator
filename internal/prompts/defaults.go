package prompts

import "fmt"

// DefaultTemplates maps categories to their default template names
var DefaultTemplates = map[string]string{
	"deployment":     "deploy_railway",    // Railway for servers (most common), Cloudflare for static sites
	"database":       "database_tiger",    // Tiger Cloud PostgreSQL/TimescaleDB
	"authentication": "auth_jwt",          // JWT-based authentication
	"payments":       "payments_stripe",   // Stripe for payments and subscriptions
	"email":          "email_resend",      // Resend for transactional emails
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
