# 0perator

Build full-stack applications with AI: real code you own and ship.

## Overview

0perator helps you build modern, well-designed full-stack apps without worrying about the tech stack. You focus on product functionality and design—we handle the infrastructure.

**Unlike vibe-coding platforms,** you own the code: it lives in your local repo with professional workflows like git, branches, and PR reviews. You get full-stack apps with real databases and auth, not just UI prototypes. No vendor lock-in: change providers, add any library, extend beyond what we support.

**Unlike plain AI coding,** you don't have to research the best modern stack or how to wire it together. 0perator gives AI assistants deep, specific guidance on exact patterns—how to structure tRPC routers, which Drizzle patterns to use, how auth integrates with the database. The result is idiomatic, high-quality code with less manual steps.

We also provide optional hardening: backend integration tests and stricter TypeScript checks. These act as a feedback signal—when AI makes a mistake, tests fail and type errors surface immediately, helping it iterate toward correct solutions faster.

## Example Queries

After installing 0perator, try asking your AI coding assistant:

- "Create a collaborative TODO app with user accounts"
- "Build a real-time chat application"
- "Create a dashboard to track my fitness goals"
- "Build a blog with markdown support and comments"
- "Create an expense tracker with categories and charts"

Once your app is built, we also support deployment:

- "Deploy my app to Vercel"

And hardening for better AI-assisted development:

- "Add backend testing"
- "Add stricter TypeScript checks"

Finally, we support you all the way to production:

- "Make this application production ready" (coming soon)

## Installation

```bash
npx 0perator@latest init
```

This will configure your IDE with the MCP servers. Select which IDE to configure when prompted.

### Init Options

```bash
# Interactive mode (select IDE with arrow keys)
npx 0perator@latest init

# Configure specific IDE
npx 0perator@latest init --client claude-code
npx 0perator@latest init --client cursor

# Pin to current version instead of always using latest
npx 0perator init --no-latest
```

The init command will:
1. Install tiger-cli if needed
2. Authenticate with Tiger Cloud (opens browser)
3. Configure MCP servers for your selected IDE(s)

## The Stack

0perator creates apps using the **T3 Stack**:

- **Next.js** with App Router
- **tRPC** for end-to-end type-safe APIs
- **Drizzle ORM** for type-safe database queries
- **Better Auth** for authentication (optional)
- **Tailwind CSS** with shadcn/ui components
- **Tiger Cloud PostgreSQL** with TimescaleDB

Everything is configured and connected automatically.

## How it Works

0perator provides AI assistants with tools and skills via MCP (Model Context Protocol).

Because we're opinionated about the stack, we can write skills with precise, battle-tested instructions—not generic advice that works across many frameworks. This is why being opinionated matters: it lets us teach AI exactly how to build with this stack.

**Skills** are step-by-step guides for complex workflows. AI accesses them via the `view_skill` tool:
- `create-app` - Full app creation: database, auth, shadcn components
- `deploy-app` - Deploy to Vercel with environment configuration
- `add-backend-testing` - Vitest integration tests with isolated test database
- `add-strict-checks` - Stricter TypeScript and linting

**Tools** handle atomic operations that skills orchestrate. Examples:
- `view_skill` - Load a skill to guide the current workflow
- `create_database` - Provision Tiger Cloud PostgreSQL
- `create_web_app` - Scaffold T3 Stack app with database connection
- `open_app` - Open app in browser
- `upload_env_to_vercel` - Upload .env variables to Vercel

## Development

**Dev mode:** Run `npm run dev -- init --dev` from the repo to configure IDEs to run the MCP server from source. Code changes take effect on IDE restart without rebuilding.

See [CLAUDE.md](CLAUDE.md) for development setup, adding new tools, and debugging.

### Testing

When testing app creation, verify that the AI coding agent calls `view_skill('create-app')` in the first few tool calls. This ensures the agent follows the structured workflow for setting up the database, auth, and UI components correctly.

## Design Philosophy

**You own the code:** No platform lock-in. Your code lives in your repo, runs on your infrastructure, and you can change anything—even replace us entirely.

**Humans design product not infrastructure:** You think about the product experience, AI handles the infrastructure decisions.

**Agentic Ergonomics:** Abstraction layers that help humans can hinder AI agents. 0perator embraces direct, parameter-driven interfaces that AI can use efficiently.

**Best Practices Built-in:** T3 Stack provides type-safety from database to UI with tRPC and Drizzle.

## Success Metrics

- **Quality without the effort**: Production-ready code with TypeScript, error handling, best practices
- **Completeness**: Auth includes UI, database includes schema, everything works
- **Minimal Config**: No manual steps, no .env editing, no npm install

## Future

- More UI themes
- Payment integration
- Real-time features
