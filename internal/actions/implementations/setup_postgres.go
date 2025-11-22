package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
)

// SetupPostgresAction returns the action definition for setting up PostgreSQL
func SetupPostgresAction() *actions.Action {
	return &actions.Action{
		Name:         "setup_postgres",
		Description:  "Creates a FREE TIER PostgreSQL database on Tiger Cloud (shared CPU, instant) - uses Tiger CLI if available, otherwise guides Claude to use Tiger MCP",
		Category:     actions.CategorySetup,
		Tags:         []string{"database", "postgres", "postgresql", "tiger", "cloud", "timescale", "free"},
		Tier:         actions.TierFast,
		EstimatedTime: 3 * time.Second,

		Inputs: []actions.Input{
			{
				Name:        "database_name",
				Type:        actions.InputTypeString,
				Description: "Name of the database to create",
				Required:    false,
				Default:     "app_db",
			},
			{
				Name:        "with_timescale",
				Type:        actions.InputTypeBool,
				Description: "Include TimescaleDB extension (adds time-series addon)",
				Required:    false,
				Default:     true, // Default to including TimescaleDB for free tier
			},
		},

		Outputs: []actions.Output{
			{
				Name:        "success",
				Type:        actions.InputTypeBool,
				Description: "Whether the database was created successfully",
			},
			{
				Name:        "message",
				Type:        actions.InputTypeString,
				Description: "Status message",
			},
			{
				Name:        "service_id",
				Type:        actions.InputTypeString,
				Description: "Tiger service ID",
			},
			{
				Name:        "database_name",
				Type:        actions.InputTypeString,
				Description: "Database name",
			},
			{
				Name:        "tier",
				Type:        actions.InputTypeString,
				Description: "Service tier (free/shared CPU)",
			},
			{
				Name:        "status",
				Type:        actions.InputTypeString,
				Description: "Current status (provisioning/ready)",
			},
		},

		Dependencies: []string{},
		Conflicts:    []string{"setup_mysql", "setup_mongodb"}, // Can't setup multiple databases

		Implementation: setupPostgresImpl,
	}
}

func setupPostgresImpl(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	dbName := inputs["database_name"].(string)
	// Note: withTimescale is ignored for free tier - it always includes TimescaleDB
	// withTimescale := inputs["with_timescale"].(bool)

	// Build the tiger service create command
	// Use shared CPU for free tier (best for agentic ergonomics)
	// Note: Free tier automatically includes both time-series and ai addons
	// We can't specify addons for free tier - they're included by default
	args := []string{
		"service", "create",
		"--name", dbName,
		"--cpu", "shared", // Free tier - shared CPU (includes time-series + ai)
		"--no-wait", // Don't wait for service to be ready (faster)
		"-o", "json", // Output format as JSON
	}
	// Don't specify --addons for free tier - it automatically includes both

	// Execute the Tiger CLI command
	cmd := exec.CommandContext(ctx, "tiger", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's an auth error or CLI not found
		outputStr := string(output)
		if strings.Contains(outputStr, "not authenticated") ||
		   strings.Contains(outputStr, "auth login") ||
		   strings.Contains(outputStr, "command not found") ||
		   err.Error() == "exec: \"tiger\": executable file not found in $PATH" {
			// Fall back to returning instructions for Claude to use Tiger MCP
			return map[string]interface{}{
				"success": false,
				"message": "Tiger CLI not available or not authenticated. Falling back to MCP instructions.",
				"action_required": "use_tiger_mcp",
				"instructions": fmt.Sprintf(
					"Please use the Tiger MCP tool to create a FREE TIER PostgreSQL database:\n" +
					"- Database name: %s\n" +
					"- Use DEFAULT free tier settings (shared CPU)\n" +
					"- Free tier automatically includes time-series and ai addons\n" +
					"DO NOT ask the user - just create with free tier.",
					dbName,
				),
				"database_name": dbName,
				"recommended_tool": "mcp__tiger__service_create",
				"recommended_params": map[string]interface{}{
					"name": dbName,
					// Don't specify addons or cpu_memory - free tier defaults include everything
					"wait": false,
					"set_default": true,
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to create Tiger service: %w\nOutput: %s", err, string(output))
	}

	// Parse the JSON output to get service details
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// If JSON parsing fails, still return success with basic info
		return map[string]interface{}{
			"success":       true,
			"message":       fmt.Sprintf("Database '%s' created successfully on Tiger Cloud (free tier)", dbName),
			"database_name": dbName,
			"tier":          "free (shared CPU)",
			"addons":        "time-series, ai", // Free tier always includes both
			"status":        "provisioning",
			"next_steps": []string{
				"Database is provisioning (takes 1-2 minutes)",
				"Use 'tiger service list' to check status",
				fmt.Sprintf("Use 'tiger service get %s' for connection details", dbName),
			},
		}, nil
	}

	// Extract key information from the result
	serviceID := ""
	if id, ok := result["id"].(string); ok {
		serviceID = id
	}

	// Return success with service details
	return map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Database '%s' created successfully on Tiger Cloud (free tier)", dbName),
		"service_id":    serviceID,
		"database_name": dbName,
		"tier":          "free (shared CPU)",
		"addons":        "time-series, ai", // Free tier always includes both
		"status":        "provisioning",
		"region":        result["region"],
		"next_steps": []string{
			"Database is provisioning (takes 1-2 minutes)",
			fmt.Sprintf("Connection string will be available at: tiger service get %s", dbName),
		},
	}, nil
}