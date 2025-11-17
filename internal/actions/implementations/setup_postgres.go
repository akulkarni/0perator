package implementations

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
)

// SetupPostgresAction returns the action definition for setting up PostgreSQL
func SetupPostgresAction() *actions.Action {
	return &actions.Action{
		Name:         "setup_postgres",
		Description:  "Setup PostgreSQL database with optional TimescaleDB extension",
		Category:     actions.CategorySetup,
		Tags:         []string{"database", "postgres", "postgresql", "timescale", "sql"},
		Tier:         actions.TierFast,
		EstimatedTime: 30 * time.Second,

		Inputs: []actions.Input{
			{
				Name:        "database_name",
				Type:        actions.InputTypeString,
				Description: "Name of the database to create",
				Required:    false,
				Default:     "app_db",
			},
			{
				Name:        "port",
				Type:        actions.InputTypeInt,
				Description: "Port to run PostgreSQL on",
				Required:    false,
				Default:     5432,
			},
			{
				Name:        "username",
				Type:        actions.InputTypeString,
				Description: "Database username",
				Required:    false,
				Default:     "postgres",
			},
			{
				Name:        "password",
				Type:        actions.InputTypeString,
				Description: "Database password",
				Required:    false,
				Default:     "postgres",
			},
			{
				Name:        "with_timescale",
				Type:        actions.InputTypeBool,
				Description: "Include TimescaleDB extension",
				Required:    false,
				Default:     false,
			},
			{
				Name:        "docker",
				Type:        actions.InputTypeBool,
				Description: "Use Docker to run PostgreSQL",
				Required:    false,
				Default:     true,
			},
		},

		Outputs: []actions.Output{
			{
				Name:        "connection_string",
				Type:        actions.InputTypeString,
				Description: "PostgreSQL connection string",
			},
			{
				Name:        "host",
				Type:        actions.InputTypeString,
				Description: "Database host",
			},
			{
				Name:        "port",
				Type:        actions.InputTypeInt,
				Description: "Database port",
			},
			{
				Name:        "database_name",
				Type:        actions.InputTypeString,
				Description: "Database name",
			},
			{
				Name:        "container_name",
				Type:        actions.InputTypeString,
				Description: "Docker container name (if using Docker)",
			},
		},

		Dependencies: []string{},
		Conflicts:    []string{"setup_mysql", "setup_mongodb"}, // Can't setup multiple databases

		Implementation: setupPostgresImpl,
	}
}

func setupPostgresImpl(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	dbName := inputs["database_name"].(string)
	port := inputs["port"].(int)
	username := inputs["username"].(string)
	password := inputs["password"].(string)
	withTimescale := inputs["with_timescale"].(bool)
	useDocker := inputs["docker"].(bool)

	if !useDocker {
		return nil, fmt.Errorf("non-Docker PostgreSQL setup not yet implemented")
	}

	// Docker container name
	containerName := fmt.Sprintf("postgres_%s", dbName)

	// Choose image based on TimescaleDB requirement
	image := "postgres:15-alpine"
	if withTimescale {
		image = "timescale/timescaledb:latest-pg15"
	}

	// Check if container already exists
	checkCmd := exec.CommandContext(ctx, "docker", "ps", "-a", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	if output, _ := checkCmd.Output(); len(output) > 0 {
		// Container exists, try to start it
		startCmd := exec.CommandContext(ctx, "docker", "start", containerName)
		if err := startCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to start existing container: %w", err)
		}
	} else {
		// Create new container
		args := []string{
			"run", "-d",
			"--name", containerName,
			"-e", fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"-e", fmt.Sprintf("POSTGRES_USER=%s", username),
			"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
			"-p", fmt.Sprintf("%d:5432", port),
			image,
		}

		cmd := exec.CommandContext(ctx, "docker", args...)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL container: %w", err)
		}

		// Wait for PostgreSQL to be ready
		time.Sleep(3 * time.Second)

		// If TimescaleDB, enable the extension
		if withTimescale {
			enableCmd := exec.CommandContext(ctx, "docker", "exec", containerName,
				"psql", "-U", username, "-d", dbName, "-c", "CREATE EXTENSION IF NOT EXISTS timescaledb;")
			if err := enableCmd.Run(); err != nil {
				// Non-fatal, extension might already exist
				fmt.Printf("Warning: Could not enable TimescaleDB extension: %v\n", err)
			}
		}
	}

	// Build connection string
	connectionString := fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s?sslmode=disable",
		username, password, port, dbName)

	return map[string]interface{}{
		"connection_string": connectionString,
		"host":              "localhost",
		"port":              port,
		"database_name":     dbName,
		"container_name":    containerName,
	}, nil
}