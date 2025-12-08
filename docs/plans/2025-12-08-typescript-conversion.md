# 0perator TypeScript Conversion Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Convert 0perator CLI tool from Go to TypeScript with MCP server support

**Architecture:** Full CLI tool using Commander.js with commands: `init`, `uninstall`, `mcp start`. MCP server uses mcp-boilerplate factory pattern with tools defined as ApiFactory functions. Skills loaded from markdown files with gray-matter frontmatter.

**Tech Stack:** TypeScript, Commander.js, @tigerdata/mcp-boilerplate, Zod, gray-matter, Node.js 20+

---

### Task 1: Initialize Project Structure

**Files:**
- Create: `package.json`
- Create: `tsconfig.json`
- Create: `.gitignore`

**Step 1: Create package.json**

```json
{
  "name": "0perator",
  "version": "2.0.3",
  "description": "Build full-stack applications instantly through natural conversation",
  "author": "Tiger Data",
  "type": "module",
  "bin": {
    "0perator": "dist/index.js"
  },
  "files": [
    "dist",
    "skills",
    "templates"
  ],
  "scripts": {
    "build": "tsc && shx chmod +x dist/*.js",
    "prepare": "npm run build",
    "preuninstall": "node dist/scripts/cleanup.js",
    "dev": "tsx src/index.ts",
    "dev:mcp": "tsx src/index.ts mcp start",
    "start": "node dist/index.js",
    "inspector": "npx @modelcontextprotocol/inspector node dist/index.js mcp start",
    "lint": "biome check",
    "lint:fix": "biome check --fix",
    "format": "biome format --write"
  },
  "dependencies": {
    "@inquirer/prompts": "^7.2.0",
    "@tigerdata/mcp-boilerplate": "^0.6.0",
    "commander": "^12.1.0",
    "gray-matter": "^4.0.3",
    "picocolors": "^1.1.1",
    "zod": "^3.25.76"
  },
  "devDependencies": {
    "@biomejs/biome": "^1.9.4",
    "@types/node": "^22.19.1",
    "shx": "^0.4.0",
    "tsx": "^4.20.6",
    "typescript": "^5.9.3"
  }
}
```

**Step 2: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src",
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true
  },
  "include": ["./src/**/*.ts"],
  "exclude": ["node_modules", "dist"]
}
```

**Step 3: Create biome.json**

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.4/schema.json",
  "vcs": {
    "enabled": true,
    "clientKind": "git",
    "useIgnoreFile": true
  },
  "organizeImports": {
    "enabled": true
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true
    }
  },
  "formatter": {
    "enabled": true
  }
}
```

**Step 4: Update .gitignore for Node.js**

```gitignore
# Dependencies
node_modules/

# Build output
dist/

# Environment
.env
.env.local

# IDE
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
npm-debug.log*
```

**Step 5: Install dependencies**

Run: `npm install`
Expected: Dependencies installed, node_modules created

**Step 6: Verify TypeScript compiles**

Run: `npx tsc --noEmit`
Expected: No errors (no source files yet, but config is valid)

**Step 7: Commit**

```bash
git add package.json tsconfig.json biome.json .gitignore
git commit -m "chore: initialize TypeScript project structure"
```

---

### Task 2: Create Core Files and Directory Structure

**Files:**
- Create: `src/types.ts`
- Create: `src/config.ts`
- Create: `src/mcp/serverInfo.ts`

**Step 1: Create src/types.ts**

```typescript
export interface ServerContext extends Record<string, unknown> {
  // No database connection needed for 0perator
  // Context can be extended later if needed
}

export interface ClientInfo {
  name: string;
  displayName: string;
}
```

**Step 2: Create src/config.ts**

```typescript
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import { readFileSync } from 'fs';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Package root directory (relative to dist/)
export const packageRoot = join(__dirname, '..');

// Skills directory at package root level
export const skillsDir = join(packageRoot, 'skills');

// Templates directory at package root level
export const templatesDir = join(packageRoot, 'templates');

// Read version from package.json
const pkg = JSON.parse(readFileSync(join(packageRoot, 'package.json'), 'utf-8'));
export const version: string = pkg.version;
```

**Step 3: Create src/mcp/serverInfo.ts**

