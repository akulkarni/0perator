package prompts

import (
	"testing"
)

func TestAllTemplatesLoad(t *testing.T) {
	templates, err := ListTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	expectedTemplates := []string{
		"create_web_app",
		"database_tiger",
		"auth_jwt",
		"email_resend",
		"payments_stripe",
		"deploy_cloudflare",
		"deploy_railway",
	}

	if len(templates) < len(expectedTemplates) {
		t.Errorf("Expected at least %d templates, got %d", len(expectedTemplates), len(templates))
	}

	for _, name := range expectedTemplates {
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == name {
				found = true
				// Verify template has required fields
				if tmpl.Title == "" {
					t.Errorf("Template %s missing title", name)
				}
				if tmpl.Description == "" {
					t.Errorf("Template %s missing description", name)
				}
				if len(tmpl.Tags) == 0 {
					t.Errorf("Template %s missing tags", name)
				}
				if tmpl.Category == "" {
					t.Errorf("Template %s missing category", name)
				}
				if tmpl.Content == "" {
					t.Errorf("Template %s missing content", name)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found", name)
		}
	}

	t.Logf("Successfully loaded %d templates", len(templates))
}

func TestDefaultTemplates(t *testing.T) {
	defaults := ListDefaults()

	expectedDefaults := map[string]string{
		"deployment":     "deploy_railway",
		"database":       "database_tiger",
		"authentication": "auth_jwt",
		"payments":       "payments_stripe",
		"email":          "email_resend",
	}

	for category, expectedTemplate := range expectedDefaults {
		actual, ok := defaults[category]
		if !ok {
			t.Errorf("Category %s not found in defaults", category)
			continue
		}
		if actual != expectedTemplate {
			t.Errorf("Category %s: expected %s, got %s", category, expectedTemplate, actual)
		}

		// Verify the default template exists and loads
		tmpl, err := GetDefaultTemplate(category)
		if err != nil {
			t.Errorf("Failed to load default template for %s: %v", category, err)
		}
		if tmpl.Name != expectedTemplate {
			t.Errorf("GetDefaultTemplate(%s) returned %s, expected %s", category, tmpl.Name, expectedTemplate)
		}
	}

	t.Logf("All %d default templates verified", len(expectedDefaults))
}

func TestDiscoverPatterns(t *testing.T) {
	testCases := []struct {
		query          string
		expectedInTop3 string
	}{
		{"create web app", "create_web_app"},
		{"database postgres", "database_tiger"},
		{"authentication jwt", "auth_jwt"},
		{"email sending", "email_resend"},
		{"payments stripe", "payments_stripe"},
		{"deploy static site", "deploy_cloudflare"},
		{"deploy backend server", "deploy_railway"},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			result, err := DiscoverPatterns(tc.query)
			if err != nil {
				t.Fatalf("DiscoverPatterns failed: %v", err)
			}

			if len(result.Matches) == 0 {
				t.Errorf("No matches found for query: %s", tc.query)
				return
			}

			// Check if expected template is in top 3
			found := false
			for i := 0; i < min(3, len(result.Matches)); i++ {
				if result.Matches[i].Name == tc.expectedInTop3 {
					found = true
					t.Logf("Query '%s' found '%s' at position %d", tc.query, tc.expectedInTop3, i+1)
					break
				}
			}

			if !found {
				t.Errorf("Expected template '%s' not in top 3 results for query '%s'", tc.expectedInTop3, tc.query)
				t.Logf("Top 3 results: %v", result.Matches[:min(3, len(result.Matches))])
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
