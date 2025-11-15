# 0perator

Infrastructure for AI-native development - build and deploy full-stack applications through natural conversation.

## Overview

0perator is an MCP (Model Context Protocol) server that enables AI agents in Claude Code, Cursor, Windsurf, and other IDEs to build, deploy, and manage real applications through prompt templates and execution primitives.

**Key Innovation:** Instead of hardcoded scaffolding, 0perator uses comprehensive prompt templates that guide LLMs to build applications step-by-step with best practices baked in.

## Installation

```bash
# Install 0perator CLI
curl -fsSL https://cli.0p.dev | sh

# Initialize and configure MCP
0perator init
```

This will:
1. Install tiger-cli (for database provisioning)
2. Authenticate with Tiger Cloud
3. Configure MCP servers in your IDE (Claude Code, Cursor, or Windsurf)

## Quick Start

Once installed, you can say things like:

```
"Build me a trading card marketplace"
"Create a blog with user authentication"
"Deploy my app to Cloudflare"
```

The AI will:
1. Discover relevant templates (web app, database, auth, payments, deployment)
2. Read template guides with best practices
3. Execute operations to build your app
4. Deploy and run it locally or in production

## Architecture

```
┌─────────────────────────────────────────┐
│  Claude Code / Cursor / Windsurf        │
│  (MCP Client)                           │
└────────────────┬────────────────────────┘
                 │ MCP Protocol
                 │
┌────────────────▼────────────────────────┐
│  0perator MCP Server (Go)               │
│                                          │
│  ┌──────────────────────────────────┐   │
│  │ 3 MCP Tools:                     │   │
│  │                                  │   │
│  │ 1. discover_patterns             │   │
│  │    Search templates by tags      │   │
│  │                                  │   │
│  │ 2. get_template                  │   │
│  │    Retrieve template content     │   │
│  │                                  │   │
│  │ 3. execute                       │   │
│  │    8 primitives:                 │   │
│  │    - run_command                 │   │
│  │    - read_file                   │   │
│  │    - create_file                 │   │
│  │    - edit_file                   │   │
│  │    - start_process               │   │
│  │    - stop_process                │   │
│  │    - get_logs                    │   │
│  │    - list_processes              │   │
│  └──────────────────────────────────┘   │
│                                          │
│  ┌──────────────────────────────────┐   │
│  │ Prompt Template System:          │   │
│  │ - Tag-based discovery            │   │
│  │ - Semantic search                │   │
│  │ - Category defaults              │   │
│  │ - Template composition           │   │
│  └──────────────────────────────────┘   │
│                                          │
│  ┌──────────────────────────────────┐   │
│  │ Process Management:              │   │
│  │ - Local deployment               │   │
│  │ - Log streaming                  │   │
│  │ - Port allocation                │   │
│  │ - Health checking                │   │
│  └──────────────────────────────────┘   │
└──────────────────────────────────────────┘
                 │
                 │ Shells out to
                 │
┌────────────────▼────────────────────────┐
│  External Tools:                        │
│  - Tiger MCP (database provisioning)    │
│  - npm/node (dependencies)              │
│  - Deployment CLIs (Vercel, CF, etc.)   │
└─────────────────────────────────────────┘
```

## How It Works

### 1. User Request
```
User: "Create a blog with authentication"
```

### 2. AI Uses discover_patterns
```json
{
  "tool": "discover_patterns",
  "query": "blog authentication"
}
```

Returns matching templates:
- `create_web_app` - Build web application
- `database_tiger` - Add PostgreSQL database
- `auth_jwt` - Add JWT authentication

### 3. AI Uses get_template
```json
{
  "tool": "get_template",
  "name": "create_web_app"
}
```

Returns comprehensive guide with:
- Architecture overview
- Complete code examples
- Step-by-step instructions
- Best practices

### 4. AI Uses execute
```json
{
  "tool": "execute",
  "operation": "create_file",
  "params": {
    "path": "blog/package.json",
    "content": "{ ... }"
  }
}
```