```typescript
import { ServerContext } from '../types.js';
import { version } from '../config.js';

export const serverInfo = {
  name: '0perator',
  version,
} as const;

export const context: ServerContext = {};
```

**Step 4: Verify files compile**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 5: Commit**

```bash
git add src/types.ts src/config.ts src/mcp/serverInfo.ts
git commit -m "feat: add core types and configuration"
```

---

### Task 3: Create Skills Loading Utility

**Files:**
- Create: `src/mcp/skillutils/index.ts`

**Step 1: Create src/mcp/skillutils/index.ts**

```typescript
import { readdir, readFile } from 'fs/promises';
import { join } from 'path';
import matter from 'gray-matter';
import { z } from 'zod';
import { log } from '@tigerdata/mcp-boilerplate';
import { skillsDir } from '../../config.js';

// ===== Skill Types =====

export const zSkillMatter = z.object({
  name: z.string().trim().min(1),
  description: z.string(),
});
export type SkillMatter = z.infer<typeof zSkillMatter>;

export const zSkill = z.object({
  path: z.string(),
  name: z.string(),
  description: z.string(),
});
export type Skill = z.infer<typeof zSkill>;

// ===== Skill Loading Implementation =====

// Cache for skill content
const skillContentCache: Map<string, string> = new Map();
let skillMapPromise: Promise<Map<string, Skill>> | null = null;

/**
 * Parse a SKILL.md file and validate its metadata
 */
const parseSkillFile = async (
  fileContent: string,
): Promise<{
  matter: SkillMatter;
  content: string;
}> => {
  const { data, content } = matter(fileContent);
  const skillMatter = zSkillMatter.parse(data);

  // Normalize skill name
  if (!/^[a-zA-Z0-9-_]+$/.test(skillMatter.name)) {
    const normalized = skillMatter.name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-_]/g, '_')
      .replace(/-[-_]+/g, '-')
      .replace(/_[_-]+/g, '_')
      .replace(/(^[-_]+)|([-_]+$)/g, '');
    log.warn(
      `Skill name "${skillMatter.name}" contains invalid characters. Normalizing to "${normalized}".`,
    );
    skillMatter.name = normalized;
  }

  return {
    matter: skillMatter,
    content: content.trim(),
  };
};

/**
 * Load all skills from the filesystem
 */
async function doLoadSkills(): Promise<Map<string, Skill>> {
  const skills = new Map<string, Skill>();
  skillContentCache.clear();

  const alreadyExists = (name: string, path: string): boolean => {
    const existing = skills.get(name);
    if (existing) {
      log.warn(
        `Skill with name "${name}" already loaded from path "${existing.path}". Skipping duplicate at path "${path}".`,
      );
      return true;
    }
    return false;
  };

  const loadLocalPath = async (path: string): Promise<void> => {
    const skillPath = join(path, 'SKILL.md');
    try {
      const fileContent = await readFile(skillPath, 'utf-8');
      const {
        matter: { name, description },
        content,
      } = await parseSkillFile(fileContent);

      if (alreadyExists(name, path)) return;

      skills.set(name, {
        path,
        name,
        description,
      });

      skillContentCache.set(`${name}/SKILL.md`, content);
    } catch (err) {
      log.error(`Failed to load skill at path: ${skillPath}`, err as Error);
    }
  };

  try {
    // Load skills from subdirectories with SKILL.md files
    const dirEntries = await readdir(skillsDir, { withFileTypes: true });
    for (const entry of dirEntries) {
      if (!entry.isDirectory()) continue;
      await loadLocalPath(join(skillsDir, entry.name));
    }

    if (skills.size === 0) {
      log.warn(
        'No skills found. Please add SKILL.md files to the skills/ subdirectories.',
      );
    } else {
      log.info(`Successfully loaded ${skills.size} skill(s)`);
    }
  } catch (err) {
    log.error('Failed to load skills', err as Error);
  }

  return skills;
}

/**
 * Load skills with caching
 */
export const loadSkills = async (
  force = false,
): Promise<Map<string, Skill>> => {
  if (skillMapPromise && !force) {
    return skillMapPromise;
  }

  skillMapPromise = doLoadSkills().catch((err) => {
    log.error('Failed to load skills', err as Error);
    skillMapPromise = null;
    return new Map<string, Skill>();
  });

  return skillMapPromise;
};

/**
 * View skill content
 */
export const viewSkillContent = async (
  name: string,
  targetPath = 'SKILL.md',
): Promise<string> => {
  const skillsMap = await loadSkills();
  const skill = skillsMap.get(name);
  if (!skill) {
    throw new Error(`Skill not found: ${name}`);
  }

  const cacheKey = `${name}/${targetPath}`;
  const cached = skillContentCache.get(cacheKey);
  if (cached) {
    return cached;
  }

  // Read from filesystem
  try {
    const fullPath = join(skill.path, targetPath);
    const content = await readFile(fullPath, 'utf-8');
    skillContentCache.set(cacheKey, content);
    return content;
  } catch {
    throw new Error(`Failed to read skill content: ${name}/${targetPath}`);
  }
};

// Initialize skills on module load
export const skills = await loadSkills();
```

