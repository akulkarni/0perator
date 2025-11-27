#!/usr/bin/env bash
set -euo pipefail

# Directory where this script lives
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Module root is the parent directory of the script dir
MODDIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Package to build (relative to module root)
PKG="./cmd/0perator-mcp"

# Create a temporary directory for the compiled binary
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

# Build the binary in the module dir, output to temp location
go build -C "$MODDIR" -o "$TMPDIR/0perator-mcp" "$PKG"

# Run the compiled binary from your actual current working dir
"$TMPDIR/0perator-mcp" "$@"
