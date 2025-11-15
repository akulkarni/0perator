---
title: Create Web Application
description: Build a production-ready web application with Node.js, TypeScript, Fastify, and a flexible frontend
tags: [web, backend, nodejs, typescript, fastify, api, foundational]
category: foundational
dependencies: []
related: [database_tiger, auth_jwt, payments_stripe, deploy_cloudflare]
---

# Create Web Application

⚡ **Target: Complete in 1-2 minutes**

## Speed Optimization

This template is optimized for fast implementation:
- **Create all files in parallel** (use multiple Write tool calls in one message)
- **Copy code exactly as shown** (minimal modifications needed)
- **Skip testing until user requests it**
- **If database needed, start provisioning FIRST** (see database_tiger template)

## Overview

This guide walks you through creating a production-ready web application using:
- **Backend:** Node.js 20+ with TypeScript
- **Framework:** Fastify (fast, low-overhead, TypeScript-first)
- **Validation:** Zod schemas with type inference
- **Frontend:** HTML/CSS/JavaScript (adaptable to requirements)
- **Dev Tools:** Hot reload, linting, formatting

## Architecture

```
my-app/
├── package.json           # Dependencies and scripts
├── tsconfig.json          # TypeScript configuration
├── .env                   # Environment variables (not committed)
├── .gitignore            # Git ignore rules
├── .prettierrc           # Code formatting rules
├── .eslintrc.js          # Linting rules
└── src/
    ├── index.ts          # Server entry point
    ├── routes/           # Route handlers
    │   ├── index.ts      # Homepage route
    │   └── api.ts        # API endpoints
    ├── lib/              # Utilities and helpers
    │   └── validation.ts # Zod schemas
    ├── types/            # TypeScript type definitions
    │   └── index.ts      # Shared types
    └── public/           # Static files (HTML, CSS, JS)
        ├── index.html    # Homepage
        └── styles.css    # Styles
```

## Prerequisites

- Node.js 20+ installed
- Basic TypeScript knowledge
- Text editor or IDE

## Step-by-Step Implementation

### Step 1: Initialize Project

Use `execute` operation to create the project directory and initialize:

```json
{
  "operation": "run_command",
  "params": {
    "command": "mkdir my-app && cd my-app && npm init -y"
  }
}
```

### Step 2: Create package.json

Create a complete `package.json` with all dependencies:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/package.json",
    "content": "{
  \"name\": \"my-app\",
  \"version\": \"1.0.0\",
  \"type\": \"module\",
  \"scripts\": {
    \"dev\": \"nodemon --exec tsx src/index.ts\",
    \"build\": \"tsc\",
    \"start\": \"node dist/index.js\",
    \"format\": \"prettier --write \\\"src/**/*.ts\\\"\",
    \"lint\": \"eslint src --ext .ts\"
  },
  \"dependencies\": {
    \"fastify\": \"^4.26.0\",
    \"@fastify/static\": \"^7.0.0\",
    \"@fastify/type-provider-zod\": \"^4.0.1\",
    \"zod\": \"^3.22.4\",
    \"dotenv\": \"^16.4.0\"
  },
  \"devDependencies\": {
    \"@types/node\": \"^20.11.0\",
    \"typescript\": \"^5.3.3\",
    \"tsx\": \"^4.7.0\",
    \"nodemon\": \"^3.0.3\",
    \"prettier\": \"^3.2.4\",
    \"eslint\": \"^8.56.0\",
    \"@typescript-eslint/parser\": \"^6.19.0\",
    \"@typescript-eslint/eslint-plugin\": \"^6.19.0\"
  }
}"
  }
}
```

### Step 3: Create tsconfig.json

Configure TypeScript for modern Node.js:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/tsconfig.json",
    "content": "{
  \"compilerOptions\": {
    \"target\": \"ES2022\",
    \"module\": \"Node16\",
    \"moduleResolution\": \"Node16\",
    \"lib\": [\"ES2022\"],
    \"outDir\": \"./dist\",
    \"rootDir\": \"./src\",
    \"strict\": true,
    \"esModuleInterop\": true,
    \"skipLibCheck\": true,
    \"forceConsistentCasingInFileNames\": true,
    \"resolveJsonModule\": true
  },
  \"include\": [\"src/**/*\"],
  \"exclude\": [\"node_modules\", \"dist\"]
}"
  }
}
```

### Step 4: Create .env File

Set up environment variables:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/.env",
    "content": "PORT=3000\nNODE_ENV=development"
  }
}
```

### Step 5: Create .gitignore

Ignore unnecessary files:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/.gitignore",
    "content": "node_modules/\ndist/\n.env\n*.log\n.DS_Store"
  }
}
```

### Step 6: Create .prettierrc

