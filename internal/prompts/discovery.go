package prompts

import (
	"fmt"
	"sort"
	"strings"
)

// DiscoverPatterns finds templates matching the query using tag-based scoring
func DiscoverPatterns(query string) (*DiscoverResult, error) {
	templates, err := LoadTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Tokenize query
	queryWords := tokenize(query)

	// Score all templates
	scored := scoreTemplates(templates, queryWords)

	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Determine category from query and top matches
	category := inferCategory(queryWords, scored)

	// Get default template for category
	var defaultTemplate *Template
	if category != "" {
		// Check if user specified a particular option (e.g., "vercel" in "deploy to vercel")
		userSpecified := detectSpecificTemplate(queryWords, scored)
		if userSpecified != nil {
			defaultTemplate = userSpecified
		} else {
			// Use configured default for category
			defaultTemplate, _ = GetDefaultTemplate(category)
		}
	}

	// Convert to Pattern (lightweight response)
	matches := make([]Pattern, 0, len(scored))
	for _, t := range scored {
		if t.Score > 0 {
			matches = append(matches, Pattern{
				Name:        t.Name,
				Title:       t.Title,
				Description: t.Description,
				Tags:        t.Tags,
				Category:    t.Category,
				Score:       t.Score,
			})
		}
	}

	// Generate message
	message := generateDiscoveryMessage(query, matches, defaultTemplate)

	return &DiscoverResult{
		Query:   query,
		Matches: toTemplates(matches, templates),
		Default: defaultTemplate,
		Message: message,
	}, nil
}

// tokenize splits a query into lowercase words
func tokenize(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	result := make([]string, 0, len(words))
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if word != "" {
			result = append(result, word)
		}
	}
	return result
}

// scoredTemplate is a template with a relevance score
type scoredTemplate struct {
	Template
	Score float64
}

// scoreTemplates scores all templates based on query match
func scoreTemplates(templates map[string]Template, queryWords []string) []scoredTemplate {
	scored := make([]scoredTemplate, 0, len(templates))

	for _, tmpl := range templates {
		score := 0.0

		// Score based on tag matches
		for _, word := range queryWords {
			for _, tag := range tmpl.Tags {
				if strings.Contains(strings.ToLower(tag), word) {
					score += 2.0 // Tag match is high value
				}
			}

			// Score based on title match
			if strings.Contains(strings.ToLower(tmpl.Title), word) {
				score += 1.5
			}

			// Score based on description match
			if strings.Contains(strings.ToLower(tmpl.Description), word) {
				score += 1.0
			}

			// Score based on name match
			if strings.Contains(strings.ToLower(tmpl.Name), word) {
				score += 1.5
			}
		}

		scored = append(scored, scoredTemplate{
			Template: tmpl,
			Score:    score,
		})
	}

	return scored
}

// inferCategory determines the category from query and top matches
func inferCategory(queryWords []string, scored []scoredTemplate) string {
	// Direct category keywords
	categoryKeywords := map[string][]string{
		"deployment":     {"deploy", "deployment", "hosting", "production", "launch"},
		"database":       {"database", "db", "postgres", "sql", "data"},
		"authentication": {"auth", "authentication", "login", "signup", "user"},
		"payments":       {"payment", "payments", "billing", "stripe", "checkout"},
		"email":          {"email", "mail", "sendgrid", "resend"},
		"storage":        {"storage", "files", "upload", "s3", "r2"},
	}

	// Check query words for category keywords
	for category, keywords := range categoryKeywords {
		for _, word := range queryWords {
			for _, keyword := range keywords {
				if word == keyword {
					return category
				}
			}
		}
	}

	// Infer from top match category
	if len(scored) > 0 && scored[0].Score > 0 {
		return scored[0].Category
	}

	return ""
}

// detectSpecificTemplate checks if user specified a particular platform/option
func detectSpecificTemplate(queryWords []string, scored []scoredTemplate) *Template {
	// Look for specific platform names in query
	for _, word := range queryWords {
		for _, tmpl := range scored {
			// Check if word matches template name (e.g., "vercel" matches "deploy_vercel")
			if strings.Contains(tmpl.Name, word) && tmpl.Score > 0 {
				return &tmpl.Template
			}
		}
	}
	return nil
}

// generateDiscoveryMessage creates a helpful message for the user
func generateDiscoveryMessage(query string, matches []Pattern, defaultTemplate *Template) string {
	if len(matches) == 0 {
		return fmt.Sprintf("No templates found matching '%s'. Try a different search term.", query)
	}

	if defaultTemplate != nil {
		otherOptions := make([]string, 0)
		for _, m := range matches {
			if m.Name != defaultTemplate.Name && len(otherOptions) < 3 {
				otherOptions = append(otherOptions, m.Title)
			}
		}

		if len(otherOptions) > 0 {
			return fmt.Sprintf("Found %d templates. Defaulting to %s. Other options: %s",
				len(matches),
				defaultTemplate.Title,
				strings.Join(otherOptions, ", "))
		}
		return fmt.Sprintf("Found %d templates. Using %s.", len(matches), defaultTemplate.Title)
	}

	return fmt.Sprintf("Found %d templates matching '%s'.", len(matches), query)
}

// toTemplates converts Patterns back to full Templates
func toTemplates(patterns []Pattern, allTemplates map[string]Template) []Template {
	result := make([]Template, 0, len(patterns))
	for _, p := range patterns {
		if tmpl, ok := allTemplates[p.Name]; ok {
			result = append(result, tmpl)
		}
	}
	return result
}
