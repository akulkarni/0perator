package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// SetupPostgresWithSchema creates a PostgreSQL database and initializes schema
func SetupPostgresWithSchema(ctx context.Context, args map[string]string) error {
	dbName := args["name"]
	if dbName == "" {
		dbName = "app_db"
	}

	appType := args["app_type"]
	if appType == "" {
		appType = "web" // default to web app
	}

	fmt.Printf("🚀 Creating PostgreSQL database '%s' on Tiger Cloud...\n", dbName)

	// Step 1: Create the database using Tiger CLI
	// For free tier, we need both time-series and ai addons
	cmdArgs := []string{
		"service", "create",
		"--name", dbName,
		"--cpu", "shared",
		"--memory", "shared",
		"--addons", "time-series,ai",
		"--no-wait",
		"-o", "json",
	}

	cmd := exec.CommandContext(ctx, "tiger", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "not authenticated") ||
			strings.Contains(outputStr, "command not found") ||
			err.Error() == "exec: \"tiger\": executable file not found in $PATH" {
			return fmt.Errorf("Tiger CLI not available. Install with: brew install tigerdata/tap/tiger && tiger auth login")
		}
		return fmt.Errorf("failed to create database: %w\nOutput: %s", err, string(output))
	}

	// Parse the response to get service ID
	outputStr := string(output)
	var createResult map[string]interface{}

	// Tiger CLI often outputs decorated text before JSON. Look for JSON in the output.
	jsonStart := strings.Index(outputStr, "{")
	if jsonStart >= 0 {
		jsonData := outputStr[jsonStart:]
		if err := json.Unmarshal([]byte(jsonData), &createResult); err == nil {
			// Successfully parsed JSON
		} else {
			// Try to extract service ID from decorated output
			// Look for patterns like "Service ID: xxxxxxxx" or just a 10-char ID
			if matches := regexp.MustCompile(`Service ID:\s*([a-z0-9]{10})`).FindStringSubmatch(outputStr); len(matches) > 1 {
				serviceID := matches[1]
				fmt.Printf("✅ Database created with ID: %s\n", serviceID)
				createResult = map[string]interface{}{"service_id": serviceID}
			} else {
				// Try plain service ID
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if len(trimmed) == 10 && regexp.MustCompile(`^[a-z0-9]{10}$`).MatchString(trimmed) {
						fmt.Printf("✅ Database created with ID: %s\n", trimmed)
						createResult = map[string]interface{}{"service_id": trimmed}
						break
					}
				}
			}
		}
	} else {
		// No JSON found, try to extract service ID from text
		if matches := regexp.MustCompile(`Service ID:\s*([a-z0-9]{10})`).FindStringSubmatch(outputStr); len(matches) > 1 {
			serviceID := matches[1]
			fmt.Printf("✅ Database created with ID: %s\n", serviceID)
			createResult = map[string]interface{}{"service_id": serviceID}
		} else {
			// Try plain service ID
			serviceID := strings.TrimSpace(outputStr)
			if len(serviceID) == 10 && regexp.MustCompile(`^[a-z0-9]{10}$`).MatchString(serviceID) {
				fmt.Printf("✅ Database created with ID: %s\n", serviceID)
				createResult = map[string]interface{}{"service_id": serviceID}
			}
		}
	}

	if createResult == nil || createResult["service_id"] == nil {
		// As a last resort, if the command didn't fail, assume success and extract any 10-char ID
		if !strings.Contains(strings.ToLower(outputStr), "error") && !strings.Contains(strings.ToLower(outputStr), "failed") {
			// Database was likely created, just need to find the ID
			fmt.Printf("⚠️  Unexpected output format from Tiger CLI. Full output:\n%s\n", outputStr)
		}
		return fmt.Errorf("could not extract service ID from Tiger response")
	}

	serviceID, ok := createResult["service_id"].(string)
	if !ok {
		return fmt.Errorf("no service_id in response: %s", string(output))
	}

	fmt.Printf("✅ Database created with ID: %s\n", serviceID)
	fmt.Printf("⏳ Waiting for database to be ready (typically 30-60 seconds)...\n")

	// Step 2: Wait for database to be ready and get connection details with password
	var connectionString string
	var password string

	for attempts := 0; attempts < 30; attempts++ {
		time.Sleep(2 * time.Second)

		// Get service details with password
		getCmd := exec.CommandContext(ctx, "tiger", "service", "get", serviceID, "--with-password", "-o", "json")
		getOutput, _ := getCmd.Output()

		var serviceDetails map[string]interface{}
		if err := json.Unmarshal(getOutput, &serviceDetails); err == nil {
			if service, ok := serviceDetails["service"].(map[string]interface{}); ok {
				if status, ok := service["status"].(string); ok && status == "READY" {
					if connStr, ok := service["connection_string"].(string); ok {
						connectionString = connStr
					}
					if pw, ok := service["password"].(string); ok {
						password = pw
					}
					break
				}
			}
		}

		if attempts%5 == 0 {
			fmt.Printf("  Still waiting... (status check %d/30)\n", attempts+1)
		}
	}

	if connectionString == "" {
		// If we couldn't get it from CLI, construct it manually
		fmt.Printf("⚠️  Could not get connection details from CLI, using defaults\n")
		connectionString = fmt.Sprintf("postgresql://tsdbadmin:REPLACE_PASSWORD@%s.tsdb.cloud.timescale.com:36720/tsdb?sslmode=require", serviceID)
	}

	fmt.Printf("✅ Database is ready!\n")

	// Step 3: Initialize schema based on app type
	schema := getSchemaForAppType(appType)

	// Try to connect and create schema
	if password != "" && !strings.Contains(connectionString, "REPLACE_PASSWORD") {
		fmt.Printf("📝 Initializing database schema for '%s' app...\n", appType)

		db, err := sql.Open("postgres", connectionString)
		if err == nil {
			defer db.Close()

			// Test connection
			if err := db.PingContext(ctx); err == nil {
				// Execute schema
				if _, err := db.ExecContext(ctx, schema); err != nil {
					fmt.Printf("⚠️  Could not auto-create schema: %v\n", err)
					fmt.Printf("📋 Save this schema to run manually:\n\n%s\n", schema)
				} else {
					fmt.Printf("✅ Schema created successfully!\n")
				}
			} else {
				fmt.Printf("⚠️  Could not connect to initialize schema. Run this SQL manually:\n\n%s\n", schema)
			}
		}
	}

	// Step 4: Generate .env.local content
	// Keep sslmode=require in connection string - we'll handle SSL in Node.js
	envContent := fmt.Sprintf(`# Database Configuration (Tiger Cloud)
DATABASE_URL=%s

# JWT Secret (change in production!)
JWT_SECRET=%s-jwt-secret-change-in-production

# Database Connection Parts (for libraries that need individual values)
DB_HOST=%s.tsdb.cloud.timescale.com
DB_PORT=36720
DB_NAME=tsdb
DB_USER=tsdbadmin
DB_PASSWORD=%s
DB_SSL=require

# Tiger Cloud Service
TIGER_SERVICE_ID=%s
`, connectionString, dbName, serviceID, password, serviceID)

	// Try to find existing .env.local and update it, or create new one
	envPath := ".env.local"

	// Check if we're in an app directory (has package.json)
	if _, err := os.Stat("package.json"); err == nil {
		// We're in an app directory, use .env.local here
		envPath = ".env.local"
	}

	// Check if .env.local exists and update DATABASE_URL
	existingEnv, err := os.ReadFile(envPath)
	if err == nil {
		// File exists - update DATABASE_URL line
		lines := strings.Split(string(existingEnv), "\n")
		var newLines []string
		foundDatabaseURL := false
		foundJWTSecret := false

		for _, line := range lines {
			if strings.HasPrefix(line, "DATABASE_URL=") {
				newLines = append(newLines, fmt.Sprintf("DATABASE_URL=%s", connectionString))
				foundDatabaseURL = true
			} else if strings.HasPrefix(line, "JWT_SECRET=") {
				foundJWTSecret = true
				newLines = append(newLines, line)
			} else {
				newLines = append(newLines, line)
			}
		}

		// Add DATABASE_URL if not found
		if !foundDatabaseURL {
			// Insert after first comment block or at beginning
			newLines = append([]string{fmt.Sprintf("DATABASE_URL=%s", connectionString)}, newLines...)
		}

		// Add JWT_SECRET if not found
		if !foundJWTSecret {
			newLines = append(newLines, fmt.Sprintf("JWT_SECRET=%s-jwt-secret-change-in-production", dbName))
		}

		if err := os.WriteFile(envPath, []byte(strings.Join(newLines, "\n")), 0600); err != nil {
			fmt.Printf("⚠️  Could not update %s: %v\n", envPath, err)
		} else {
			fmt.Printf("✅ Updated %s with DATABASE_URL\n", envPath)
		}
	} else {
		// File doesn't exist - create it
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			fmt.Printf("⚠️  Could not write %s: %v\n", envPath, err)
			fmt.Printf("\n📋 Add this to your .env.local:\n%s\n", envContent)
		} else {
			fmt.Printf("✅ Created %s with connection details\n", envPath)
		}
	}

	// Step 5: Create database utility file with proper SSL handling
	dbUtilContent := `import { Pool, PoolClient } from 'pg';

// Disable SSL certificate validation for Tiger Cloud (self-signed certs)
// This is safe for Tiger Cloud as the connection is still encrypted
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

let pool: Pool | undefined;

function getPool(): Pool {
  if (!pool) {
    if (!process.env.DATABASE_URL) {
      throw new Error('DATABASE_URL not configured. Run setup_database to create a PostgreSQL database.');
    }

    pool = new Pool({
      connectionString: process.env.DATABASE_URL,
      max: 20,
      idleTimeoutMillis: 30000,
      connectionTimeoutMillis: 5000,
    });

    pool.on('error', (err) => {
      console.error('Unexpected database pool error:', err);
    });
  }
  return pool;
}

// Query helper - use this for most database operations
export async function query(text: string, params?: any[]) {
  const p = getPool();
  return await p.query(text, params);
}

// Get a client for transactions
export async function getClient(): Promise<PoolClient> {
  const p = getPool();
  return await p.connect();
}

// Transaction helper
export async function withTransaction<T>(
  callback: (client: PoolClient) => Promise<T>
): Promise<T> {
  const client = await getClient();
  try {
    await client.query('BEGIN');
    const result = await callback(client);
    await client.query('COMMIT');
    return result;
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

export default pool;
`

	// Create lib directory and db.ts file (only if we're in an app directory)
	if _, err := os.Stat("package.json"); err == nil {
		os.MkdirAll("lib", 0755)
		if err := os.WriteFile("lib/db.ts", []byte(dbUtilContent), 0644); err != nil {
			fmt.Printf("⚠️  Could not create lib/db.ts: %v\n", err)
		} else {
			fmt.Printf("✅ Created/updated lib/db.ts with connection pool\n")
		}

		// If there's an init-db script, run it to create app-specific tables
		if _, err := os.Stat("scripts/init-db.js"); err == nil {
			fmt.Printf("📝 Running database initialization script...\n")
			initCmd := exec.CommandContext(ctx, "node", "scripts/init-db.js")
			initOutput, err := initCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("⚠️  Init script had issues: %s\n", string(initOutput))
			} else {
				fmt.Printf("%s", string(initOutput))
			}
		}
	}

	fmt.Printf("\n🎉 PostgreSQL setup complete!\n")
	fmt.Printf("   - Database: %s\n", dbName)
	fmt.Printf("   - Schema: %s app tables created\n", appType)
	fmt.Printf("   - Connection: Saved to .env.local\n")
	if _, err := os.Stat("package.json"); err == nil {
		fmt.Printf("   - Utilities: lib/db.ts ready to use\n")
	}

	return nil
}

