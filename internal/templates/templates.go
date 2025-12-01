package templates

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:web/*
var webFS embed.FS

// TemplateData holds variables for template rendering
type TemplateData struct {
	Name string
}

// WriteDir recursively copies/renders a directory from embedded FS to destination.
// Files ending in .tmpl are rendered with data (and .tmpl is stripped from filename).
// Other files are copied as-is.
func WriteDir(srcDir, destDir string, data TemplateData) error {
	return fs.WalkDir(webFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from srcDir
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file content
		content, err := webFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Check if it's a template file
		if strings.HasSuffix(path, ".tmpl") {
			// Strip .tmpl extension from destination
			destPath = strings.TrimSuffix(destPath, ".tmpl")

			// Parse and execute template (use [[ ]] delimiters to avoid conflict with JSX {{ }})
			tmpl, err := template.New(filepath.Base(path)).Delims("[[", "]]").Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("failed to execute template %s: %w", path, err)
			}
			content = buf.Bytes()
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		// Write file
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}

		return nil
	})
}
