package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/akulkarni/0perator/internal/templates"
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

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	// Create package.json with database scripts (dynamic, kept in code)
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

	// Template data for variable substitution
	data := templates.TemplateData{Name: name}

	// Copy base templates (shared across all variants)
	if err := templates.WriteDir("web/nextjs/base", projectPath, data); err != nil {
		return fmt.Errorf("failed to write base templates: %w", err)
	}

	// Copy variant templates (brutalist or tailwind)
	if brutalist && !tailwind {
		if err := templates.WriteDir("web/nextjs/brutalist", projectPath, data); err != nil {
			return fmt.Errorf("failed to write brutalist templates: %w", err)
		}
	} else if tailwind {
		if err := templates.WriteDir("web/nextjs/tailwind", projectPath, data); err != nil {
			return fmt.Errorf("failed to write tailwind templates: %w", err)
		}
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

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	// Create package.json for Vite React (dynamic, kept in code)
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

	// Template data for variable substitution
	data := templates.TemplateData{Name: name}

	// Copy base templates
	if err := templates.WriteDir("web/react/base", projectPath, data); err != nil {
		return fmt.Errorf("failed to write base templates: %w", err)
	}

	// Copy brutalist variant (default for React)
	if err := templates.WriteDir("web/react/brutalist", projectPath, data); err != nil {
		return fmt.Errorf("failed to write brutalist templates: %w", err)
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

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	// Create package.json (dynamic, kept in code)
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

	// Template data for variable substitution
	data := templates.TemplateData{Name: name}

	// Copy base templates
	if err := templates.WriteDir("web/express/base", projectPath, data); err != nil {
		return fmt.Errorf("failed to write base templates: %w", err)
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
