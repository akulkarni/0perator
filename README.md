# 0perator

Build full-stack applications instantly through natural conversation.

## Overview

0perator helps you build modern, well-designed full-stack apps without worrying about the tech stack. We've chosen a best-in-class TypeScript stack and built world-class AI support through MCP. You focus on product functionality and design—we handle the infrastructure.

## Example Queries

After installing 0perator, try asking your AI coding assistant:

- "Create a collaborative TODO app with user accounts"
- "Build a real-time chat application"
- "Create a dashboard to track my fitness goals"
- "Build a blog with markdown support and comments"
- "Create an expense tracker with categories and charts"

Once your app is built, we also support deployment:

- "Deploy my app to Vercel"

(More deployment options coming soon)

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

- **Next.js 15** with App Router and React 19
- **tRPC v11** for end-to-end type-safe APIs
- **Drizzle ORM** for type-safe database queries
- **Better Auth** for authentication (optional)
- **Tailwind CSS 4** with shadcn/ui components
- **Tiger Cloud PostgreSQL** with TimescaleDB

Everything is configured and connected automatically.

## Direct Tools (What Claude Sees)

```
🚀 create_web_app        - Create T3 Stack app with database connection
🗄️  create_database       - Tiger Cloud PostgreSQL (free tier)
🌐 open_app              - Open app in browser
📤 upload_env_to_vercel  - Upload .env variables to Vercel
📖 view_skill            - View step-by-step instructions for complex tasks
```

### Skills

Skills provide step-by-step instructions for complex workflows:

| Skill | Description |
|-------|-------------|
| `create-app` | Full app creation workflow: database setup, auth configuration, shadcn components |
| `deploy-app` | Deploy your app to Vercel with environment configuration |



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

## Development

**Dev mode:** Run `npm run dev -- init --dev` from the repo to configure IDEs to run the MCP server from source. Code changes take effect on IDE restart without rebuilding.

See [CLAUDE.md](CLAUDE.md) for development setup, adding new tools, and debugging.

### Testing

When testing app creation, verify that the AI coding agent calls `view_skill('create-app')` in the first few tool calls. This ensures the agent follows the structured workflow for setting up the database, auth, and UI components correctly.

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