**Step 2: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add src/mcp/skillutils/index.ts
git commit -m "feat: add skills loading utility with gray-matter parsing"
```

---

### Task 4: Create view_skill Tool

**Files:**
- Create: `src/mcp/tools/viewSkill.ts`
- Create: `src/mcp/tools/index.ts`

**Step 1: Create src/mcp/tools/viewSkill.ts**

```typescript
import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { ServerContext } from '../../types.js';
import { skills, viewSkillContent } from '../skillutils/index.js';

// Create enum schema dynamically from loaded skills
const inputSchema = {
  name: z
    .enum(Array.from(skills.keys()) as [string, ...string[]])
    .describe('Skill name (directory name)'),
} as const;

const outputSchema = {
  success: z.boolean(),
  name: z.string().optional(),
  description: z.string().optional(),
  body: z.string().optional(),
  error: z.string().optional(),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

export const viewSkillFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'view_skill',
    config: {
      title: 'View Skill',
      description: `📖 View instructions for a specific skill by name.

Available skills:
${Array.from(skills.values())
  .map((s) => `- ${s.name}: ${s.description}`)
  .join('\n')}
`,
      inputSchema,
      outputSchema,
    },
    fn: async ({ name }): Promise<OutputSchema> => {
      const skill = skills.get(name);

      if (!skill) {
        return {
          success: false,
          error: `Skill '${name}' not found`,
        };
      }

      const body = await viewSkillContent(name);

      return {
        success: true,
        name: skill.name,
        description: skill.description || '',
        body,
      };
    },
  };
};
```

**Step 2: Create src/mcp/tools/index.ts (partial - will add more tools)**

```typescript
import { viewSkillFactory } from './viewSkill.js';

export const apiFactories = [
  viewSkillFactory,
] as const;
```

**Step 3: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add src/mcp/tools/viewSkill.ts src/mcp/tools/index.ts
git commit -m "feat: add view_skill tool"
```

---

### Task 5: Create create_database Tool

**Files:**
- Create: `src/mcp/tools/createDatabase.ts`
- Modify: `src/mcp/tools/index.ts`

**Step 1: Create src/mcp/tools/createDatabase.ts**

```typescript
import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { exec } from 'child_process';
import { promisify } from 'util';
import { ServerContext } from '../../types.js';

const execAsync = promisify(exec);

const inputSchema = {
  name: z
    .string()
    .optional()
    .describe('Database name (default: app-db)'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether the database was created successfully'),
  service_id: z.string().optional().describe('The Tiger Cloud service ID'),
  error: z.string().optional().describe('Error message if creation failed'),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

export const createDatabaseFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'create_database',
    config: {
      title: 'Create Database',
      description: '🗄️ Set up any database - PostgreSQL on Tiger Cloud (default, FREE). Auto-configures with schema, migrations, and connection handling. Use for any database request.',
      inputSchema,
      outputSchema,
    },
    fn: async ({ name }): Promise<OutputSchema> => {
      const dbName = name || 'app-db';

      const cmdArgs = [
        'tiger',
        'service', 'create',
        '--name', dbName,
        '--cpu', 'shared',
        '--memory', 'shared',
        '--addons', 'time-series,ai',
        '--wait-timeout', '2m',
        '-o', 'json',
      ];

      try {
        const { stdout, stderr } = await execAsync(cmdArgs.join(' '));
        const result = JSON.parse(stdout) as { service_id?: string };

        if (!result.service_id) {
          return {
            success: false,
            error: `No service_id in response: ${stdout}${stderr}`,
          };
        }

        return {
          success: true,
          service_id: result.service_id,
        };
      } catch (err) {
        const error = err as Error & { stdout?: string; stderr?: string };
        return {
          success: false,
          error: `Failed to create database: ${error.message}\n${error.stdout || ''}${error.stderr || ''}`,
        };
      }
    },
  };
};
```

