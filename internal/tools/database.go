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

// SetupPostgresFree creates a free tier PostgreSQL database on Tiger Cloud
func SetupPostgresFree(ctx context.Context, args map[string]string) error {
	dbName := args["name"]
	if dbName == "" {
		dbName = "app_db"
	}

	// Build the tiger service create command
	// Free tier automatically includes both time-series and ai addons
	cmdArgs := []string{
		"service", "create",
		"--name", dbName,
		"--cpu", "shared", // Free tier - shared CPU
		"--no-wait",       // Don't wait for service to be ready
		"-o", "json",      // Output format as JSON
	}

	// Execute the Tiger CLI command
	cmd := exec.CommandContext(ctx, "tiger", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's an auth error or CLI not found
		outputStr := string(output)
		if strings.Contains(outputStr, "not authenticated") ||
			strings.Contains(outputStr, "auth login") ||
			strings.Contains(outputStr, "command not found") ||
			err.Error() == "exec: \"tiger\": executable file not found in $PATH" {

			// Log the issue but don't fail - let the user know what to do
			fmt.Printf("Note: Tiger CLI not available or not authenticated.\n")
			fmt.Printf("To create a database, please:\n")
			fmt.Printf("1. Install Tiger CLI: brew install tigerdata/tap/tiger\n")
			fmt.Printf("2. Authenticate: tiger auth login\n")
			fmt.Printf("3. Or use Tiger Cloud console: https://console.cloud.timescale.com\n")

			// Return success with instructions
			return nil
		}
		return fmt.Errorf("failed to create Tiger service: %w\nOutput: %s", err, string(output))
	}

	// Parse the JSON output to get service details
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err == nil {
		if serviceID, ok := result["service_id"].(string); ok {
			fmt.Printf("✅ Database '%s' created successfully (ID: %s)\n", dbName, serviceID)
			fmt.Printf("📝 Status: Provisioning (1-2 minutes)\n")
			fmt.Printf("🔗 Connection details: tiger service get %s\n", dbName)
		}
	} else {
		// Non-JSON output, probably success message
		fmt.Printf("✅ Database '%s' creation initiated\n", dbName)
	}

	return nil
}

// SetupSQLite creates a local SQLite database
func SetupSQLite(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "database.db"
	}

	path := args["path"]
	if path == "" {
		path = "."
	}

	dbPath := filepath.Join(path, name)

	// Create the database file
	file, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create SQLite database: %w", err)
	}
	file.Close()

	fmt.Printf("✅ SQLite database created at: %s\n", dbPath)

	// Optionally create a schema file
	schemaPath := filepath.Join(path, "schema.sql")
	schema := `-- Example schema for SQLite
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
);`

	if err := os.WriteFile(schemaPath, []byte(schema), 0644); err == nil {
		fmt.Printf("📝 Schema template created at: %s\n", schemaPath)
	}

	return nil
}

// CreateDatabaseSchema creates a database schema file
func CreateDatabaseSchema(ctx context.Context, args map[string]string) error {
	dbType := args["type"]
	if dbType == "" {
		dbType = "postgres"
	}

	path := args["path"]
	if path == "" {
		path = "."
	}

	var schema string
	filename := "schema.sql"

	switch dbType {
	case "postgres":
		schema = `-- PostgreSQL Schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table (for authentication)
CREATE TABLE sessions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Subscriptions table (for SaaS)
CREATE TABLE subscriptions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);`

	case "mysql":
		schema = `-- MySQL Schema

-- Users table
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Subscriptions table
CREATE TABLE subscriptions (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    current_period_end TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);`

	case "sqlite":
		schema = `-- SQLite Schema

-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    password_hash TEXT NOT NULL,
    email_verified INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Subscriptions table
CREATE TABLE subscriptions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL,
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    status TEXT NOT NULL,
    plan TEXT NOT NULL,
    current_period_end DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);`

	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	schemaPath := filepath.Join(path, filename)
	if err := os.WriteFile(schemaPath, []byte(schema), 0644); err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}

	fmt.Printf("✅ Database schema created at: %s\n", schemaPath)
	return nil
}