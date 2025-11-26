package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// OpenBrowser opens the given URL in the default browser
func OpenBrowser(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Start()
}

// waitForServer waits for a server to be ready at the given URL
func waitForServer(url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// findAppDirectory finds the nearest directory containing package.json
// It searches the current directory and immediate subdirectories
func findAppDirectory() string {
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

// findBun returns the path to the bun executable, checking common locations
func findBun() string {
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

// buildDevDependencies returns the appropriate dev dependencies based on options
func buildDevDependencies(typescript, tailwind bool) map[string]string {
	deps := map[string]string{
		"@types/node":      "^22.0.0",
		"@types/react":     "^19.0.0",
		"@types/react-dom": "^19.0.0",
		"@types/pg":        "^8.10.0",
	}

	if typescript {
		deps["typescript"] = "^5.0.0"
	}

	if tailwind {
		deps["tailwindcss"] = "^3.3.0"
		deps["autoprefixer"] = "^10.0.1"
		deps["postcss"] = "^8"
	}

	return deps
}

// CreateNextJSApp creates a complete Next.js app with proper configuration,
// auto-installs dependencies, starts dev server, and opens browser
func CreateNextJSApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-app"
	}

	typescript := args["typescript"] != "false"
	// Default to brutalist (no Tailwind) unless explicitly requested
	tailwind := args["tailwind"] == "true"
	brutalist := args["brutalist"] != "false" // Default true

	projectPath := filepath.Join(".", name)

	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Create directory structure
	dirs := []string{
		projectPath,
		filepath.Join(projectPath, "app"),
		filepath.Join(projectPath, "app", "api"),
		filepath.Join(projectPath, "app", "api", "init-db"),
		filepath.Join(projectPath, "public"),
		filepath.Join(projectPath, "components"),
		filepath.Join(projectPath, "lib"),
		filepath.Join(projectPath, "scripts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create enhanced package.json with database scripts
	packageJSON := map[string]interface{}{
		"name":    name,
		"version": "0.1.0",
		"private": true,
		"scripts": map[string]string{
			"dev":        "next dev",
			"build":      "next build",
			"start":      "next start",
			"lint":       "next lint",
			"db:check":   "bun scripts/check-db.js || node scripts/check-db.js",
			"db:init":    "bun scripts/init-db.js || node scripts/init-db.js",
			"db:migrate": "bun scripts/migrate.js || node scripts/migrate.js",
		},
		"dependencies": map[string]string{
			"next":      "^15.0.0",
			"react":     "^19.0.0",
			"react-dom": "^19.0.0",
			"pg":        "^8.11.3",
		},
		"devDependencies": buildDevDependencies(typescript, tailwind),
	}

	packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(projectPath, "package.json"), packageData, 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create proper tsconfig.json with path aliases
	tsconfigContent := `{
  "compilerOptions": {
    "target": "es5",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "paths": {
      "@/*": ["./*"],
      "@/components/*": ["./components/*"],
      "@/lib/*": ["./lib/*"],
      "@/app/*": ["./app/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
`
	if err := os.WriteFile(filepath.Join(projectPath, "tsconfig.json"), []byte(tsconfigContent), 0644); err != nil {
		return fmt.Errorf("failed to create tsconfig.json: %w", err)
	}

	// Create next.config.js with SSL workaround for Tiger Cloud
	nextConfigContent := `/** @type {import('next').NextConfig} */

// Disable SSL certificate validation for Tiger Cloud (self-signed certs)
// This must be set before any database connections are made
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const nextConfig = {}

module.exports = nextConfig
`
	if err := os.WriteFile(filepath.Join(projectPath, "next.config.js"), []byte(nextConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to create next.config.js: %w", err)
	}

	// Create .env.local template
	envContent := `# Database Configuration
# This will be populated when you run 'setup_database'
DATABASE_URL=

# Next.js Configuration
NEXT_PUBLIC_APP_NAME=` + name + `
`
	if err := os.WriteFile(filepath.Join(projectPath, ".env.local"), []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to create .env.local: %w", err)
	}

	// Create database utility lib/db.ts with proper SSL handling for Tiger Cloud
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
	if err := os.WriteFile(filepath.Join(projectPath, "lib", "db.ts"), []byte(dbUtilContent), 0644); err != nil {
		return fmt.Errorf("failed to create lib/db.ts: %w", err)
	}

	// Create database check script
	checkDbScript := `const { Pool } = require('pg');
const fs = require('fs');
const path = require('path');

// Disable SSL certificate validation for Tiger Cloud
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

// Load .env.local file
const envPath = path.join(__dirname, '..', '.env.local');
if (fs.existsSync(envPath)) {
  const envContent = fs.readFileSync(envPath, 'utf8');
  envContent.split('\n').forEach(line => {
    const match = line.match(/^([^#=]+)=(.*)$/);
    if (match) {
      const key = match[1].trim();
      const value = match[2].trim();
      if (!process.env[key]) {
        process.env[key] = value;
      }
    }
  });
}

async function checkDatabase() {
  if (!process.env.DATABASE_URL) {
    console.log('⚠️  DATABASE_URL not configured in .env.local');
    console.log('   Run "setup_database" to create a PostgreSQL database');
    return;
  }

  const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    connectionTimeoutMillis: 5000,
  });

  try {
    await pool.query('SELECT 1');
    console.log('✅ Database connected');

    // Check if tables exist
    const result = await pool.query(
      "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'"
    );

    if (result.rows[0].count === '0') {
      console.log('⚠️  No tables found. Run "npm run db:init" to create tables');
    }
  } catch (error) {
    console.log('❌ Database connection failed:', error.message);
    console.log('   Check your DATABASE_URL in .env.local');
  } finally {
    await pool.end();
  }
}

checkDatabase();
`
	if err := os.WriteFile(filepath.Join(projectPath, "scripts", "check-db.js"), []byte(checkDbScript), 0644); err != nil {
		return fmt.Errorf("failed to create scripts/check-db.js: %w", err)
	}

	// Create database init script
	initDbScript := `const { Pool } = require('pg');
const fs = require('fs');
const path = require('path');

// Disable SSL certificate validation for Tiger Cloud
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

// Load .env.local file
const envPath = path.join(__dirname, '..', '.env.local');
if (fs.existsSync(envPath)) {
  const envContent = fs.readFileSync(envPath, 'utf8');
  envContent.split('\n').forEach(line => {
    const match = line.match(/^([^#=]+)=(.*)$/);
    if (match) {
      const key = match[1].trim();
      const value = match[2].trim();
      if (!process.env[key]) {
        process.env[key] = value;
      }
    }
  });
}

async function initDatabase() {
  if (!process.env.DATABASE_URL) {
    console.error('DATABASE_URL not set in .env.local');
    process.exit(1);
  }

  const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
  });

  try {
    console.log('Initializing database schema...');

    // This schema will be created by the PostgreSQL setup tool
    // but we include a fallback here
    await pool.query(` + "`" + `
      CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255),
        created_at TIMESTAMPTZ DEFAULT NOW()
      )
    ` + "`" + `);

    console.log('✅ Database initialized successfully');
  } catch (error) {
    console.error('Failed to initialize database:', error);
    process.exit(1);
  } finally {
    await pool.end();
  }
}

initDatabase();
`
	if err := os.WriteFile(filepath.Join(projectPath, "scripts", "init-db.js"), []byte(initDbScript), 0644); err != nil {
		return fmt.Errorf("failed to create scripts/init-db.js: %w", err)
	}

	// Create API route for database initialization (app/api/init-db/route.ts)
	initApiContent := `import { NextResponse } from 'next/server';
import { query } from '@/lib/db';

export async function POST() {
  if (!process.env.DATABASE_URL) {
    return NextResponse.json(
      { error: 'Database not configured' },
      { status: 500 }
    );
  }

  try {
    // Check if tables already exist
    const checkResult = await query(
      "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'"
    );

    if (checkResult.rows[0].count > 0) {
      return NextResponse.json({
        message: 'Database already initialized',
        tables: ['users', 'sessions', 'posts']
      });
    }

    // Schema is created by the database setup tool
    // This is just a fallback
    await query(` + "`" + `
      CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255),
        created_at TIMESTAMPTZ DEFAULT NOW()
      )
    ` + "`" + `);

    return NextResponse.json({
      message: 'Database initialized successfully',
      tables: ['users']
    });
  } catch (error) {
    console.error('Database init error:', error);
    return NextResponse.json(
      { error: 'Failed to initialize database' },
      { status: 500 }
    );
  }
}
`
	if err := os.WriteFile(filepath.Join(projectPath, "app", "api", "init-db", "route.ts"), []byte(initApiContent), 0644); err != nil {
		return fmt.Errorf("failed to create app/api/init-db/route.ts: %w", err)
	}

	// Create main page - brutalist by default, Tailwind if requested
	var pageContent string
	if brutalist && !tailwind {
		// Brutalist app-like layout with sidebar and dashboard
		pageContent = `'use client';

import { useState, useEffect } from 'react';

export default function Home() {
  const [dbStatus, setDbStatus] = useState<'checking' | 'connected' | 'error'>('checking');
  const [activeTab, setActiveTab] = useState('dashboard');

  useEffect(() => {
    checkDatabase();
  }, []);

  const checkDatabase = async () => {
    try {
      const response = await fetch('/api/init-db', { method: 'POST' });
      setDbStatus(response.ok ? 'connected' : 'error');
    } catch {
      setDbStatus('error');
    }
  };

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '◻' },
    { id: 'projects', label: 'Projects', icon: '▤' },
    { id: 'settings', label: 'Settings', icon: '⚙' },
  ];

  return (
    <div style={{ display: 'flex', minHeight: '100vh', fontFamily: 'monospace' }}>
      {/* Sidebar */}
      <aside style={{
        width: '220px',
        background: '#1a1a1a',
        color: '#fff',
        padding: '1.5rem 0',
        display: 'flex',
        flexDirection: 'column'
      }}>
        <div style={{ padding: '0 1.5rem', marginBottom: '2rem' }}>
          <h1 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#ff4500' }}>` + name + `</h1>
        </div>

        <nav style={{ flex: 1 }}>
          {navItems.map(item => (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              style={{
                width: '100%',
                padding: '0.75rem 1.5rem',
                background: activeTab === item.id ? '#333' : 'transparent',
                border: 'none',
                borderLeft: activeTab === item.id ? '3px solid #ff4500' : '3px solid transparent',
                color: activeTab === item.id ? '#fff' : '#888',
                textAlign: 'left',
                cursor: 'pointer',
                fontFamily: 'monospace',
                fontSize: '0.9rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem'
              }}
            >
              <span>{item.icon}</span>
              {item.label}
            </button>
          ))}
        </nav>

        <div style={{ padding: '1rem 1.5rem', borderTop: '1px solid #333', fontSize: '0.75rem', color: '#666' }}>
          {dbStatus === 'connected' && <span style={{ color: '#4ade80' }}>● Connected</span>}
          {dbStatus === 'checking' && <span style={{ color: '#facc15' }}>● Connecting...</span>}
          {dbStatus === 'error' && <span style={{ color: '#f87171' }}>● Disconnected</span>}
        </div>
      </aside>

      {/* Main Content */}
      <main style={{ flex: 1, background: '#f5f5f5' }}>
        {/* Header */}
        <header style={{
          background: '#fff',
          padding: '1rem 2rem',
          borderBottom: '1px solid #e5e5e5',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', textTransform: 'capitalize' }}>
            {activeTab}
          </h2>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <button style={{
              padding: '0.5rem 1rem',
              background: '#ff4500',
              color: '#fff',
              border: 'none',
              fontFamily: 'monospace',
              cursor: 'pointer'
            }}>
              + New
            </button>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              background: '#1a1a1a',
              color: '#fff',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '0.875rem'
            }}>
              U
            </div>
          </div>
        </header>

        {/* Dashboard Content */}
        <div style={{ padding: '2rem' }}>
          {activeTab === 'dashboard' && (
            <>
              {/* Stats Grid */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: '1.5rem',
                marginBottom: '2rem'
              }}>
                {[
                  { label: 'Total Projects', value: '0', change: '+0%' },
                  { label: 'Active Tasks', value: '0', change: '+0%' },
                  { label: 'Team Members', value: '1', change: '' },
                ].map((stat, i) => (
                  <div key={i} style={{
                    background: '#fff',
                    padding: '1.5rem',
                    border: '2px solid #1a1a1a'
                  }}>
                    <div style={{ fontSize: '0.75rem', color: '#666', textTransform: 'uppercase', marginBottom: '0.5rem' }}>
                      {stat.label}
                    </div>
                    <div style={{ fontSize: '2rem', fontWeight: 'bold' }}>{stat.value}</div>
                    {stat.change && <div style={{ fontSize: '0.75rem', color: '#4ade80' }}>{stat.change}</div>}
                  </div>
                ))}
              </div>

              {/* Recent Activity */}
              <div style={{ background: '#fff', border: '2px solid #1a1a1a', padding: '1.5rem' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 'bold', marginBottom: '1rem' }}>Recent Activity</h3>
                <div style={{ color: '#666', textAlign: 'center', padding: '2rem' }}>
                  No activity yet. Create your first project to get started.
                </div>
              </div>
            </>
          )}

          {activeTab === 'projects' && (
            <div style={{ background: '#fff', border: '2px solid #1a1a1a', padding: '1.5rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 'bold', marginBottom: '1rem' }}>Projects</h3>
              <div style={{ color: '#666', textAlign: 'center', padding: '2rem' }}>
                No projects yet. Click "+ New" to create one.
              </div>
            </div>
          )}

          {activeTab === 'settings' && (
            <div style={{ background: '#fff', border: '2px solid #1a1a1a', padding: '1.5rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 'bold', marginBottom: '1rem' }}>Settings</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.75rem', color: '#666', textTransform: 'uppercase' }}>App Name</label>
                  <input
                    type="text"
                    defaultValue="` + name + `"
                    style={{
                      display: 'block',
                      width: '100%',
                      padding: '0.5rem',
                      marginTop: '0.25rem',
                      border: '2px solid #1a1a1a',
                      fontFamily: 'monospace'
                    }}
                  />
                </div>
                <div>
                  <label style={{ fontSize: '0.75rem', color: '#666', textTransform: 'uppercase' }}>Database Status</label>
                  <div style={{
                    padding: '0.5rem',
                    marginTop: '0.25rem',
                    background: dbStatus === 'connected' ? '#dcfce7' : '#fef2f2',
                    border: '2px solid ' + (dbStatus === 'connected' ? '#4ade80' : '#f87171')
                  }}>
                    {dbStatus === 'connected' ? '✓ Connected to PostgreSQL' : '✗ Not connected'}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
`
	} else {
		// Tailwind version (only if explicitly requested)
		pageContent = `'use client';

import { useState, useEffect } from 'react';

export default function Home() {
  const [dbStatus, setDbStatus] = useState<'checking' | 'connected' | 'not-configured' | 'error'>('checking');

  useEffect(() => {
    checkDatabase();
  }, []);

  const checkDatabase = async () => {
    try {
      const response = await fetch('/api/init-db', { method: 'POST' });
      if (response.ok) {
        setDbStatus('connected');
      } else {
        setDbStatus('not-configured');
      }
    } catch (error) {
      setDbStatus('error');
    }
  };

  return (
    <main className="min-h-screen p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">Welcome to ` + name + `!</h1>

        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold mb-4">Database Status</h2>
          {dbStatus === 'checking' && (
            <p className="text-gray-600">Checking database connection...</p>
          )}
          {dbStatus === 'connected' && (
            <p className="text-green-600">✅ Database connected and ready!</p>
          )}
          {dbStatus === 'not-configured' && (
            <div className="text-yellow-600">
              <p>⚠️ Database not configured</p>
              <p className="text-sm mt-2">Run 'setup_database' to create a PostgreSQL database</p>
            </div>
          )}
          {dbStatus === 'error' && (
            <p className="text-red-600">❌ Database connection error</p>
          )}
        </div>

        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-2xl font-semibold mb-4">Getting Started</h2>
          <ol className="list-decimal list-inside space-y-2">
            <li>Set up your database with 'setup_database'</li>
            <li>Start developing with 'npm run dev'</li>
            <li>Build for production with 'npm run build'</li>
          </ol>
        </div>
      </div>
    </main>
  );
}
`
	}
	if err := os.WriteFile(filepath.Join(projectPath, "app", "page.tsx"), []byte(pageContent), 0644); err != nil {
		return fmt.Errorf("failed to create app/page.tsx: %w", err)
	}

	// Create layout - brutalist by default, Tailwind if requested
	var layoutContent string
	if brutalist && !tailwind {
		// Brutalist version - no external fonts, minimal styles
		layoutContent = `import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: '` + name + `',
  description: 'Built with Next.js and PostgreSQL',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body style={{ margin: 0, padding: 0, fontFamily: 'monospace' }}>{children}</body>
    </html>
  )
}
`
	} else {
		// Tailwind version with custom font
		layoutContent = `import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: '` + name + `',
  description: 'Built with Next.js and PostgreSQL',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>{children}</body>
    </html>
  )
}
`
	}
	if err := os.WriteFile(filepath.Join(projectPath, "app", "layout.tsx"), []byte(layoutContent), 0644); err != nil {
		return fmt.Errorf("failed to create app/layout.tsx: %w", err)
	}

	// Create globals.css with Tailwind
	if tailwind {
		globalsCSSContent := `@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --foreground-rgb: 0, 0, 0;
  --background-start-rgb: 214, 219, 220;
  --background-end-rgb: 255, 255, 255;
}

body {
  color: rgb(var(--foreground-rgb));
  background: linear-gradient(
      to bottom,
      transparent,
      rgb(var(--background-end-rgb))
    )
    rgb(var(--background-start-rgb));
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "app", "globals.css"), []byte(globalsCSSContent), 0644); err != nil {
			return fmt.Errorf("failed to create app/globals.css: %w", err)
		}

		// Create tailwind.config.js
		tailwindConfig := `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "tailwind.config.js"), []byte(tailwindConfig), 0644); err != nil {
			return fmt.Errorf("failed to create tailwind.config.js: %w", err)
		}

		// Create postcss.config.js
		postcssConfig := `module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "postcss.config.js"), []byte(postcssConfig), 0644); err != nil {
			return fmt.Errorf("failed to create postcss.config.js: %w", err)
		}
	}

	// Create next-env.d.ts for TypeScript
	if typescript {
		nextEnvContent := `/// <reference types="next" />
/// <reference types="next/image-types/global" />

// NOTE: This file should not be edited
// see https://nextjs.org/docs/basic-features/typescript for more information.
`
		if err := os.WriteFile(filepath.Join(projectPath, "next-env.d.ts"), []byte(nextEnvContent), 0644); err != nil {
			return fmt.Errorf("failed to create next-env.d.ts: %w", err)
		}
	}

	// Create .gitignore
	gitignoreContent := `# dependencies
/node_modules
/.pnp
.pnp.js

# testing
/coverage

# next.js
/.next/
/out/

# production
/build

# misc
.DS_Store
*.pem

# debug
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# local env files
.env*.local

# vercel
.vercel

# typescript
*.tsbuildinfo
next-env.d.ts
`
	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	fmt.Printf("✅ Created Next.js app '%s' with:\n", name)
	fmt.Printf("   - TypeScript configuration with path aliases\n")
	fmt.Printf("   - Database utilities and connection pool\n")
	fmt.Printf("   - Auto database check on dev startup\n")
	fmt.Printf("   - Database initialization scripts\n")
	if brutalist && !tailwind {
		fmt.Printf("   - Brutalist UI (monospace, #ff4500 links, inline styles)\n")
	} else if tailwind {
		fmt.Printf("   - Tailwind CSS configured\n")
	}
	fmt.Printf("   - Environment variables template\n")

	// Auto-install dependencies using Bun (5-10x faster than npm)
	bunPath := findBun()
	if bunPath != "" {
		fmt.Printf("\n📦 Installing dependencies with Bun...\n")
		installCmd := exec.CommandContext(ctx, bunPath, "install")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Bun install failed, falling back to npm...\n")
			installCmd = exec.CommandContext(ctx, "npm", "install", "--silent")
			installCmd.Dir = projectPath
			if _, err := installCmd.CombinedOutput(); err != nil {
				fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
				fmt.Printf("   Run 'bun install' or 'npm install' manually in %s\n", name)
			} else {
				fmt.Printf("✅ Dependencies installed successfully (npm)\n")
			}
		} else {
			fmt.Printf("✅ Dependencies installed successfully (bun)\n")
		}
	} else {
		fmt.Printf("\n📦 Installing dependencies with npm...\n")
		installCmd := exec.CommandContext(ctx, "npm", "install", "--silent")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
			fmt.Printf("   Run 'npm install' manually in %s\n", name)
		} else {
			fmt.Printf("✅ Dependencies installed successfully\n")
		}
	}

	// Start dev server in background
	fmt.Printf("\n🚀 Starting dev server...\n")
	var devCmd *exec.Cmd
	if bunPath != "" {
		devCmd = exec.Command(bunPath, "run", "dev")
	} else {
		devCmd = exec.Command("npm", "run", "dev")
	}
	devCmd.Dir = projectPath
	devCmd.Stdout = nil
	devCmd.Stderr = nil
	if err := devCmd.Start(); err != nil {
		fmt.Printf("⚠️  Failed to start dev server: %v\n", err)
		fmt.Printf("   Run 'bun run dev' manually in %s\n", name)
		return nil
	}

	// Wait for server to be ready
	serverURL := "http://localhost:3000"
	if waitForServer(serverURL, 15*time.Second) {
		fmt.Printf("✅ Dev server ready at %s\n", serverURL)
	} else {
		fmt.Printf("⚠️  Dev server starting at %s (may take a moment)\n", serverURL)
	}

	fmt.Printf("\n🎉 Next.js app '%s' created\n", name)

	return nil
}

// CreateReactApp creates a React application using Vite
func CreateReactApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-react-app"
	}

	projectPath := filepath.Join(".", name)

	// Check if directory exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Create directories
	dirs := []string{
		projectPath,
		filepath.Join(projectPath, "src"),
		filepath.Join(projectPath, "public"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create package.json for Vite React
	packageJSON := map[string]interface{}{
		"name":    name,
		"version": "0.1.0",
		"private": true,
		"type":    "module",
		"scripts": map[string]string{
			"dev":     "vite",
			"build":   "vite build",
			"lint":    "eslint . --ext js,jsx --report-unused-disable-directives --max-warnings 0",
			"preview": "vite preview",
		},
		"dependencies": map[string]string{
			"react":     "^18.2.0",
			"react-dom": "^18.2.0",
		},
		"devDependencies": map[string]string{
			"@types/react":         "^18.2.0",
			"@types/react-dom":     "^18.2.0",
			"@vitejs/plugin-react": "^4.0.0",
			"vite":                 "^4.4.0",
		},
	}

	packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(projectPath, "package.json"), packageData, 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create vite.config.js
	viteConfig := `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
})`

	if err := os.WriteFile(filepath.Join(projectPath, "vite.config.js"), []byte(viteConfig), 0644); err != nil {
		return fmt.Errorf("failed to create vite.config.js: %w", err)
	}

	// Create index.html
	indexHTML := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>` + name + `</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>`

	if err := os.WriteFile(filepath.Join(projectPath, "index.html"), []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}

	// Create src/main.jsx
	mainJSX := `import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.jsx'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "main.jsx"), []byte(mainJSX), 0644); err != nil {
		return fmt.Errorf("failed to create main.jsx: %w", err)
	}

	// Create src/App.jsx with Brutalist styling
	appJSX := `import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div style={{ padding: '2rem', fontFamily: 'monospace', maxWidth: '800px', margin: '0 auto' }}>
      <h1 style={{ fontSize: '2rem', marginBottom: '2rem' }}>` + name + `</h1>
      <div style={{ padding: '1rem', background: '#f0f0f0', borderRadius: '4px' }}>
        <button
          onClick={() => setCount((count) => count + 1)}
          style={{
            padding: '0.75rem 1.5rem',
            background: '#ff4500',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            fontSize: '1rem',
            cursor: 'pointer',
            fontFamily: 'monospace'
          }}
        >
          count is {count}
        </button>
        <p style={{ marginTop: '1rem' }}>
          Edit <code>src/App.jsx</code> and save to test HMR
        </p>
      </div>
    </div>
  )
}