**Step 2: Update src/mcp/tools/index.ts**

```typescript
import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';

export const apiFactories = [
  createDatabaseFactory,
  viewSkillFactory,
] as const;
```

**Step 3: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add src/mcp/tools/createDatabase.ts src/mcp/tools/index.ts
git commit -m "feat: add create_database tool"
```

---

### Task 6: Create Template Utilities

**Files:**
- Create: `src/lib/templates.ts`
- Create: `templates/app/CLAUDE.md` (copy from Go project)
- Create: `templates/app/src/styles/globals.css` (copy from Go project)

**Step 1: Create src/lib/templates.ts**

```typescript
import { readdir, readFile, writeFile, mkdir } from 'fs/promises';
import { dirname, join, relative } from 'path';
import { templatesDir } from '../config.js';

/**
 * Copy app templates (CLAUDE.md, globals.css, etc.) to destination
 */
export async function writeAppTemplates(destDir: string): Promise<void> {
  const appDir = join(templatesDir, 'app');

  async function copyDir(srcDir: string, destBase: string): Promise<void> {
    const entries = await readdir(srcDir, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = join(srcDir, entry.name);
      const relPath = relative(appDir, srcPath);
      const destPath = join(destBase, relPath);

      if (entry.isDirectory()) {
        await mkdir(destPath, { recursive: true });
        await copyDir(srcPath, destBase);
      } else {
        // Ensure parent directory exists
        await mkdir(dirname(destPath), { recursive: true });
        const content = await readFile(srcPath);
        await writeFile(destPath, content);
      }
    }
  }

  await copyDir(appDir, destDir);
}
```

**Step 2: Create templates/app/CLAUDE.md**

Copy from `/Users/cevian/Development/0peratorGo/internal/templates/app/CLAUDE.md`

**Step 3: Create templates/app/src/styles/globals.css**

Copy from `/Users/cevian/Development/0peratorGo/internal/templates/app/src/styles/globals.css`

**Step 4: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 5: Commit**

```bash
git add src/lib/templates.ts templates/
git commit -m "feat: add template utilities and app templates"
```

---

### Task 7: Create create_web_app Tool

**Files:**
- Create: `src/mcp/tools/createWebApp.ts`
- Modify: `src/mcp/tools/index.ts`

**Step 1: Create src/mcp/tools/createWebApp.ts**

```typescript
import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { exec } from 'child_process';
import { promisify } from 'util';
import { readFile, writeFile, unlink } from 'fs/promises';
import { join } from 'path';
import { ServerContext } from '../../types.js';
import { writeAppTemplates } from '../../lib/templates.js';

const execAsync = promisify(exec);

