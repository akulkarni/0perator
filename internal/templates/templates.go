package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:app/*
var appFS embed.FS

// WriteAppTemplates copies the app templates (CLAUDE.md, globals.css, etc.) to destination
func WriteAppTemplates(destDir string) error {
	return fs.WalkDir(appFS, "app", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from "app"
		relPath, err := filepath.Rel("app", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file content
		content, err := appFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
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
