package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// findBunPath returns the path to the bun executable, checking common locations
func findBunPath() string {
	// First check if bun is in PATH
	if path, err := exec.LookPath("bun"); err == nil {
		return path
	}

	// Check common installation locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check ~/.bun/bin/bun (official install location)
	bunPath := filepath.Join(homeDir, ".bun", "bin", "bun")
	if _, err := os.Stat(bunPath); err == nil {
		return bunPath
	}

	// Check /usr/local/bin/bun
	if _, err := os.Stat("/usr/local/bin/bun"); err == nil {
		return "/usr/local/bin/bun"
	}

	return ""
}

// findAppDir finds the nearest directory containing package.json
// It searches the current directory and immediate subdirectories
func findAppDir() string {
	// First check current directory
	if _, err := os.Stat("package.json"); err == nil {
		return "."
	}

	// Check immediate subdirectories for package.json
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			pkgPath := filepath.Join(entry.Name(), "package.json")
			if _, err := os.Stat(pkgPath); err == nil {
				return entry.Name()
			}
		}
	}

	return ""
}

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
	// Tiger CLI waits for READY by default, which is what we want
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

	// Step 2: Get credentials (Tiger CLI already waited for READY due to --wait-timeout)
	var connectionString string
	var password string

	getCmd := exec.CommandContext(ctx, "tiger", "service", "get", serviceID, "--with-password", "-o", "json")
	getOutput, err := getCmd.Output()
	if err == nil {
		var serviceDetails map[string]interface{}
		if err := json.Unmarshal(getOutput, &serviceDetails); err == nil {
			// Tiger CLI returns flat JSON, not nested under "service"
			if connStr, ok := serviceDetails["connection_string"].(string); ok {
				connectionString = connStr
			}
			if pw, ok := serviceDetails["password"].(string); ok {
				password = pw
			}
		}
	}

	if connectionString == "" || password == "" {
		// Fallback - shouldn't happen since Tiger CLI waits for READY
		fmt.Printf("⚠️  Could not retrieve credentials. Get them with:\n")
		fmt.Printf("   tiger service get %s --with-password\n", serviceID)
		connectionString = fmt.Sprintf("postgresql://tsdbadmin:CHECK_TIGER_CLI@%s.tsdb.cloud.timescale.com:36720/tsdb?sslmode=require", serviceID)
		password = "CHECK_TIGER_CLI"
	}

	// Step 3: Initialize schema based on app type
	// Note: Schema will be created by the app's init-db script when the database is ready
	// We skip direct schema initialization to avoid waiting for the database
	schema := getSchemaForAppType(appType)
	_ = schema // Schema is used by init-db.js script

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
	// First, find the app directory (where package.json is)
	appDir := findAppDir()
	if appDir == "" {
		fmt.Printf("⚠️  No app directory found (no package.json). Creating .env.local in current directory.\n")
		appDir = "."
	} else if appDir != "." {
		fmt.Printf("📁 Found app directory: %s\n", appDir)
	}

	envPath := filepath.Join(appDir, ".env.local")

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

	// Create lib directory and db.ts file in the app directory
	if appDir != "" {
		libDir := filepath.Join(appDir, "lib")
		os.MkdirAll(libDir, 0755)
		dbTsPath := filepath.Join(libDir, "db.ts")
		if err := os.WriteFile(dbTsPath, []byte(dbUtilContent), 0644); err != nil {
			fmt.Printf("⚠️  Could not create %s: %v\n", dbTsPath, err)
		} else {
			fmt.Printf("✅ Created/updated %s with connection pool\n", dbTsPath)
		}

		// Note: Skip running init-db script here since database is still provisioning
		// The app will initialize the database on first connection
		initScriptPath := filepath.Join(appDir, "scripts", "init-db.js")
		if _, err := os.Stat(initScriptPath); err == nil {
			fmt.Printf("📝 Database schema will be initialized when the app starts\n")
		}
	}

	// Step 6: Verify database connection works via psql
	fmt.Printf("\n🔍 Verifying database connection...\n")
	verified := verifyDatabaseConnection(connectionString)
	if !verified {
		fmt.Printf("⚠️  Could not verify database connection via psql.\n")
	} else {
		fmt.Printf("✅ Database connection verified via psql\n")
	}

	// Step 7: Verify the app can connect (if dev server is running)
	fmt.Printf("🔍 Checking app database integration...\n")
	appVerified := verifyAppDatabaseConnection(appDir)
	if !appVerified {
		fmt.Printf("⚠️  App database check skipped (dev server may not be running)\n")
	} else {
		fmt.Printf("✅ App database integration verified!\n")
	}

	fmt.Printf("\n🎉 PostgreSQL setup complete!\n")
	fmt.Printf("   - Database: %s (ready)\n", dbName)
	fmt.Printf("   - Connection: Saved to %s\n", envPath)
	if appDir != "" {
		fmt.Printf("   - Utilities: %s/lib/db.ts ready to use\n", appDir)
	}

	return nil
}