const inputSchema = {
  name: z
    .string()
    .optional()
    .describe('Application name'),
  db_service_id: z
    .string()
    .optional()
    .describe('Database service ID to connect to'),
  use_auth: z
    .boolean()
    .optional()
    .describe('Enable authentication'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether the app was created successfully'),
  message: z.string().describe('Status message'),
  path: z.string().optional().describe('Path to created app'),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

/**
 * Replace the value of a variable in a .env file
 */
async function replaceEnvValue(envPath: string, key: string, value: string): Promise<void> {
  const envData = await readFile(envPath, 'utf-8');
  const lines = envData.split('\n');

  for (let i = 0; i < lines.length; i++) {
    if (lines[i].startsWith(`${key}=`)) {
      lines[i] = `${key}=${value}`;
      break;
    }
  }

  await writeFile(envPath, lines.join('\n'));
}

export const createWebAppFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'create_web_app',
    config: {
      title: 'Create Web App',
      description: '🚀 Create any web application - Build an opinionated next.js app',
      inputSchema,
      outputSchema,
    },
    fn: async ({ name, db_service_id, use_auth }): Promise<OutputSchema> => {
      const appName = name || 'my-app';

      if (!db_service_id) {
        return {
          success: false,
          message: 'db_service_id is required',
        };
      }

      try {
        // Create T3 app
        const t3Args = [
          'npx', 'create-t3-app@latest', appName,
          '--noGit', '--CI', '--tailwind', '--drizzle', '--trpc',
          '--dbProvider', 'postgres', '--appRouter',
        ];
        if (use_auth) {
          t3Args.push('--betterAuth');
        }

        await execAsync(t3Args.join(' '));

        // Initialize shadcn UI
        await execAsync('npx shadcn@latest init --base-color=neutral', {
          cwd: appName,
        });

        // Get database connection string from Tiger
        const { stdout: serviceJson } = await execAsync(
          `tiger service get ${db_service_id} --with-password -o json`
        );
        const serviceDetails = JSON.parse(serviceJson) as { connection_string?: string };

        if (!serviceDetails.connection_string) {
          return {
            success: false,
            message: 'connection_string not found in service details',
          };
        }

        // Update .env with database connection
        const envPath = join(appName, '.env');
        await replaceEnvValue(envPath, 'DATABASE_URL', serviceDetails.connection_string);

        // Remove start-database script if it exists
        try {
          await unlink(join(appName, 'start-database.sh'));
        } catch {
          // Ignore if file doesn't exist
        }

        // Copy app templates (CLAUDE.md, globals.css)
        await writeAppTemplates(appName);

        return {
          success: true,
          message: `Created app '${appName}'`,
          path: appName,
        };
      } catch (err) {
        const error = err as Error & { stderr?: string };
        return {
          success: false,
          message: `Failed to create app: ${error.message}\n${error.stderr || ''}`,
        };
      }
    },
  };
};
```

**Step 2: Update src/mcp/tools/index.ts**

```typescript
import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';
import { createWebAppFactory } from './createWebApp.js';

export const apiFactories = [
  createDatabaseFactory,
  createWebAppFactory,
  viewSkillFactory,
] as const;
```

**Step 3: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add src/mcp/tools/createWebApp.ts src/mcp/tools/index.ts
git commit -m "feat: add create_web_app tool"
```

---

### Task 8: Create open_app Tool

**Files:**
- Create: `src/mcp/tools/openApp.ts`
- Modify: `src/mcp/tools/index.ts`

**Step 1: Create src/mcp/tools/openApp.ts**

```typescript
import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { exec } from 'child_process';
import { ServerContext } from '../../types.js';

const inputSchema = {
  url: z
    .string()
    .optional()
    .describe('URL to open (default: http://localhost:3000)'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether the browser was opened successfully'),
  message: z.string().describe('Status message'),
  url: z.string().describe('The URL that was opened'),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

export const openAppFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'open_app',
    config: {
      title: 'Open App',
      description: '🌐 Open the app in browser. Call this AFTER all setup (database, auth, UI) is complete to show the user their running app.',
      inputSchema,
      outputSchema,
    },
    fn: async ({ url }): Promise<OutputSchema> => {
      const targetUrl = url || 'http://localhost:3000';

      return new Promise((resolve) => {
        // Use 'open' command on macOS
        exec(`open "${targetUrl}"`, (err) => {
          if (err) {
            resolve({
              success: false,
              message: `Failed to open browser: ${err.message}`,
              url: targetUrl,
            });
          } else {
            resolve({
              success: true,
              message: `Opened ${targetUrl} in browser`,
              url: targetUrl,
            });
          }
        });
      });
    },
  };
};
```

**Step 2: Update src/mcp/tools/index.ts**

```typescript
import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';
import { createWebAppFactory } from './createWebApp.js';
import { openAppFactory } from './openApp.js';

export const apiFactories = [
  createDatabaseFactory,
  createWebAppFactory,
  openAppFactory,
  viewSkillFactory,
] as const;
```