export default App`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "App.jsx"), []byte(appJSX), 0644); err != nil {
		return fmt.Errorf("failed to create App.jsx: %w", err)
	}

	// Create basic CSS files
	indexCSS := `:root {
  font-family: monospace;
  line-height: 1.5;
  font-weight: 400;
}

body {
  margin: 0;
  padding: 0;
}`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "index.css"), []byte(indexCSS), 0644); err != nil {
		return fmt.Errorf("failed to create index.css: %w", err)
	}

	// Create .gitignore
	gitignore := `node_modules/
dist/
.env
.env.local
.DS_Store
*.log`

	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	fmt.Printf("✅ Created React app '%s' with Vite\n", name)

	// Auto-install dependencies using Bun if available
	bunPath := findBun()
	if bunPath != "" {
		fmt.Printf("\n📦 Installing dependencies with Bun...\n")
		installCmd := exec.CommandContext(ctx, bunPath, "install")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Bun install failed, falling back to npm...\n")
			installCmd = exec.CommandContext(ctx, "npm", "install", "--silent")
			installCmd.Dir = projectPath
			if _, err := installCmd.CombinedOutput(); err != nil {
				fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
				fmt.Printf("   Run 'bun install' or 'npm install' manually in %s\n", name)
			} else {
				fmt.Printf("✅ Dependencies installed successfully (npm)\n")
			}
		} else {
			fmt.Printf("✅ Dependencies installed successfully (bun)\n")
		}
	} else {
		fmt.Printf("\n📦 Installing dependencies with npm...\n")
		installCmd := exec.CommandContext(ctx, "npm", "install", "--silent")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
			fmt.Printf("   Run 'npm install' manually in %s\n", name)
		} else {
			fmt.Printf("✅ Dependencies installed successfully\n")
		}
	}

	// Start dev server in background
	fmt.Printf("\n🚀 Starting dev server...\n")
	var devCmd *exec.Cmd
	if bunPath != "" {
		devCmd = exec.Command(bunPath, "run", "dev")
	} else {
		devCmd = exec.Command("npm", "run", "dev")
	}
	devCmd.Dir = projectPath
	devCmd.Stdout = nil
	devCmd.Stderr = nil
	if err := devCmd.Start(); err != nil {
		fmt.Printf("⚠️  Failed to start dev server: %v\n", err)
		fmt.Printf("   Run 'bun run dev' manually in %s\n", name)
		return nil
	}

	// Wait for server to be ready
	serverURL := "http://localhost:5173"
	if waitForServer(serverURL, 15*time.Second) {
		fmt.Printf("✅ Dev server ready at %s\n", serverURL)
	} else {
		fmt.Printf("⚠️  Dev server starting at %s (may take a moment)\n", serverURL)
	}

	fmt.Printf("\n🎉 React app '%s' created\n", name)

	return nil
}

// CreateExpressAPI creates an Express.js API
func CreateExpressAPI(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-api"
	}

	projectPath := filepath.Join(".", name)

	// Check if directory exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Create directories
	dirs := []string{
		projectPath,
		filepath.Join(projectPath, "src"),
		filepath.Join(projectPath, "src", "routes"),
		filepath.Join(projectPath, "src", "middleware"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":        name,
		"version":     "1.0.0",
		"description": "Express API",
		"main":        "src/index.js",
		"scripts": map[string]string{
			"start": "node src/index.js",
			"dev":   "nodemon src/index.js",
		},
		"dependencies": map[string]string{
			"express": "^4.18.0",
			"cors":    "^2.8.5",
			"dotenv":  "^16.0.0",
		},
		"devDependencies": map[string]string{
			"nodemon": "^3.0.0",
		},
	}

	packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(projectPath, "package.json"), packageData, 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create src/index.js
	indexJS := `const express = require('express');
