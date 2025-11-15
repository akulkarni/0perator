package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	fmt.Fprintln(os.Stderr, "Test MCP server starting...")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)

	type TestInput struct {
		Message string `json:"message" jsonschema:"A test message"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TestInput) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Hello: " + input.Message},
			},
		}, nil, nil
	})

	fmt.Fprintln(os.Stderr, "Test MCP server ready")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