Executes operations to build the app.

## Available Templates

### v0 (Current)
- ✅ **create_web_app** - Node.js + TypeScript + Fastify web applications
- 🚧 **database_tiger** - PostgreSQL/TimescaleDB with Tiger Cloud
- 🚧 **auth_jwt** - JWT authentication
- 🚧 **payments_stripe** - Stripe payment integration
- 🚧 **deploy_cloudflare** - Cloudflare Pages deployment

### Future Templates
- API-only backends
- CLI tools
- Real-time features (WebSockets)
- Email integration
- File storage
- Search functionality
- Testing setup
- CI/CD pipelines

## Project Structure

```
0perator/
├── cmd/
│   └── 0perator-mcp/          # MCP server entry point
├── internal/
│   ├── server/                # MCP server implementation
│   │   ├── server.go          # Server setup
│   │   ├── tools.go           # Tool definitions
│   │   └── execute.go         # Execute primitives
│   ├── prompts/               # Template system
│   │   ├── types.go           # Template types
│   │   ├── loader.go          # Template loading
│   │   ├── discovery.go       # Tag-based search
│   │   ├── defaults.go        # Category defaults
│   │   └── md/                # Template files
│   │       ├── create_web_app.md
│   │       ├── database_tiger.md
│   │       └── ...
│   ├── runtime/               # Process management
│   │   └── process.go         # Local deployment
│   ├── cli/                   # CLI commands
│   │   ├── init.go            # Setup wizard
│   │   └── uninstall.go       # Cleanup
│   └── mcp/                   # MCP utilities
├── scripts/
│   ├── build.sh               # Multi-platform builds
│   └── install.sh             # Installation script
├── go.mod
└── README.md
```

## Development

### Build

```bash
# Build for current platform
go build -o bin/0perator ./cmd/0perator-mcp

# Build for all platforms
./scripts/build.sh
```

### Test Locally

```bash
# Install locally
cp bin/0perator ~/.local/bin/

# Run init
0perator init

# Test MCP server directly
0perator mcp start
```

### Run in Claude Code

Once installed with `0perator init`, the MCP server will automatically start when you open Claude Code. You can test by asking:

```
"Show me available templates"
"Create a simple web app"
```

## Template Development

Templates are markdown files with YAML frontmatter:

```markdown
---
title: My Template
description: What this template does
tags: [web, nodejs, api]
category: foundational
dependencies: []
related: [other_template]
---

# Template Content

Step-by-step guide with code examples...
```

**Template Guidelines:**
- Comprehensive: Include full working code examples
- Execute-friendly: Show actual `execute` operations
- Composable: Reference other templates
- Best practices: Guide LLM to write quality code

## Why Prompt Templates?

### Traditional Approach (Scaffolding)
```
create_app(name, template) → Generated boilerplate
```

**Problems:**
- Fixed structure, limited flexibility
- Hard to customize
- Can't adapt to user needs
- Requires 100s of templates for variations

### 0perator Approach (Prompts)
```
get_template(name) → Comprehensive guide
execute(operations) → Custom implementation
```

**Benefits:**
- LLM adapts to user requirements
- Best practices built into guidance
- Infinite flexibility with finite templates
- Easy to add new patterns

## v0 Goals

- ✅ 3-tool MCP architecture
- ✅ Prompt template system
- ✅ Tag-based discovery
- ✅ 8 execution primitives
- ✅ Process management
- 🚧 5 foundational templates
- 🚧 End-to-end testing

## Success Metrics

- **Speed**: Full-stack app deployed locally in under 2 minutes
- **Quality**: Production-ready code with TypeScript, validation, error handling
- **AI-Native**: Works through natural conversation, no YAML or config files
- **Composable**: Templates build on each other (web → db → auth → payments)
- **Portable**: Works in any MCP-compatible IDE

## License

TBD