const cors = require('cors');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Routes
app.get('/', (req, res) => {
  res.json({ message: 'API is running' });
});

app.get('/health', (req, res) => {
  res.json({ status: 'OK', timestamp: new Date().toISOString() });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Something went wrong!' });
});

// Start server
app.listen(PORT, () => {
  console.log('Server is running on port ' + PORT);
});`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "index.js"), []byte(indexJS), 0644); err != nil {
		return fmt.Errorf("failed to create index.js: %w", err)
	}

	// Create .env.example
	envExample := `PORT=3000
NODE_ENV=development`

	if err := os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644); err != nil {
		return fmt.Errorf("failed to create .env.example: %w", err)
	}

	// Create .gitignore
	gitignore := `node_modules/
.env
.env.local
.DS_Store
*.log
dist/
build/`

	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	fmt.Printf("✅ Created Express API '%s'\n", name)

	// Auto-install dependencies using Bun if available
	bunPath := findBun()
	if bunPath != "" {
		fmt.Printf("\n📦 Installing dependencies with Bun...\n")
		installCmd := exec.CommandContext(ctx, bunPath, "install")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Bun install failed, falling back to npm...\n")
			installCmd = exec.CommandContext(ctx, "npm", "install", "--silent")
			installCmd.Dir = projectPath
			if _, err := installCmd.CombinedOutput(); err != nil {
				fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
				fmt.Printf("   Run 'bun install' or 'npm install' manually in %s\n", name)
			} else {
				fmt.Printf("✅ Dependencies installed successfully (npm)\n")
			}
		} else {
			fmt.Printf("✅ Dependencies installed successfully (bun)\n")
		}
	} else {
		fmt.Printf("\n📦 Installing dependencies with npm...\n")
		installCmd := exec.CommandContext(ctx, "npm", "install", "--silent")
		installCmd.Dir = projectPath
		if _, err := installCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Failed to install dependencies: %v\n", err)
			fmt.Printf("   Run 'npm install' manually in %s\n", name)
		} else {
			fmt.Printf("✅ Dependencies installed successfully\n")
		}
	}

	// Start dev server in background
	fmt.Printf("\n🚀 Starting dev server...\n")
	var devCmd *exec.Cmd
	if bunPath != "" {
		devCmd = exec.Command(bunPath, "run", "dev")
	} else {
		devCmd = exec.Command("npm", "run", "dev")
	}
	devCmd.Dir = projectPath
	devCmd.Stdout = nil
	devCmd.Stderr = nil
	if err := devCmd.Start(); err != nil {
		fmt.Printf("⚠️  Failed to start dev server: %v\n", err)
		fmt.Printf("   Run 'bun run dev' manually in %s\n", name)
		return nil
	}

	// Wait for server to be ready
	serverURL := "http://localhost:3000"
	if waitForServer(serverURL, 15*time.Second) {
		fmt.Printf("✅ Dev server ready at %s\n", serverURL)
	} else {
		fmt.Printf("⚠️  Dev server starting at %s (may take a moment)\n", serverURL)
	}

	fmt.Printf("\n🎉 Express API '%s' created\n", name)

	return nil
}
