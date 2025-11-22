package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateNextJSApp creates a Next.js application with TypeScript and Tailwind
func CreateNextJSApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-app"
	}

	// Get typescript and tailwind settings
	typescript := args["typescript"] != "false"
	tailwind := args["tailwind"] != "false"

	// Create project directory
	projectPath := filepath.Join(".", name)

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Create directory structure
	dirs := []string{
		projectPath,
		filepath.Join(projectPath, "app"),
		filepath.Join(projectPath, "public"),
		filepath.Join(projectPath, "components"),
		filepath.Join(projectPath, "lib"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":    name,
		"version": "0.1.0",
		"private": true,
		"scripts": map[string]string{
			"dev":   "next dev",
			"build": "next build",
			"start": "next start",
			"lint":  "next lint",
		},
		"dependencies": map[string]string{
			"next":      "14.0.0",
			"react":     "^18.2.0",
			"react-dom": "^18.2.0",
		},
	}

	if typescript {
		packageJSON["devDependencies"] = map[string]string{
			"@types/node":      "^20.0.0",
			"@types/react":     "^18.2.0",
			"@types/react-dom": "^18.2.0",
			"typescript":       "^5.0.0",
		}
	}

	if tailwind {
		devDeps := packageJSON["devDependencies"].(map[string]string)
		devDeps["tailwindcss"] = "^3.3.0"
		devDeps["autoprefixer"] = "^10.0.1"
		devDeps["postcss"] = "^8"
	}

	packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(projectPath, "package.json"), packageData, 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create basic app/page.tsx or app/page.jsx
	ext := "jsx"
	if typescript {
		ext = "tsx"
	}

	pageContent := `export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <h1 className="text-4xl font-bold">Welcome to Next.js!</h1>
      <p className="mt-4 text-xl">Get started by editing app/page.` + ext + `</p>
    </main>
  )
}`

	if err := os.WriteFile(filepath.Join(projectPath, "app", fmt.Sprintf("page.%s", ext)), []byte(pageContent), 0644); err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	// Create app/layout.tsx or app/layout.jsx
	layoutContent := `export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}`

	if err := os.WriteFile(filepath.Join(projectPath, "app", fmt.Sprintf("layout.%s", ext)), []byte(layoutContent), 0644); err != nil {
		return fmt.Errorf("failed to create layout: %w", err)
	}

	// Create app/globals.css if using Tailwind
	if tailwind {
		globalsCSS := `@tailwind base;
@tailwind components;
@tailwind utilities;`

		if err := os.WriteFile(filepath.Join(projectPath, "app", "globals.css"), []byte(globalsCSS), 0644); err != nil {
			return fmt.Errorf("failed to create globals.css: %w", err)
		}

		// Update layout to import CSS
		layoutWithCSS := `import './globals.css'

` + layoutContent
		if err := os.WriteFile(filepath.Join(projectPath, "app", fmt.Sprintf("layout.%s", ext)), []byte(layoutWithCSS), 0644); err != nil {
			return fmt.Errorf("failed to update layout with CSS import: %w", err)
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
}`
		if err := os.WriteFile(filepath.Join(projectPath, "tailwind.config.js"), []byte(tailwindConfig), 0644); err != nil {
			return fmt.Errorf("failed to create tailwind.config.js: %w", err)
		}

		// Create postcss.config.js
		postcssConfig := `module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}`
		if err := os.WriteFile(filepath.Join(projectPath, "postcss.config.js"), []byte(postcssConfig), 0644); err != nil {
			return fmt.Errorf("failed to create postcss.config.js: %w", err)
		}
	}

	// Create next.config.js
	nextConfig := `/** @type {import('next').NextConfig} */
const nextConfig = {}

module.exports = nextConfig`

	if err := os.WriteFile(filepath.Join(projectPath, "next.config.js"), []byte(nextConfig), 0644); err != nil {
		return fmt.Errorf("failed to create next.config.js: %w", err)
	}

	// Create tsconfig.json if using TypeScript
	if typescript {
		tsConfig := `{
  "compilerOptions": {
    "target": "es5",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "node",
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
      "@/*": ["./*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}`
		if err := os.WriteFile(filepath.Join(projectPath, "tsconfig.json"), []byte(tsConfig), 0644); err != nil {
			return fmt.Errorf("failed to create tsconfig.json: %w", err)
		}
	}

	// Create .gitignore
	gitignore := `# See https://help.github.com/articles/ignoring-files/ for more about ignoring files.

# dependencies
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
next-env.d.ts`

	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

// CreateReactApp creates a React application using Vite
func CreateReactApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-react-app"
	}

	// For now, create a basic React structure
	// In production, this would use Vite or create-react-app
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
			"@types/react":          "^18.2.0",
			"@types/react-dom":      "^18.2.0",
			"@vitejs/plugin-react":  "^4.0.0",
			"vite":                  "^4.4.0",
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

	// Create src/App.jsx
	appJSX := `import { useState } from 'react'
import './App.css'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <h1>Vite + React</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.jsx</code> and save to test HMR
        </p>
      </div>
    </>
  )
}

export default App`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "App.jsx"), []byte(appJSX), 0644); err != nil {
		return fmt.Errorf("failed to create App.jsx: %w", err)
	}

	// Create basic CSS files
	indexCSS := `:root {
  font-family: Inter, system-ui, Avenir, Helvetica, Arial, sans-serif;
  line-height: 1.5;
  font-weight: 400;
}`

	appCSS := `#root {
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem;
  text-align: center;
}`

	if err := os.WriteFile(filepath.Join(projectPath, "src", "index.css"), []byte(indexCSS), 0644); err != nil {
		return fmt.Errorf("failed to create index.css: %w", err)
	}

	if err := os.WriteFile(filepath.Join(projectPath, "src", "App.css"), []byte(appCSS), 0644); err != nil {
		return fmt.Errorf("failed to create App.css: %w", err)
	}

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
		"name":    name,
		"version": "1.0.0",
		"description": "Express API",
		"main":    "src/index.js",
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

	return nil
}

// CreateNodeApp creates a basic Node.js application
func CreateNodeApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-node-app"
	}

	projectPath := filepath.Join(".", name)

	// Check if directory exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":    name,
		"version": "1.0.0",
		"description": "",
		"main":    "index.js",
		"scripts": map[string]string{
			"start": "node index.js",
			"dev":   "node index.js",
		},
		"dependencies": map[string]string{
			"dotenv": "^16.0.0",
		},
	}

	packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(projectPath, "package.json"), packageData, 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create index.js
	indexJS := `require('dotenv').config();

console.log('Hello from ` + name + `');

// Your code here`

	if err := os.WriteFile(filepath.Join(projectPath, "index.js"), []byte(indexJS), 0644); err != nil {
		return fmt.Errorf("failed to create index.js: %w", err)
	}

	return nil
}

// CreateAstroApp creates an Astro static site
func CreateAstroApp(ctx context.Context, args map[string]string) error {
	name := args["name"]
	if name == "" {
		name = "my-astro-site"
	}

	// For now, create a basic structure
	// In production, would use npm create astro@latest
	return CreateNodeApp(ctx, map[string]string{
		"name": name,
	})
}