Configure code formatting:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/.prettierrc",
    "content": "{
  \"semi\": true,
  \"trailingComma\": \"es5\",
  \"singleQuote\": true,
  \"printWidth\": 100,
  \"tabWidth\": 2
}"
  }
}
```

### Step 7: Create .eslintrc.js

Configure linting:

```json
{
  "operation": "create_file",
  "params": {
    "path": "my-app/.eslintrc.js",
    "content": "module.exports = {
  parser: '@typescript-eslint/parser',
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
  ],
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
  },
  env: {
    node: true,
    es6: true,
  },
  rules: {
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
  },
};"
  }
}
```

### Step 8: Create src/index.ts (Server Entry Point)

Set up Fastify server with Zod validation:

**IMPORTANT: dotenv MUST be imported first, before any other imports that use process.env**

```typescript
// Load environment variables FIRST before any other imports
import 'dotenv/config';

import Fastify from 'fastify';
import fastifyStatic from '@fastify/static';
import { serializerCompiler, validatorCompiler, ZodTypeProvider } from '@fastify/type-provider-zod';
import path from 'path';
import { fileURLToPath } from 'url';

// ES modules __dirname equivalent
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Create Fastify instance with Zod type provider
const server = Fastify({
  logger: {
    level: process.env.NODE_ENV === 'production' ? 'info' : 'debug',
  },
}).withTypeProvider<ZodTypeProvider>();

// Set Zod validators
server.setValidatorCompiler(validatorCompiler);
server.setSerializerCompiler(serializerCompiler);

// Serve static files
server.register(fastifyStatic, {
  root: path.join(__dirname, 'public'),
  prefix: '/public/',
});

// Register routes
import indexRoutes from './routes/index.js';
import apiRoutes from './routes/api.js';

server.register(indexRoutes);
server.register(apiRoutes);

// Health check
server.get('/health', async () => {
  return { status: 'ok', timestamp: new Date().toISOString() };
});

// Start server
const start = async () => {
  try {
    const port = parseInt(process.env.PORT || '3000', 10);
    await server.listen({ port, host: '0.0.0.0' });
    console.log(`Server running at http://localhost:${port}`);
  } catch (err) {
    server.log.error(err);
    process.exit(1);
  }
};

start();
```

### Step 9: Create src/routes/index.ts (Homepage Route)

Serve the main HTML page:

```typescript
import { FastifyPluginAsync } from 'fastify';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const indexRoutes: FastifyPluginAsync = async (server) => {
  server.get('/', async (request, reply) => {
    return reply.sendFile('index.html', path.join(__dirname, '../public'));
  });
};

export default indexRoutes;
```

### Step 10: Create src/routes/api.ts (API Endpoints)

Example CRUD API with Zod validation:

```typescript
import { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import { CreateItemSchema, UpdateItemSchema } from '../lib/validation.js';

// In-memory store (replace with database)
const items: Array<{ id: string; name: string; description: string }> = [];

const apiRoutes: FastifyPluginAsync = async (server) => {
  // List all items
  server.get('/api/items', async () => {
    return { items };
  });

  // Get single item
  server.get<{ Params: { id: string } }>(
    '/api/items/:id',
    {
      schema: {
        params: z.object({ id: z.string() }),
      },
    },
    async (request, reply) => {
      const item = items.find((i) => i.id === request.params.id);
      if (!item) {
        return reply.status(404).send({ error: 'Item not found' });
      }
      return { item };
    }
  );

  // Create item
  server.post<{ Body: z.infer<typeof CreateItemSchema> }>(
    '/api/items',
    {
      schema: {
        body: CreateItemSchema,
      },
    },
    async (request, reply) => {
      const newItem = {
        id: Math.random().toString(36).substr(2, 9),
        ...request.body,
      };
      items.push(newItem);
      return reply.status(201).send({ item: newItem });
    }
  );

  // Update item
  server.put<{ Params: { id: string }; Body: z.infer<typeof UpdateItemSchema> }>(
    '/api/items/:id',
    {
      schema: {
        params: z.object({ id: z.string() }),
        body: UpdateItemSchema,
      },
    },
    async (request, reply) => {
      const index = items.findIndex((i) => i.id === request.params.id);
      if (index === -1) {
        return reply.status(404).send({ error: 'Item not found' });
      }
      items[index] = { ...items[index], ...request.body };
      return { item: items[index] };
    }
  );

  // Delete item
  server.delete<{ Params: { id: string } }>(
    '/api/items/:id',
    {
      schema: {
        params: z.object({ id: z.string() }),
      },
    },
    async (request, reply) => {
      const index = items.findIndex((i) => i.id === request.params.id);
      if (index === -1) {
        return reply.status(404).send({ error: 'Item not found' });
      }
      items.splice(index, 1);
      return reply.status(204).send();
    }
  );
};

export default apiRoutes;
```

### Step 11: Create src/lib/validation.ts (Zod Schemas)

Define validation schemas:

```typescript
import { z } from 'zod';

export const CreateItemSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
});

export const UpdateItemSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
});