**Step 3: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add src/mcp/tools/openApp.ts src/mcp/tools/index.ts
git commit -m "feat: add open_app tool"
```

---

### Task 9: Create MCP Server Module

**Files:**
- Create: `src/mcp/server.ts`

**Step 1: Create src/mcp/server.ts**

```typescript
import { stdioServerFactory } from '@tigerdata/mcp-boilerplate';
import { apiFactories } from './tools/index.js';
import { context, serverInfo } from './serverInfo.js';

/**
 * Start the MCP server in stdio mode
 */
export async function startMcpServer(): Promise<void> {
  await stdioServerFactory({
    ...serverInfo,
    context,
    apiFactories,
  });
}
```

**Step 2: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 3: Commit**

```bash
git add src/mcp/server.ts
git commit -m "feat: add MCP server module"
```

---

### Task 10: Create MCP Installation Library

**Files:**
- Create: `src/lib/clients.ts`
- Create: `src/lib/install.ts`

**Step 1: Create src/lib/clients.ts**

```typescript
import { ClientInfo } from '../types.js';

export const supportedClients: ClientInfo[] = [
  { name: 'claude-code', displayName: 'Claude Code' },
  { name: 'cursor', displayName: 'Cursor' },
  { name: 'windsurf', displayName: 'Windsurf' },
];

export function getClientByName(name: string): ClientInfo | undefined {
  return supportedClients.find(c => c.name === name);
}
```

**Step 2: Create src/lib/install.ts**

```typescript
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface InstallOptions {
  devMode?: boolean;
}

/**
 * Install Tiger MCP for the given IDE client
 */
export async function installTigerMcp(clientName: string): Promise<void> {
  try {
    await execAsync(`tiger mcp install ${clientName} --no-backup`);
  } catch (err) {
    const error = err as Error & { stderr?: string };
    // Ignore if already installed
    if (!error.stderr?.includes('already exists')) {
      throw new Error(`Failed to install Tiger MCP: ${error.message}`);
    }
  }
}

/**
 * Install 0perator MCP for the given IDE client
 * Uses tiger-cli's mcpinstall under the hood
 */
export async function install0peratorMcp(
  clientName: string,
  options: InstallOptions = {}
): Promise<void> {
  // For now, use tiger mcp install with 0perator binary path
  // In production, this will use the installed 0perator binary
  const operatorPath = process.argv[1]; // Current executable
  const args = options.devMode ? '' : 'mcp start';

  // Use tiger CLI to install MCP config
  await execAsync(
    `tiger mcp install-raw ${clientName} --name 0perator --command "${operatorPath}" --args "${args}" --no-backup`
  );
}

/**
 * Install both Tiger and 0perator MCP servers
 */
export async function installBoth(
  clientName: string,
  options: InstallOptions = {}
): Promise<void> {
  await installTigerMcp(clientName);
  await install0peratorMcp(clientName, options);
}
```

**Step 3: Verify compilation**

Run: `npx tsc --noEmit`
Expected: No errors

**Step 4: Commit**

```bash
git add src/lib/clients.ts src/lib/install.ts
git commit -m "feat: add MCP installation library"
```

---

### Task 11: Create CLI Entry Point with Commander.js

**Files:**
- Create: `src/index.ts`
- Create: `src/commands/mcp.ts`

**Step 1: Create src/commands/mcp.ts**

```typescript
import { Command } from 'commander';
import { startMcpServer } from '../mcp/server.js';

export function createMcpCommand(): Command {
  const mcp = new Command('mcp')
    .description('MCP server commands');

  mcp
    .command('start')
    .description('Start the MCP server')
    .action(async () => {
      await startMcpServer();
    });

  return mcp;
}
```

**Step 2: Create src/index.ts**

```typescript
#!/usr/bin/env node
import { Command } from 'commander';
import { version } from './config.js';
import { createMcpCommand } from './commands/mcp.js';

const program = new Command();

program
  .name('0perator')
  .description('Infrastructure for AI native development')
  .version(version);

program.addCommand(createMcpCommand());

program.parse();
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Compiles successfully

**Step 4: Test CLI**

Run: `npm run dev -- --help`
Expected: Shows help with `mcp` command listed

**Step 5: Commit**

```bash
git add src/index.ts src/commands/mcp.ts
git commit -m "feat: add CLI entry point with Commander.js"
```

---

### Task 12: Create init Command

