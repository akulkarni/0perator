# 0perator

Build full-stack applications instantly through natural conversation.

## Overview

0perator is an MCP (Model Context Protocol) server that enables AI agents to create, deploy, and manage real applications. It provides **direct, parameter-driven tools** that create complete, working applications with a single command.

**Key Innovation:** The Direct Tools Architecture - instead of templates or scaffolding, 0perator uses intelligent, parameter-driven tools that adapt to your needs and create everything automatically: dependencies installed, database connected, auth UI built, and dev server running.

## Installation

```bash
# Clone and build
git clone https://github.com/akulkarni/0perator
cd 0perator
go build -o ~/.local/bin/0perator ./cmd/0perator

# Initialize (installs tiger-cli, authenticates, configures IDEs)
0perator init
```

### Init Options

```bash
# Interactive mode (select IDEs with arrow keys)
0perator init

# Configure specific IDE(s)
0perator init --client claude-code
0perator init --client cursor --client windsurf

# Development mode (uses the equivalent of 'go run' instead of binary)
0perator init --dev
```

The init command will:
1. Install tiger-cli if needed
2. Authenticate with Tiger Cloud (opens browser)
3. Configure MCP servers for your selected IDE(s)

## The Stack

0perator creates apps using the **T3 Stack**:

- **Next.js 15** with App Router and React 19
- **tRPC v11** for end-to-end type-safe APIs
- **Drizzle ORM** for type-safe database queries
- **Better Auth** for authentication (optional)
- **Tailwind CSS 4** with shadcn/ui components
- **Tiger Cloud PostgreSQL** with TimescaleDB

Everything is configured and connected automatically.

## Direct Tools (What Claude Sees)

```
🚀 create_web_app   - Create T3 Stack app with database connection
🗄️  create_database - Tiger Cloud PostgreSQL (free tier)
🌐 open_app        - Open app in browser
📖 view_skill      - View step-by-step instructions for complex tasks
```

### Skills

Skills provide step-by-step instructions for complex workflows:

| Skill | Description |
|-------|-------------|
| `create-app` | Full app creation workflow: database setup, auth configuration, shadcn components |



### What's Different?

**Before (Traditional Tools):**
```
- Create app
- "Now cd into directory"
- "Run npm install"
- "Create a .env file"
- "Run the database migrations"
- "Start the dev server"
- 😩 10 manual steps later...
```

**Now (0perator):**
```
- Create app with auth and database
- 🎉 Your app is running at http://localhost:3000!
```

## Real Examples

### Create a Full-Stack App
```
"Create a web app with authentication and a database"
```

0perator will:
- ✅ Create a T3 Stack app (Next.js + tRPC + Drizzle)
- ✅ Set up Tiger Cloud PostgreSQL
- ✅ Configure Better Auth with your chosen providers
- ✅ Initialize shadcn/ui components
- ✅ Connect database and run migrations
- ✅ Install all dependencies

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for build instructions and adding new tools.

## Design Philosophy

**Humans design product not infrastructure:** The developers should think about the product experience, AI should know how to set up the best infrastucture possible. 

**Agentic Ergonomics:** Abstraction layers that help humans can hinder AI agents. 0perator embraces direct, parameter-driven interfaces that AI can use efficiently.

**Zero Config:** When you create something, it should work immediately. No manual steps.

**Best Practices Built-in:** T3 Stack provides type-safety from database to UI with tRPC and Drizzle.

## Success Metrics

- **Speed**: Full-stack app deployed locally in under 30 seconds (after it's well specified).
- **Completeness**: Auth includes UI, database includes schema, everything works
- **Zero Config**: No manual steps, no .env editing, no npm install
- **Quality**: Production-ready code with TypeScript, error handling, best practices

## Future

- Cloud deployment tools
- More UI themes
- Payment integration
- Real-time features
- Testing tools