package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// replaceEnvValue replaces the value of a variable in a .env file
func replaceEnvValue(envPath, key, value string) error {
	envData, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := strings.Split(string(envData), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = key + "=" + value
			break
		}
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// OpenBrowser opens the given URL in the default browser
func OpenBrowser(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Start()
}

// CreateNextJSApp creates a complete Next.js app with proper configuration,
// auto-installs dependencies, starts dev server, and opens browser
func CreateNextJSApp(ctx context.Context, name string, dbServiceID string) error {
	if name == "" {
		name = "my-app"
	}

	cmd := exec.CommandContext(ctx, "npm", "create", "t3-app@latest", "--", name, "--noGit", "--CI", "--tailwind", "--drizzle", "--trpc", "--dbProvider", "postgres", "--appRouter", "--betterAuth")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run create-t3-app: %w\n%s", err, output)
	}
	fmt.Fprintf(os.Stderr, "create-t3-app output: %s\n", string(output))

	if dbServiceID == "" {
		return fmt.Errorf("dbServiceID is required")
	}

	getCmd := exec.CommandContext(ctx, "tiger", "service", "get", dbServiceID, "--with-password", "-o", "json")
	getOutput, err := getCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get database service details: %w", err)
	}

	var serviceDetails map[string]any
	if err := json.Unmarshal(getOutput, &serviceDetails); err != nil {
		return fmt.Errorf("failed to parse database service details: %w", err)
	}

	connectionString, ok := serviceDetails["connection_string"].(string)
	if !ok || connectionString == "" {
		return fmt.Errorf("connection_string not found in service details")
	}

	envPath := filepath.Join(name, ".env")
	if err := replaceEnvValue(envPath, "DATABASE_URL", connectionString); err != nil {
		return err
	}

	return nil
}
