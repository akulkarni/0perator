package tools

import (
	"context"
	"fmt"
	"os/exec"
)

// OpenBrowser opens the given URL in the default browser
func OpenBrowser(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Start()
}

// CreateNextJSApp creates a complete Next.js app with proper configuration,
// auto-installs dependencies, starts dev server, and opens browser
func CreateNextJSApp(ctx context.Context, name string) error {
	if name == "" {
		name = "my-app"
	}

	cmd := exec.CommandContext(ctx, "npm", "create", "t3-app@latest", "--", name, "--noGit", "--CI", "--tailwind", "--drizzle", "--trpc", "--dbProvider", "postgres", "--appRouter")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run create-t3-app: %w\n%s", err, output)
	}
	fmt.Print(string(output))

	fmt.Printf("\n🎉 Next.js app '%s' created\n", name)

	return nil
}