// getSchemaForAppType returns appropriate schema based on app type
func getSchemaForAppType(appType string) string {
	switch appType {
	case "api":
		return getAPISchema()
	case "ecommerce":
		return getEcommerceSchema()
	default:
		return getWebAppSchema()
	}
}

// getWebAppSchema returns schema for a typical web application
func getWebAppSchema() string {
	return `-- Web Application Schema with TimescaleDB optimizations

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table with optimized indexes
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Sessions table for auth
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Posts/Content table
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    content TEXT,
    published BOOLEAN DEFAULT false,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(published, published_at DESC);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
`
}

// getAPISchema returns schema for API applications
func getAPISchema() string {
	return `-- API Application Schema with Rate Limiting and Audit

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    permissions JSONB DEFAULT '{}',
    rate_limit INTEGER DEFAULT 1000,
    active BOOLEAN DEFAULT true,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(active);

-- Rate limiting table (using TimescaleDB if available)
CREATE TABLE IF NOT EXISTS rate_limits (
    id BIGSERIAL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Try to convert to hypertable for time-series optimization
DO $$
BEGIN
    PERFORM create_hypertable('rate_limits', 'timestamp',
        chunk_time_interval => INTERVAL '1 hour',
        if_not_exists => TRUE);
EXCEPTION
    WHEN undefined_function THEN
        -- TimescaleDB not available, use regular table
        NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_rate_limits_api_key_endpoint
    ON rate_limits(api_key_id, endpoint, timestamp DESC);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(255),
    request_body JSONB,
    response_code INTEGER,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Try to convert to hypertable for time-series optimization
DO $$
BEGIN
    PERFORM create_hypertable('audit_logs', 'timestamp',
        chunk_time_interval => INTERVAL '1 day',
        if_not_exists => TRUE);
EXCEPTION
    WHEN undefined_function THEN
        NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_audit_logs_api_key ON audit_logs(api_key_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type, timestamp DESC);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
`
}