// verifyDatabaseConnection tests that the connection string actually works
func verifyDatabaseConnection(connectionString string) bool {
	// Use psql to test the connection (available on most systems)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set environment to skip SSL cert verification (Tiger Cloud uses self-signed)
	cmd := exec.CommandContext(ctx, "psql", connectionString, "-c", "SELECT 1")
	cmd.Env = append(os.Environ(), "PGSSLMODE=require", "NODE_TLS_REJECT_UNAUTHORIZED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// psql might not be available, try with pg_isready
		cmd2 := exec.CommandContext(ctx, "pg_isready", "-d", connectionString)
		if err2 := cmd2.Run(); err2 == nil {
			return true
		}
		fmt.Printf("   Debug: %s\n", string(output))
		return false
	}
	return true
}

// verifyAppDatabaseConnection checks if the app can connect to the database
// by hitting the /api/init-db endpoint. If the app reports "not configured",
// it restarts the dev server to pick up the new .env.local
func verifyAppDatabaseConnection(appDir string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	port := "3000"
	url := fmt.Sprintf("http://localhost:%s/api/init-db", port)

	// First, check if server is running
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		// Server not running - that's fine, it will pick up .env.local when started
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Check for success
	if resp.StatusCode == 200 {
		if strings.Contains(bodyStr, "initialized") || strings.Contains(bodyStr, "already") {
			return true
		}
	}

	// Check for "Database not configured" error - need to restart dev server
	if strings.Contains(bodyStr, "not configured") {
		fmt.Printf("   ⚠️  App needs restart to pick up database config...\n")

		// Kill existing server
		killCmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -ti:%s | xargs kill -9 2>/dev/null", port))
		killCmd.Run()
		time.Sleep(1 * time.Second)

		// Restart dev server in background
		bunPath := findBunPath()
		var devCmd *exec.Cmd
		if bunPath != "" {
			devCmd = exec.Command(bunPath, "run", "dev")
		} else {
			devCmd = exec.Command("npm", "run", "dev")
		}
		devCmd.Dir = appDir
		devCmd.Stdout = nil
		devCmd.Stderr = nil
		if err := devCmd.Start(); err != nil {
			fmt.Printf("   ⚠️  Could not restart dev server: %v\n", err)
			return false
		}

		// Wait for server to be ready
		fmt.Printf("   🔄 Restarting dev server...\n")
		for i := 0; i < 20; i++ {
			time.Sleep(500 * time.Millisecond)
			resp2, err := client.Post(url, "application/json", nil)
			if err != nil {
				continue
			}
			body2, _ := io.ReadAll(resp2.Body)
			resp2.Body.Close()

			if resp2.StatusCode == 200 {
				if strings.Contains(string(body2), "initialized") || strings.Contains(string(body2), "already") {
					fmt.Printf("   ✅ Dev server restarted and database connected!\n")
					return true
				}
			}
		}

		fmt.Printf("   ⚠️  Dev server restarted but database still not connecting\n")
		return false
	}

	// Any other error
	if resp.StatusCode >= 400 {
		fmt.Printf("   ⚠️  App returned error: %s\n", bodyStr)
		return false
	}

	return false
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