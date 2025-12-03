package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// CreateNextJSApp creates a complete Next.js app with proper configuration,
// auto-installs dependencies, starts dev server, and opens browser
func CreateDatabase(ctx context.Context, dbName string) (string, error) {
	if dbName == "" {
		dbName = "my-app-db"
	}

	cmdArgs := []string{
		"service", "create",
		"--name", dbName,
		"--cpu", "shared",
		"--memory", "shared",
		"--addons", "time-series,ai",
		"--wait-timeout", "2m",
		"-o", "json",
	}

	cmd := exec.CommandContext(ctx, "tiger", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("failed to create database: %w\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	var createResult map[string]interface{}

	if err := json.Unmarshal([]byte(outputStr), &createResult); err != nil {
		return "", fmt.Errorf("failed to parse create database output: %w\nOutput: %s", err, string(output))
	}

	serviceId, ok := createResult["service_id"].(string)
	if !ok {
		return "", fmt.Errorf("no service_id in response: %s", string(output))
	}

	return serviceId, nil
}