**Files:**
- Create: `src/commands/init.ts`
- Modify: `src/index.ts`

**Step 1: Create src/commands/init.ts**

```typescript
import { Command } from 'commander';
import { checkbox } from '@inquirer/prompts';
import pc from 'picocolors';
import { supportedClients } from '../lib/clients.js';
import { installBoth } from '../lib/install.js';

interface InitOptions {
  client?: string[];
  dev?: boolean;
}

export function createInitCommand(): Command {
  const init = new Command('init')
    .description('Configure IDEs with MCP servers')
    .option('--client <name>', 'Client to configure (can be repeated)', collect, [])
    .option('--dev', 'Use development mode')
    .action(async (options: InitOptions) => {
      let clients = options.client || [];

      // If no clients specified, prompt interactively
      if (clients.length === 0) {
        clients = await checkbox({
          message: 'Select IDEs to configure:',
          choices: supportedClients.map(c => ({
            name: c.displayName,
            value: c.name,
          })),
        });
      }

      if (clients.length === 0) {
        console.log(pc.yellow('No IDEs selected. Exiting.'));
        return;
      }

      console.log(pc.blue('\nConfiguring MCP servers...\n'));

      for (const clientName of clients) {
        const client = supportedClients.find(c => c.name === clientName);
        if (!client) {
          console.log(pc.red(`Unknown client: ${clientName}`));
          continue;
        }

        try {
          console.log(`  ${pc.cyan('→')} ${client.displayName}...`);
          await installBoth(clientName, { devMode: options.dev });
          console.log(`  ${pc.green('✓')} ${client.displayName} configured`);
        } catch (err) {
          const error = err as Error;
          console.log(`  ${pc.red('✗')} ${client.displayName}: ${error.message}`);
        }
      }

      console.log(pc.green('\nDone! Restart your IDE to use the MCP servers.'));
    });

  return init;
}

function collect(value: string, previous: string[]): string[] {
  return previous.concat([value]);
}
```

**Step 2: Update src/index.ts**

```typescript
#!/usr/bin/env node
import { Command } from 'commander';
import { version } from './config.js';
import { createMcpCommand } from './commands/mcp.js';
import { createInitCommand } from './commands/init.js';

const program = new Command();

program
  .name('0perator')
  .description('Infrastructure for AI native development')
  .version(version);

program.addCommand(createInitCommand());
program.addCommand(createMcpCommand());

program.parse();
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Compiles successfully

**Step 4: Test init command**

Run: `npm run dev -- init --help`
Expected: Shows init command help with --client and --dev options

**Step 5: Commit**

```bash
git add src/commands/init.ts src/index.ts
git commit -m "feat: add init command with interactive IDE selection"
```

---

### Task 13: Create preuninstall Cleanup Script

**Files:**
- Create: `src/scripts/cleanup.ts`
- Modify: `package.json` (add preuninstall script)

**Step 1: Create src/scripts/cleanup.ts**

```typescript
#!/usr/bin/env node
/**
 * Cleanup script that runs during `npm uninstall -g 0perator`
 * Removes configuration files and directories.
 */
import { rm } from 'fs/promises';
import { join } from 'path';
import { homedir } from 'os';

async function cleanup(): Promise<void> {
  // Remove config directory
  const configDir = join(homedir(), '.config', '0perator');
  try {
    await rm(configDir, { recursive: true, force: true });
    console.log('0perator: Removed config directory');
  } catch {
    // Directory might not exist, that's fine
  }

  console.log('0perator: Cleanup complete');
  console.log('0perator: Please manually remove \'0perator\' from your IDE\'s MCP configuration');
}

cleanup().catch((err) => {
  console.error('0perator: Cleanup failed:', err);
  // Don't exit with error code - let npm uninstall continue
});
```

**Step 2: Update package.json scripts**

Add to the "scripts" section:
```json
"preuninstall": "node dist/scripts/cleanup.js"
```

**Step 3: Update files array in package.json**

The `dist` folder already includes all compiled files, so `dist/scripts/cleanup.js` will be included.

**Step 4: Verify build**

Run: `npm run build`
Expected: Compiles successfully, `dist/scripts/cleanup.js` exists

**Step 5: Commit**

```bash
git add src/scripts/cleanup.ts package.json
git commit -m "feat: add preuninstall cleanup script"
```

---

### Task 14: Update Documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md`

