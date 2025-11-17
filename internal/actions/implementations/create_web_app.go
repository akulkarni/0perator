package implementations

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
)

// CreateWebAppAction returns the action definition for creating a web application
func CreateWebAppAction() *actions.Action {
	return &actions.Action{
		Name:         "create_web_app",
		Description:  "Create a new web application with specified framework",
		Category:     actions.CategoryCreate,
		Tags:         []string{"web", "frontend", "nextjs", "react", "vue", "svelte"},
		Tier:         actions.TierFast,
		EstimatedTime: 15 * time.Second,

		Inputs: []actions.Input{
			{
				Name:        "framework",
				Type:        actions.InputTypeString,
				Description: "Web framework to use",
				Required:    true,
				Options:     []string{"nextjs", "react", "vue", "svelte"},
				Default:     "nextjs",
			},
			{
				Name:        "directory",
				Type:        actions.InputTypeString,
				Description: "Directory to create the app in",
				Required:    false,
				Default:     ".",
			},
			{
				Name:        "name",
				Type:        actions.InputTypeString,
				Description: "Name of the application",
				Required:    false,
				Default:     "my-app",
			},
			{
				Name:        "typescript",
				Type:        actions.InputTypeBool,
				Description: "Use TypeScript",
				Required:    false,
				Default:     true,
			},
			{
				Name:        "styling",
				Type:        actions.InputTypeString,
				Description: "CSS framework to use",
				Required:    false,
				Options:     []string{"tailwind", "css", "scss", "styled-components"},
				Default:     "tailwind",
			},
		},

		Outputs: []actions.Output{
			{
				Name:        "project_path",
				Type:        actions.InputTypeString,
				Description: "Absolute path to the created project",
			},
			{
				Name:        "package_json_path",
				Type:        actions.InputTypeString,
				Description: "Path to package.json file",
			},
			{
				Name:        "framework",
				Type:        actions.InputTypeString,
				Description: "Framework used",
			},
			{
				Name:        "port",
				Type:        actions.InputTypeInt,
				Description: "Default development server port",
			},
		},

		Dependencies: []string{},
		Conflicts:    []string{}, // Could add conflicts with other create_* actions

		Implementation: createWebAppImpl,
	}
}

func createWebAppImpl(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	framework := inputs["framework"].(string)
	directory := inputs["directory"].(string)
	name := inputs["name"].(string)
	typescript := inputs["typescript"].(bool)
	styling := inputs["styling"].(string)

	// Create full path
	var projectPath string
	if directory == "." {
		projectPath = name
	} else {
		projectPath = filepath.Join(directory, name)
	}

	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory already exists
	if _, err := os.Stat(absPath); err == nil {
		return nil, fmt.Errorf("directory %s already exists", absPath)
	}

	// Execute based on framework
	var port int
	switch framework {
	case "nextjs":
		err = createNextApp(ctx, absPath, name, typescript, styling)
		port = 3000
	case "react":
		err = createReactApp(ctx, absPath, name, typescript)
		port = 3000
	case "vue":
		err = createVueApp(ctx, absPath, name, typescript)
		port = 5173
	case "svelte":
		err = createSvelteApp(ctx, absPath, name, typescript)
		port = 5173
	default:
		return nil, fmt.Errorf("unsupported framework: %s", framework)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s app: %w", framework, err)
	}

	return map[string]interface{}{
		"project_path":      absPath,
		"package_json_path": filepath.Join(absPath, "package.json"),
		"framework":         framework,
		"port":              port,
	}, nil
}

func createNextApp(ctx context.Context, path, name string, typescript bool, styling string) error {
	args := []string{
		"create-next-app@latest",
		path,
		"--use-pnpm",
		"--app", // Use App Router
	}

	if typescript {
		args = append(args, "--typescript")
	} else {
		args = append(args, "--javascript")
	}

	if styling == "tailwind" {
		args = append(args, "--tailwind")
	} else {
		args = append(args, "--no-tailwind")
	}

	// Add non-interactive flags
	args = append(args, "--eslint", "--import-alias", "@/*")

	cmd := exec.CommandContext(ctx, "pnpx", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("y\n") // Auto-confirm

	return cmd.Run()
}

func createReactApp(ctx context.Context, path, name string, typescript bool) error {
	// Use Vite for modern React apps (much faster than create-react-app)
	template := "react"
	if typescript {
		template = "react-ts"
	}

	cmd := exec.CommandContext(ctx, "pnpm", "create", "vite@latest", path, "--template", template)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("y\n")

	return cmd.Run()
}

func createVueApp(ctx context.Context, path, name string, typescript bool) error {
	template := "vue"
	if typescript {
		template = "vue-ts"
	}

	cmd := exec.CommandContext(ctx, "pnpm", "create", "vite@latest", path, "--template", template)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("y\n")

	return cmd.Run()
}

func createSvelteApp(ctx context.Context, path, name string, typescript bool) error {
	// First create with Vite
	cmd := exec.CommandContext(ctx, "pnpm", "create", "vite@latest", path, "--template", "svelte")
	if typescript {
		cmd = exec.CommandContext(ctx, "pnpm", "create", "vite@latest", path, "--template", "svelte-ts")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("y\n")

	return cmd.Run()
}