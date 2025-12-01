#!/bin/bash
# Run the MCP server with the MCP Inspector for development/debugging
# Opens a web UI at http://127.0.0.1:6274 with stdio transport

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Disable auto-open so we can add transport=stdio to the URL
export MCP_AUTO_OPEN_ENABLED=false

# Create a temp file for output
TMPFILE=$(mktemp)
trap "rm -f $TMPFILE" EXIT

# Start inspector and tee output
npx -y @modelcontextprotocol/inspector --config mcp-inspector.json --server 0perator 2>&1 | tee "$TMPFILE" &
INSPECTOR_PID=$!

# Wait for token URL to appear in output
for i in {1..10}; do
  sleep 1
  if grep -q "MCP_PROXY_AUTH_TOKEN=" "$TMPFILE" 2>/dev/null; then
    # Extract token
    TOKEN=$(grep -o 'MCP_PROXY_AUTH_TOKEN=[^&"]*' "$TMPFILE" | head -1 | cut -d= -f2)

    # Wait for server to be ready
    for j in {1..10}; do
      if curl -s -o /dev/null -w "%{http_code}" "http://localhost:6274" 2>/dev/null | grep -q "200\|304"; then
        open "http://localhost:6274/?transport=stdio&MCP_PROXY_AUTH_TOKEN=$TOKEN"
        break 2
      fi
      sleep 0.5
    done
    break
  fi
done

# Keep running until interrupted
wait $INSPECTOR_PID