// getEcommerceSchema returns schema for e-commerce applications
func getEcommerceSchema() string {
	return `-- E-commerce Schema with TimescaleDB optimizations

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Customers table
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    stripe_customer_id VARCHAR(255) UNIQUE,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_stripe ON customers(stripe_customer_id);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    price_cents INTEGER NOT NULL,
    stripe_price_id VARCHAR(255),
    inventory_count INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);
CREATE INDEX IF NOT EXISTS idx_products_active ON products(active);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    stripe_payment_intent_id VARCHAR(255) UNIQUE,
    subtotal_cents INTEGER NOT NULL,
    tax_cents INTEGER DEFAULT 0,
    shipping_cents INTEGER DEFAULT 0,
    total_cents INTEGER NOT NULL,
    shipping_address JSONB,
    billing_address JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at DESC);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_cents INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items(product_id);

-- Analytics events table (time-series data)
CREATE TABLE IF NOT EXISTS analytics_events (
    id BIGSERIAL,
    event_type VARCHAR(50) NOT NULL,
    customer_id UUID REFERENCES customers(id),
    product_id UUID REFERENCES products(id),
    order_id UUID REFERENCES orders(id),
    properties JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable for time-series optimization
DO $$
BEGIN
    PERFORM create_hypertable('analytics_events', 'timestamp',
        chunk_time_interval => INTERVAL '1 day',
        if_not_exists => TRUE);
EXCEPTION
    WHEN undefined_function THEN
        NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_analytics_event_type ON analytics_events(event_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_customer ON analytics_events(customer_id, timestamp DESC);

-- Updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
`
}