// Export inferred types
export type CreateItem = z.infer<typeof CreateItemSchema>;
export type UpdateItem = z.infer<typeof UpdateItemSchema>;
```

### Step 12: Create src/types/index.ts (TypeScript Types)

Define shared types:

```typescript
export interface Item {
  id: string;
  name: string;
  description?: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
}
```

### Step 13: Create src/public/index.html (Frontend)

Create the main HTML page (adapt based on app requirements):

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>My App</title>
  <link rel="stylesheet" href="/public/styles.css">
</head>
<body>
  <div class="container">
    <h1>My Application</h1>
    <p>Welcome to your new web application!</p>

    <!-- Example: Item List -->
    <div class="items">
      <h2>Items</h2>
      <div id="item-list"></div>

      <form id="create-form">
        <input type="text" id="name" placeholder="Item name" required>
        <input type="text" id="description" placeholder="Description (optional)">
        <button type="submit">Add Item</button>
      </form>
    </div>
  </div>

  <script>
    // Fetch and display items
    async function loadItems() {
      const response = await fetch('/api/items');
      const data = await response.json();
      const list = document.getElementById('item-list');
      list.innerHTML = data.items.map(item => `
        <div class="item">
          <h3>${item.name}</h3>
          <p>${item.description || ''}</p>
          <button onclick="deleteItem('${item.id}')">Delete</button>
        </div>
      `).join('');
    }

    // Create new item
    document.getElementById('create-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const name = document.getElementById('name').value;
      const description = document.getElementById('description').value;

      await fetch('/api/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, description }),
      });

      document.getElementById('create-form').reset();
      loadItems();
    });

    // Delete item
    async function deleteItem(id) {
      await fetch(`/api/items/${id}`, { method: 'DELETE' });
      loadItems();
    }

    // Load items on page load
    loadItems();
  </script>
</body>
</html>
```

### Step 14: Create src/public/styles.css (Styles)

Basic styling (adapt to design requirements):

```css
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  line-height: 1.6;
  color: #333;
  background: #f4f4f4;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

h1 {
  color: #ff4f00;
  margin-bottom: 1rem;
}

.items {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  margin-top: 2rem;
}

#item-list {
  margin-bottom: 2rem;
}

.item {
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 1rem;
}

form {
  display: flex;
  gap: 1rem;
}

input {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  flex: 1;
}

button {
  padding: 0.5rem 1rem;
  background: #ff4f00;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background: #e64500;
}
```

### Step 15: Install Dependencies

Install all npm packages:

```json
{
  "operation": "run_command",
  "params": {
    "command": "cd my-app && npm install",
    "cwd": "my-app"
  }
}
```

### Step 16: Start Development Server

Run the app in development mode:

```json
{
  "operation": "start_process",
  "params": {
    "path": "my-app",
    "port": 3000
  }
}
```

## Frontend Adaptation Guide

The template above provides a basic CRUD interface. Adapt based on requirements:

### For Simple Static Site:
- Remove API routes
- Use only static HTML/CSS
- Add navigation between pages

### For Dynamic Dashboard:
- Add HTMX for dynamic updates: `<script src="https://unpkg.com/htmx.org"></script>`
- Use `hx-get`, `hx-post` attributes for interactivity
- No JavaScript required

### For Rich Interactivity:
- Add Alpine.js: `<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3"></script>`
- Use `x-data`, `x-show`, `x-model` for reactive UI
- Minimal JavaScript bundle

### For Real-Time Features:
- Use `@fastify/websocket` plugin
- Add WebSocket routes
- Update frontend to handle WebSocket messages

## Next Steps

### Add Database:
Use the `database_tiger` template to add PostgreSQL/TimescaleDB:
```
discover_patterns("database")
get_template("database_tiger")
```

### Add Authentication:
Use the `auth_jwt` template to add user authentication:
```
discover_patterns("authentication")
get_template("auth_jwt")
```

### Add Payments:
Use the `payments_stripe` template to add Stripe:
```
discover_patterns("payments")
get_template("payments_stripe")
```

### Deploy to Production:
Use the `deploy_cloudflare` template to deploy:
```
discover_patterns("deployment")
get_template("deploy_cloudflare")
```

## Troubleshooting

### Port Already in Use:
```bash
# Find and kill process using port 3000
lsof -ti:3000 | xargs kill -9
```

### TypeScript Errors:
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Hot Reload Not Working:
- Check nodemon is installed: `npm list nodemon`
- Verify file watching: `nodemon --watch src --exec tsx src/index.ts`

## Best Practices

1. **Environment Variables:** Never commit `.env` - use `.env.example` for reference
2. **Error Handling:** Add try-catch blocks and proper error responses
3. **Logging:** Use Fastify's built-in logger for debugging
4. **Validation:** Always validate user input with Zod
5. **Security:** Use helmet plugin for security headers: `@fastify/helmet`
6. **CORS:** Add CORS if frontend is on different domain: `@fastify/cors`

## Additional Resources

- [Fastify Documentation](https://www.fastify.io/)
- [Zod Documentation](https://zod.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Node.js Best Practices](https://github.com/goldbergyoni/nodebestpractices)
