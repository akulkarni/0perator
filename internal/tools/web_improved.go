package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateNextJSAppImproved creates a complete Next.js app with proper configuration
func CreateNextJSAppImproved(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-app"
	}

	typescript := args["typescript"] != "false"
	tailwind := args["tailwind"] != "false"

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
			"dev":        "npm run db:check && next dev",
			"build":      "next build",
			"start":      "next start",
			"lint":       "next lint",
			"db:check":   "node scripts/check-db.js",
			"db:init":    "node scripts/init-db.js",
			"db:migrate": "node scripts/migrate.js",
		},
		"dependencies": map[string]string{
			"next":      "14.0.0",
			"react":     "^18.2.0",
			"react-dom": "^18.2.0",
			"pg":        "^8.11.3",
		},
		"devDependencies": map[string]string{
			"@types/node":      "^20.0.0",
			"@types/react":     "^18.2.0",
			"@types/react-dom": "^18.2.0",
			"@types/pg":        "^8.10.0",
			"typescript":       "^5.0.0",
			"tailwindcss":      "^3.3.0",
			"autoprefixer":     "^10.0.1",
			"postcss":          "^8",
		},
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

	// Create next.config.js
	nextConfigContent := `/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    serverActions: true,
  },
}

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
NEXT_PUBLIC_APP_NAME=${name}
`
	if err := os.WriteFile(filepath.Join(projectPath, ".env.local"), []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to create .env.local: %w", err)
	}

	// Create database utility lib/db.ts
	dbUtilContent := `import { Pool } from 'pg';

let pool: Pool | undefined;

if (process.env.DATABASE_URL) {
  pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 20,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
  });

  pool.on('error', (err) => {
    console.error('Unexpected database error:', err);
  });
}

export default pool;

export async function query(text: string, params?: any[]) {
  if (!pool) {
    throw new Error('Database not configured. Please set DATABASE_URL in .env.local');
  }
  const result = await pool.query(text, params);
  return result;
}

export async function getClient() {
  if (!pool) {
    throw new Error('Database not configured. Please set DATABASE_URL in .env.local');
  }
  return await pool.connect();
}
`
	if err := os.WriteFile(filepath.Join(projectPath, "lib", "db.ts"), []byte(dbUtilContent), 0644); err != nil {
		return fmt.Errorf("failed to create lib/db.ts: %w", err)
	}

	// Create database check script
	checkDbScript := `const { Pool } = require('pg');

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
import pool from '@/lib/db';

export async function POST() {
  if (!pool) {
    return NextResponse.json(
      { error: 'Database not configured' },
      { status: 500 }
    );
  }

  try {
    // Check if tables already exist
    const checkResult = await pool.query(
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
    await pool.query(` + "`" + `
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

	// Create main page with proper TypeScript
	pageContent := `'use client';

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
        <h1 className="text-4xl font-bold mb-8">Welcome to ${name}!</h1>

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
	if err := os.WriteFile(filepath.Join(projectPath, "app", "page.tsx"), []byte(pageContent), 0644); err != nil {
		return fmt.Errorf("failed to create app/page.tsx: %w", err)
	}

	// Create layout with Tailwind
	layoutContent := `import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: '${name}',
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
	fmt.Printf("   - Tailwind CSS configured\n")
	fmt.Printf("   - Environment variables template\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. cd %s\n", name)
	fmt.Printf("2. npm install\n")
	fmt.Printf("3. Run 'setup_database' to create PostgreSQL\n")
	fmt.Printf("4. npm run dev\n")

	return nil
}

// Similar improvements for React and Express...
func CreateReactAppImproved(ctx context.Context, args map[string]string) error {
	// Similar implementation with proper Vite config, tsconfig, etc.
	return CreateReactApp(ctx, args) // For now, use existing
}

func CreateExpressAPIImproved(ctx context.Context, args map[string]string) error {
	// Similar implementation with database setup, middleware, etc.
	return CreateExpressAPI(ctx, args) // For now, use existing
}