**Step 1: Update CLAUDE.md for TypeScript**

```markdown
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
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
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
npm run start         # Run production CLI
npm run inspector     # Open MCP inspector
npm run lint          # Run Biome linter
npm run format        # Format with Biome
```

### Key Files

- `src/index.ts` - CLI entry point with Commander.js
- `src/commands/` - CLI command implementations
- `src/mcp/tools/index.ts` - MCP tool factories
- `src/mcp/server.ts` - MCP server startup
```

**Step 2: Update README.md**

Update the installation and development sections to reflect TypeScript/npm commands.

**Step 3: Commit**

```bash
git add CLAUDE.md README.md
git commit -m "docs: update documentation for TypeScript CLI"
```

---

### Task 15: Final Verification

**Step 1: Clean build**

Run: `rm -rf dist node_modules && npm install && npm run build`
Expected: Clean install and build succeeds

**Step 2: Test CLI help**

Run: `npm run start -- --help`
Expected: Shows help with commands: init, mcp, uninstall

**Step 3: Test MCP server**

Run: `echo '{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}' | npm run start -- mcp start`
Expected: Returns JSON-RPC response with server info

**Step 4: Test with MCP inspector**

Run: `npm run inspector`
Expected: Inspector opens, can see tools: create_database, create_web_app, open_app, view_skill

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: complete TypeScript conversion"
```

---

## Summary

This plan converts 0perator from Go to TypeScript with:

- **CLI Framework**: Commander.js with commands: `init`, `mcp start`
- **Interactive IDE Selection**: Using @inquirer/prompts for checkbox selection
- **4 MCP Tools**: create_database, create_web_app, open_app, view_skill
- **Skills System**: Markdown files with frontmatter, accessible via view_skill tool
- **Templates**: App templates (CLAUDE.md, globals.css) copied to new projects
- **Linting/Formatting**: Biome
- **Uninstall**: Automatic cleanup via npm preuninstall lifecycle script

Total: 15 tasks, each with explicit steps, exact file contents, and verification commands.

---

## Post-Implementation

### MCP Server Instructions

The Go version includes server instructions that guide the LLM:

```go
Instructions: `When the user asks to build a web application, SaaS app, or any app: use the view_skill tool for the skill named create-app.`
```

Currently `@tigerdata/mcp-boilerplate` does not expose the `instructions` option when creating the MCP server. To add this:

**How mcp-boilerplate would pass instructions to the SDK:**

The official `@modelcontextprotocol/sdk` Server class accepts `instructions` in its options:

```typescript
// In @modelcontextprotocol/sdk/server/index.d.ts
interface ServerOptions {
  capabilities?: ServerCapabilities;
  instructions?: string;  // <-- This is what we need
  // ...
}
```

In `mcp-boilerplate/src/mcpServer.ts`, the `McpServer` is created like this:

```typescript
const server = new McpServer({
  name,
  version,
}, {
  capabilities: { ... },
  // instructions is not passed here currently
});
```

To fix, update `mcpServerFactory` to accept and pass `instructions`:

```typescript
export const mcpServerFactory = <Context>({
  name,
  version = '1.0.0',
  instructions,  // <-- Add this parameter
  // ...other params
}) => {
  const server = new McpServer({
    name,
    version,
  }, {
    capabilities: { ... },
    instructions,  // <-- Pass it here
  });
  // ...
};
```

Then update `stdioServerFactory` similarly to accept and forward `instructions`.

**After mcp-boilerplate is updated**, update `src/mcp/server.ts` to pass instructions:

```typescript
import { stdioServerFactory } from '@tigerdata/mcp-boilerplate';
import { apiFactories } from './tools/index.js';
import { context, serverInfo } from './serverInfo.js';

const serverInstructions = `When the user asks to build a web application, SaaS app, or any app: use the view_skill tool for the skill named create-app.`;

export async function startMcpServer(): Promise<void> {
  await stdioServerFactory({
    ...serverInfo,
    context,
    apiFactories,
    instructions: serverInstructions, // Once mcp-boilerplate supports this
  });
}
```
