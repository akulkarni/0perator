package cli

import (
	"fmt"
	"os"

	"github.com/akulkarni/0perator/internal/server"
)

// MCPStart starts the 0perator MCP server
func MCPStart() error {
	fmt.Fprintln(os.Stderr, "0perator MCP server starting...")

	// Create and start the MCP server
	srv := server.New()

	fmt.Fprintln(os.Stderr, "0perator MCP server ready")

	// Start serving (stdio mode)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	return nil
}
