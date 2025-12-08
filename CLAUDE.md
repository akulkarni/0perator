# 0perator Development Guide

## TypeScript CLI with MCP Server

This project is a CLI tool built with Commander.js that includes an MCP server using `@tigerdata/mcp-boilerplate`.

### Project Structure

```
src/
├── index.ts           # CLI entrypoint with Commander.js
├── config.ts          # Configuration (paths, version)
├── types.ts           # TypeScript types
├── commands/          # CLI commands
│   ├── init.ts        # init command (configure IDEs)
│   └── mcp.ts         # mcp command group
├── scripts/           # Lifecycle scripts
│   └── cleanup.ts     # Runs during npm uninstall
├── lib/               # Shared utilities
│   ├── clients.ts     # Supported IDE clients
│   ├── install.ts     # MCP installation logic
│   └── templates.ts   # Template copy utility
└── mcp/               # MCP server
    ├── server.ts      # MCP server factory
    ├── serverInfo.ts  # Server name/version
    ├── tools/         # MCP tools (ApiFactory pattern)
    │   ├── index.ts
    │   ├── createDatabase.ts
    │   ├── createWebApp.ts
    │   ├── openApp.ts
    │   └── viewSkill.ts
    └── skillutils/    # Skills loading
        └── index.ts

skills/                # Skill markdown files
└── create-app/
    └── SKILL.md

templates/             # App templates copied to new projects
└── app/
    ├── CLAUDE.md
    └── src/styles/globals.css
```

### CLI Commands

```bash
0perator              # Show help
0perator init         # Configure IDEs with MCP servers (interactive)
0perator init --client claude-code --client cursor  # Configure specific IDEs
0perator mcp start    # Start MCP server (used by IDEs)
0perator --version    # Show version
```

### Uninstalling

```bash
npm uninstall -g 0perator  # Automatically runs cleanup script
```

### Adding New CLI Commands

1. Create a new file in `src/commands/`:

```typescript
import { Command } from 'commander';

export function createMyCommand(): Command {
  return new Command('mycommand')
    .description('What this command does')
    .option('--flag <value>', 'Flag description')
    .action(async (options) => {
      // Implementation
    });
}
```

2. Add to `src/index.ts`:

```typescript
import { createMyCommand } from './commands/mycommand.js';
program.addCommand(createMyCommand());
```

### Adding New MCP Tools

1. Create a new file in `src/mcp/tools/`:

```typescript
import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { ServerContext } from '../../types.js';

const inputSchema = {
  myParam: z.string().describe('Parameter description'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether operation succeeded'),
} as const;

type OutputSchema = {
  success: boolean;
};

export const myToolFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => ({
  name: 'my_tool',
  config: {
    title: 'My Tool',
    description: 'What this tool does',
    inputSchema,
    outputSchema,
  },
  fn: async ({ myParam }): Promise<OutputSchema> => {
    return { success: true };
  },
});
```

2. Export from `src/mcp/tools/index.ts`

### Adding New Skills

Create a directory in `skills/` with a `SKILL.md` file:

```markdown
---
name: my-skill
description: 'What this skill teaches'
---

Skill content goes here...
```

Skills are automatically loaded and accessible via the `view_skill` tool.

### Development Commands

```bash
npm run dev           # Run CLI in development mode
npm run dev:mcp       # Run MCP server in development mode
npm run build         # Build for production
npm run watch         # Watch mode - recompile on changes
npm run start         # Run production CLI
npm run inspector     # Open MCP inspector
npm run lint          # Run Biome linter
npm run format        # Format with Biome
```

### MCP Inspector

Test and debug the MCP server in a web UI:

```bash
npm run inspector
```

Opens at http://127.0.0.1:6274 where you can:
- View available tools and their schemas
- Test tool calls with custom inputs
- See request/response payloads

### Dev Mode for MCP Server

When developing, use `--dev` so code changes take effect without rebuilding:

```bash
npm run dev -- init --dev
```

This configures IDEs to run the MCP server via `npx tsx` from source instead of the installed binary. After making code changes, just restart your IDE to pick them up.

To switch back to production mode, run `init` without `--dev`:

```bash
npm run dev -- init
```

### Key Files

- `src/index.ts` - CLI entry point with Commander.js
- `src/commands/` - CLI command implementations
- `src/mcp/tools/index.ts` - MCP tool factories
- `src/mcp/server.ts` - MCP server startup
