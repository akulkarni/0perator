package implementations

import (
	"context"
	"fmt"

	"github.com/akulkarni/0perator/internal/actions"
	"github.com/akulkarni/0perator/internal/tools"
)

// SetupSQLiteAction creates a local SQLite database
func SetupSQLiteAction() *actions.Action {
	return &actions.Action{
		Name:        "setup_sqlite",
		Description: "Create a local SQLite database with zero configuration",
		Category:    actions.CategorySetup,
		Tags:        []string{"database", "sqlite", "local", "development"},
		Tier:        actions.Tier("free"),
		Inputs: []actions.Input{
			{
				Name:        "name",
				Type:        actions.InputTypeString,
				Required:    false,
				Default:     "database.db",
				Description: "Database filename",
			},
			{
				Name:        "path",
				Type:        actions.InputTypeString,
				Required:    false,
				Default:     ".",
				Description: "Directory path for the database",
			},
		},
		Outputs: []actions.Output{
			{
				Name:        "db_path",
				Type:        actions.InputTypeString,
				Description: "Full path to the SQLite database file",
			},
			{
				Name:        "schema_path",
				Type:        actions.InputTypeString,
				Description: "Path to the example schema file",
			},
		},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			// Extract inputs
			name, _ := inputs["name"].(string)
			if name == "" {
				name = "database.db"
			}

			path, _ := inputs["path"].(string)
			if path == "" {
				path = "."
			}

			// Prepare arguments for the tool
			args := map[string]string{
				"name": name,
				"path": path,
			}

			// Execute the SQLite setup
			if err := tools.SetupSQLite(ctx, args); err != nil {
				return nil, err
			}

			// Build the full path
			dbPath := fmt.Sprintf("%s/%s", path, name)
			schemaPath := fmt.Sprintf("%s/schema.sql", path)

			return map[string]interface{}{
				"db_path":     dbPath,
				"schema_path": schemaPath,
				"message":     fmt.Sprintf("SQLite database '%s' created successfully", name),
			}, nil
		},
	}
}