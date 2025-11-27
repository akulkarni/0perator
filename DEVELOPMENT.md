# Development Guide

## Dev Mode for MCP Server

When developing 0perator, you want code changes to take effect immediately without rebuilding and reinstalling the binary. The `--dev` flag enables this workflow.

### Usage

```bash
go run ./cmd/0perator-mcp init --dev
```

### What it does

Instead of configuring IDEs to use the compiled `0perator` binary:

```json
{
  "mcpServers": {
    "0perator": {
      "command": "/usr/local/bin/0perator"
    }
  }
}
```

Dev mode configures IDEs to use `go run`:

```json
{
  "mcpServers": {
    "0perator": {
      "command": "sh",
      "args": ["/path/to/your/repo/cmd/0perator-mcp/scripts/run-source.sh"]
    }
  }
}
```

### Why use dev mode

1. **Instant feedback** - Code changes take effect on the next MCP server restart (when you restart your IDE or reconnect)
2. **No rebuild step** - Skip the `go build && go install` cycle during development
3. **Easy debugging** - Add print statements or modify behavior without deployment friction

### Workflow

1. Run `0perator init --dev` from the repository root
2. Select which IDEs to configure
3. Make code changes to the MCP server
4. Restart your IDE (or reconnect to the MCP server) to pick up changes

### Switching back to production mode

Run `0perator init` without the `--dev` flag to reconfigure IDEs to use the compiled binary.
