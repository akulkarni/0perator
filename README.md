# 0perator

Infrastructure for AI agents to deploy software instantly.

## Overview

0perator is a backend service (infrastructure + Postgres) built for AI agents. It exposes simple MCP actions that allow agents in Claude Code, Cursor, and other IDEs to create, deploy, and manage real applications instantly.

## Installation

```bash
# Install 0perator CLI
curl -fsSL https://cli.0p.dev | sh

# Initialize and configure MCP
0perator init
```

## Architecture

```
┌─────────────────────────────────────────┐
│  Claude Code / IDE                      │
│  (MCP Client)                           │
└────────────────┬────────────────────────┘
                 │ MCP Protocol
                 │
┌────────────────▼────────────────────────┐
│  0perator-mcp (Go)                      │
│  ┌──────────────────────────────────┐   │
│  │ MCP Tools:                       │   │
│  │ - create_app                     │   │
│  │ - create_db                      │   │
│  │ - deploy_local                   │   │
│  └──────────────────────────────────┘   │
│                                          │
│  ┌──────────────────────────────────┐   │
│  │ Orchestration Logic:             │   │
│  │ - Template scaffolding           │   │
│  │ - Process management             │   │
│  │ - Connection string injection    │   │
│  └──────────────────────────────────┘   │
│                                          │
│  ┌──────────────────────────────────┐   │
│  │ Shells out to:                   │   │
│  │ - tiger-cli (DB provisioning)    │   │
│  │ - npm/node (dependency install)  │   │
│  └──────────────────────────────────┘   │
└──────────────────────────────────────────┘
```

## Project Structure

```
0perator/
├── cmd/
│   └── 0perator-mcp/          # MCP server entry point
├── internal/
│   ├── mcp/                   # MCP server implementation
│   ├── cli/                   # CLI commands (init, etc.)
│   ├── orchestrator/          # High-level workflows (TODO)
│   ├── templates/             # App template logic (TODO)
│   ├── tiger/                 # tiger-cli wrapper (TODO)
│   └── runtime/               # Local process management (TODO)
├── templates/                 # Node/TypeScript templates (TODO)
│   └── web-dashboard/
│       ├── template.json      # Template metadata
│       └── src/              # Template files
├── go.mod
└── README.md
```

## Development

```bash
# Build
go build -o 0perator cmd/0perator-mcp/main.go

# Run locally
./0perator init

# Install locally for testing
go install cmd/0perator-mcp/main.go
```

## v0 Scope

- **MCP tools**: `create_app`, `create_db`, `deploy_local`
- **Database layer**: Managed Postgres on Tiger Cloud (via tiger-cli)
- **App templates**: Opinionated Node/TypeScript templates with built-in auth, database, pricing hooks
- **Local Runtime**: Bare-process execution for instant spin-up
- **Artifacts**: Source code written locally

## Success Metrics

- **Speed**: Live app and database under 60 seconds from a single prompt
- **AI-Native**: Exposed via MCP, with prompt templates
- **Simplicity**: No dashboards, YAML, manual setup

## License

TBD
