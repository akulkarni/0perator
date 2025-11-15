package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed md/*.md
var templatesFS embed.FS

var (
	templatesCache     map[string]Template
	templatesCacheLock sync.RWMutex
)

// LoadTemplates loads all templates from the embedded filesystem
func LoadTemplates() (map[string]Template, error) {
	// Check cache first
	templatesCacheLock.RLock()
	if templatesCache != nil {
		defer templatesCacheLock.RUnlock()
		return templatesCache, nil
	}
	templatesCacheLock.RUnlock()

	// Load templates
	templatesCacheLock.Lock()
	defer templatesCacheLock.Unlock()

	templates := make(map[string]Template)

	entries, err := fs.ReadDir(templatesFS, "md")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		content, err := fs.ReadFile(templatesFS, filepath.Join("md", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}

		template, err := parseTemplate(name, string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		templates[name] = template
	}

	templatesCache = templates
	return templates, nil
}

// parseTemplate parses a template file with frontmatter
func parseTemplate(name, content string) (Template, error) {
	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return Template{}, fmt.Errorf("invalid template format: missing frontmatter")
	}

	// Parse frontmatter
	var tmpl Template
	if err := yaml.Unmarshal([]byte(parts[1]), &tmpl); err != nil {
		return Template{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set name and content
	tmpl.Name = name
	tmpl.Content = strings.TrimSpace(parts[2])

	// Validate required fields
	if tmpl.Title == "" {
		return Template{}, fmt.Errorf("missing required field: title")
	}
	if tmpl.Description == "" {
		return Template{}, fmt.Errorf("missing required field: description")
	}
	if len(tmpl.Tags) == 0 {
		return Template{}, fmt.Errorf("missing required field: tags")
	}
	if tmpl.Category == "" {
		return Template{}, fmt.Errorf("missing required field: category")
	}

	return tmpl, nil
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*Template, error) {
	templates, err := LoadTemplates()
	if err != nil {
		return nil, err
	}

	tmpl, ok := templates[name]
	if !ok {
		available := make([]string, 0, len(templates))
		for k := range templates {
			available = append(available, k)
		}
		return nil, fmt.Errorf("template not found: %s. Available: %s", name, strings.Join(available, ", "))
	}

	return &tmpl, nil
}

// ListTemplates returns all available templates
func ListTemplates() ([]Template, error) {
	templates, err := LoadTemplates()
	if err != nil {
		return nil, err
	}

	result := make([]Template, 0, len(templates))
	for _, tmpl := range templates {
		result = append(result, tmpl)
	}
	return result, nil
}
