package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
	"github.com/akulkarni/0perator/internal/operator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  test-actions list                    - List all available actions")
		fmt.Println("  test-actions discover <query>         - Search for actions")
		fmt.Println("  test-actions create-web               - Create a test web app")
		fmt.Println("  test-actions setup-db                 - Setup a test database")
		fmt.Println("  test-actions full-stack              - Create web app + database")
		os.Exit(1)
	}

	op := operator.New()
	ctx := context.Background()

	switch os.Args[1] {
	case "list":
		listActions(op)

	case "discover":
		if len(os.Args) < 3 {
			fmt.Println("Usage: test-actions discover <query>")
			os.Exit(1)
		}
		discoverActions(op, os.Args[2])

	case "create-web":
		createWebApp(ctx, op)

	case "setup-db":
		setupDatabase(ctx, op)

	case "full-stack":
		createFullStack(ctx, op)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func listActions(op *operator.Operator) {
	fmt.Println("📦 Available Actions:")
	fmt.Println("=" + string(make([]byte, 60)))

	actions := op.GetAvailableActions("")
	for _, action := range actions {
		fmt.Printf("• %s (%s)\n", action.Name, action.Category)
		fmt.Printf("  %s\n", action.Description)
		if len(action.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", action.Tags)
		}
		fmt.Println()
	}
}

func discoverActions(op *operator.Operator, query string) {
	fmt.Printf("🔍 Searching for '%s':\n", query)
	fmt.Println("=" + string(make([]byte, 60)))

	actions := op.DiscoverActions(query)
	if len(actions) == 0 {
		fmt.Println("No actions found matching your query.")
		return
	}

	for _, action := range actions {
		fmt.Printf("• %s\n", action.Name)
		fmt.Printf("  %s\n", action.Description)
		fmt.Println()
	}
}

func createWebApp(ctx context.Context, op *operator.Operator) {
	fmt.Println("🚀 Creating Web Application")
	fmt.Println("=" + string(make([]byte, 60)))

	result, err := op.ExecuteAction(ctx, "create_web_app", map[string]interface{}{
		"framework":  "nextjs",
		"name":       "test-app",
		"directory":  "/tmp",
		"typescript": true,
		"styling":    "tailwind",
	})

	if err != nil {
		log.Fatalf("Failed to execute action: %v", err)
	}

	if !result.Success {
		log.Fatalf("Action failed: %s", result.Error)
	}

	fmt.Printf("✅ Web app created successfully!\n")
	fmt.Printf("   Path: %s\n", result.Outputs["project_path"])
	fmt.Printf("   Framework: %s\n", result.Outputs["framework"])
	fmt.Printf("   Port: %v\n", result.Outputs["port"])
	fmt.Printf("   Time: %v\n", result.Duration)
}

func setupDatabase(ctx context.Context, op *operator.Operator) {
	fmt.Println("🗄️  Setting up PostgreSQL Database")
	fmt.Println("=" + string(make([]byte, 60)))

	result, err := op.ExecuteAction(ctx, "setup_postgres", map[string]interface{}{
		"database_name":  "test_db",
		"port":           5433,
		"with_timescale": true,
		"docker":         true,
	})

	if err != nil {
		log.Fatalf("Failed to execute action: %v", err)
	}

	if !result.Success {
		log.Fatalf("Action failed: %s", result.Error)
	}

	fmt.Printf("✅ Database setup complete!\n")
	fmt.Printf("   Connection: %s\n", result.Outputs["connection_string"])
	fmt.Printf("   Container: %s\n", result.Outputs["container_name"])
	fmt.Printf("   Time: %v\n", result.Duration)
}

func createFullStack(ctx context.Context, op *operator.Operator) {
	fmt.Println("🎯 Creating Full Stack Application")
	fmt.Println("=" + string(make([]byte, 60)))

	// Define the sequence of actions
	sequence := []actions.ActionCall{
		{
			Action: "create_web_app",
			Inputs: map[string]interface{}{
				"framework":  "nextjs",
				"name":       "fullstack-app",
				"directory":  "/tmp",
				"typescript": true,
				"styling":    "tailwind",
			},
		},
		{
			Action: "setup_postgres",
			Inputs: map[string]interface{}{
				"database_name":  "fullstack_db",
				"port":           5434,
				"with_timescale": false,
				"docker":         true,
			},
		},
	}

	startTime := time.Now()
	result, err := op.ExecuteSequence(ctx, sequence)
	if err != nil {
		log.Fatalf("Failed to execute sequence: %v", err)
	}

	if !result.Success {
		for _, action := range result.Actions {
			if !action.Success {
				log.Fatalf("Action %s failed: %s", action.Action, action.Error)
			}
		}
	}

	fmt.Printf("\n✅ Full stack created successfully!\n")
	fmt.Printf("   Total Time: %v\n\n", time.Since(startTime))

	fmt.Println("📊 Action Results:")
	for _, action := range result.Actions {
		emoji := "✅"
		if !action.Success {
			emoji = "❌"
		}
		fmt.Printf("   %s %s (%v)\n", emoji, action.Action, action.Duration)
	}

	fmt.Printf("\n🔗 Outputs:\n")
	fmt.Printf("   App Path: %s\n", result.Outputs["project_path"])
	fmt.Printf("   Database: %s\n", result.Outputs["connection_string"])
}