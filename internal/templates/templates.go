package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed web-node/**
var webNodeFS embed.FS

// Template represents an application template
type Template struct {
	Name        string
	Description string
	Files       embed.FS
}

// ScaffoldOptions contains options for scaffolding
type ScaffoldOptions struct {
	AppName     string
	Description string
	DatabaseURL string
	OutputDir   string
}

// Available templates
var templates = map[string]Template{
	"web-node": {
		Name:        "web-node",
		Description: "Full-stack web application with Node.js/TypeScript backend and React frontend",
		Files:       webNodeFS,
	},
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*Template, error) {
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

// Scaffold creates a new application from a template
func Scaffold(templateName string, opts ScaffoldOptions) error {
	tmpl, err := GetTemplate(templateName)
	if err != nil {
		return err
	}

	// Create output directory
	outputPath := filepath.Join(opts.OutputDir, opts.AppName)
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check if directory already has files
	entries, err := os.ReadDir(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("directory already exists and is not empty: %s", outputPath)
	}

	// Walk through template files and copy them
	templateRoot := filepath.Join(templateName)
	err = walkEmbedFS(tmpl.Files, templateRoot, func(path string, content []byte, isDir bool) error {
		// Calculate relative path from template root
		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(outputPath, relPath)

		if isDir {
			return os.MkdirAll(targetPath, 0755)
		}

		// Process template variables
		processedContent, err := processTemplate(string(content), opts)
		if err != nil {
			return fmt.Errorf("failed to process template %s: %w", relPath, err)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Write file
		return os.WriteFile(targetPath, []byte(processedContent), 0644)
	})

	if err != nil {
		return fmt.Errorf("failed to scaffold template: %w", err)
	}

	return nil
}

// processTemplate processes template variables in content
func processTemplate(content string, opts ScaffoldOptions) (string, error) {
	tmpl, err := template.New("content").Parse(content)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	data := map[string]string{
		"AppName":     opts.AppName,
		"Description": opts.Description,
		"DatabaseURL": opts.DatabaseURL,
	}

	if err := tmpl.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// walkEmbedFS walks an embedded filesystem
func walkEmbedFS(fsys embed.FS, root string, fn func(path string, content []byte, isDir bool) error) error {
	entries, err := fsys.ReadDir(root)
	if err != nil {
		return err
	}

	// Process root directory
	if err := fn(root, nil, true); err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())

		if entry.IsDir() {
			if err := walkEmbedFS(fsys, path, fn); err != nil {
				return err
			}
		} else {
			content, err := fsys.ReadFile(path)
			if err != nil {
				return err
			}
			if err := fn(path, content, false); err != nil {
				return err
			}
		}
	}

	return nil
}

// ListTemplates returns a list of available templates
func ListTemplates() []Template {
	result := make([]Template, 0, len(templates))
	for _, tmpl := range templates {
		result = append(result, tmpl)
	}
	return result
}
