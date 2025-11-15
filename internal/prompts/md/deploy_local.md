---
title: Deploy Locally
description: Run your application locally for development and testing with hot reloading and log monitoring
tags: [deploy, local, development, localhost, testing, dev-server]
category: deployment
dependencies: []
related: [create_web_app, deploy_railway, deploy_cloudflare]
---

# Deploy Locally

Run your Node.js application locally for development and testing with automatic restarts, log monitoring, and easy debugging.

## Overview

This guide covers:
- Starting your application locally
- Port configuration
- Monitoring logs
- Stopping and restarting
- Managing multiple local services
- Environment variables for local development

**Use local deployment for:**
- ✅ Development and testing
- ✅ Debugging before production
- ✅ Rapid iteration
- ✅ Offline work
- ✅ Testing integrations

**Use Railway/Cloudflare instead for:**
- ❌ Production deployments
- ❌ Public access
- ❌ Continuous deployment

## Prerequisites

- Node.js application (from `create_web_app`)
- Dependencies installed (`npm install`)
- Environment variables configured (`.env` file)

## Step 1: Verify Application Setup

Ensure your `package.json` has a dev script:

```json
{
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js"
  }
}
```

**Development mode (`npm run dev`):**
- Uses `tsx watch` for TypeScript execution
- Automatically restarts on file changes
- Faster iteration, no build step needed

**Production mode (`npm start`):**
- Requires building first (`npm run build`)
- Runs compiled JavaScript
- Better performance, but no hot reloading

## Step 2: Configure Environment Variables

Create `.env` file in project root:

```bash
# Server Configuration
PORT=3000
NODE_ENV=development

# Database (Tiger Cloud works locally too)
DATABASE_URL="postgres://tsdbadmin:password@xxxxx.tsdb.cloud.timescale.com:12345/tsdb"

# JWT Secrets (use different secrets than production)
JWT_ACCESS_SECRET="dev-access-secret-change-in-production"
JWT_REFRESH_SECRET="dev-refresh-secret-change-in-production"

# Email (use test mode or dev API keys)
RESEND_API_KEY="re_test_key"
FROM_EMAIL="dev@yourdomain.com"

# Stripe (use test keys)
STRIPE_SECRET_KEY="sk_test_your_test_key"
STRIPE_PUBLISHABLE_KEY="pk_test_your_test_key"

# App URL
APP_URL="http://localhost:3000"
```

**Security:** Never commit `.env` to version control. Ensure `.gitignore` includes `.env`.

## Step 3: Start Local Server

### Option A: Using npm (Simple)

```bash
npm run dev
```

The server starts on `http://localhost:3000` (or your configured PORT).

### Option B: Using Execute Primitive (MCP/Programmatic)

Use the `start_process` primitive to start the server:

```json
{
  "operation": "start_process",
  "params": {
    "name": "my-app",
    "command": "npm",
    "args": ["run", "dev"],
    "cwd": "/path/to/your/app",
    "env": {
      "PORT": "3000",
      "NODE_ENV": "development"
    }
  }
}
```

This returns a process ID that you can use to manage the server.

## Step 4: Verify Server is Running

### Check Server Status

Use the `list_processes` primitive:

```json
{
  "operation": "list_processes"
}
```

Returns list of running processes with their IDs, names, and ports.

### Test Endpoints

```bash
# Health check
curl http://localhost:3000/health

# Test API endpoint
curl http://localhost:3000/api/hello

# With authentication
curl http://localhost:3000/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Step 5: Monitor Logs

### View Real-Time Logs

Use the `get_logs` primitive:

```json
{
  "operation": "get_logs",
  "params": {
    "process_id": "abc123",
    "lines": 50
  }
}
```

Returns recent log output from your server.

### Log Output

Your Fastify server logs will show:

```
[timestamp] INFO: Server listening on http://0.0.0.0:3000
[timestamp] INFO: GET /health 200 5ms
[timestamp] INFO: POST /api/users 201 45ms
[timestamp] ERROR: Database connection failed
```

## Step 6: Port Configuration

### Default Port

By default, apps use port 3000. To change:

**Via environment variable:**
```bash
PORT=4000 npm run dev
```

**Via .env file:**
```bash
PORT=4000
```

### Multiple Services

Run multiple services on different ports:

| Service | Port |
|---------|------|
| Main app | 3000 |
| Admin panel | 3001 |
| Worker | 3002 |

Start each with different PORT values:

```bash
# Terminal 1
PORT=3000 npm run dev

# Terminal 2
cd admin && PORT=3001 npm run dev

# Terminal 3
cd worker && PORT=3002 npm run dev
```

Or use `start_process` primitive with different ports:

```json
{
  "operation": "start_process",
  "params": {
    "name": "main-app",
    "command": "npm",
    "args": ["run", "dev"],
    "cwd": "/path/to/main",
    "env": { "PORT": "3000" }
  }
}
```

```json
{
  "operation": "start_process",
  "params": {
    "name": "admin",
    "command": "npm",
    "args": ["run", "dev"],
    "cwd": "/path/to/admin",
    "env": { "PORT": "3001" }
  }
}
```

## Step 7: Stop Server

### Stop via Terminal

Press `Ctrl+C` in the terminal running the server.

### Stop via Execute Primitive

Use the `stop_process` primitive:

```json
{
  "operation": "stop_process",
  "params": {
    "process_id": "abc123"
  }
}
```

This gracefully shuts down the server.

## Step 8: Restart on Changes

### Automatic Restart (Development)

When using `tsx watch`, the server automatically restarts when you save files:

1. Edit `src/routes/users.ts`
2. Save file
3. Server automatically restarts
4. Changes immediately available

### Manual Restart

Stop the server and start again:

```bash
# Stop (Ctrl+C)
# Start
npm run dev
```

Or use primitives:

```json
// Stop
{
  "operation": "stop_process",
  "params": { "process_id": "abc123" }
}

// Start
{
  "operation": "start_process",
  "params": {
    "name": "my-app",
    "command": "npm",
    "args": ["run", "dev"],
    "cwd": "/path/to/app"
  }
}
```

## Common Workflows

### Development Workflow

1. Start local server: `npm run dev`
2. Make code changes
3. Server auto-restarts
4. Test in browser/Postman
5. View logs in terminal
6. Debug as needed
7. Commit changes

### Testing Before Deploy

1. Start local server
2. Test all endpoints
3. Verify database connections
4. Check auth flows
5. Test payment integration (use Stripe test mode)
6. Review logs for errors
7. Stop local server
8. Deploy to Railway/Cloudflare

### Multi-Service Development

Run multiple services locally:

```bash
# Terminal 1: Main API
cd api && npm run dev

# Terminal 2: Frontend
cd frontend && npm run dev

# Terminal 3: Worker
cd worker && npm run dev

# Terminal 4: Database logs
docker logs -f postgres-container
```

## Environment Variables

### Development vs Production

Use different values for local development:

| Variable | Local | Production |
|----------|-------|------------|
| `NODE_ENV` | development | production |
| `APP_URL` | http://localhost:3000 | https://yourdomain.com |
| `DATABASE_URL` | Local or dev DB | Production DB |
| `STRIPE_SECRET_KEY` | sk_test_... | sk_live_... |
| JWT secrets | Simple for dev | Strong for prod |

### Loading Environment Variables

Your app loads `.env` automatically with `dotenv`:

```typescript
import 'dotenv/config';

const port = process.env.PORT || 3000;
const dbUrl = process.env.DATABASE_URL;
```

## Debugging

### Enable Debug Logging

Set log level in development:

```typescript
const server = Fastify({
  logger: {
    level: 'debug', // Show all logs
    transport: {
      target: 'pino-pretty', // Pretty formatting
      options: {
        colorize: true,
        translateTime: 'HH:MM:ss'
      }
    }
  }
});
```

### Inspect Variables

Add console logs:

```typescript
server.get('/api/users', async (request, reply) => {
  console.log('Request headers:', request.headers);
  console.log('Query params:', request.query);

  const users = await db.select().from(users);
  console.log('Found users:', users.length);

  return users;
});
```

### Debug Database Queries

Enable Drizzle query logging:

```typescript
import { drizzle } from 'drizzle-orm/postgres-js';

export const db = drizzle(client, {
  schema,
  logger: true // Logs all SQL queries
});
```

## Troubleshooting

### Port Already in Use

```
Error: listen EADDRINUSE: address already in use :::3000
```

**Solutions:**
- Use different port: `PORT=3001 npm run dev`
- Kill process using port: `lsof -ti:3000 | xargs kill -9`
- Stop other server on that port

### Cannot Connect to Database

**Check:**
- DATABASE_URL is correct
- Database server is running
- Network/firewall allows connection
- Credentials are valid

**Test connection:**
```bash
psql $DATABASE_URL -c "SELECT 1"
```

### Module Not Found

```
Error: Cannot find module 'fastify'
```

**Solution:**
```bash
npm install
```

### TypeScript Errors

**Check:**
```bash
npx tsc --noEmit
```

Fix TypeScript errors before running.

### Environment Variables Not Loading

**Check:**
- `.env` file exists in project root
- File is not in `.gitignore` location
- Using `dotenv/config` or `dotenv.config()`
- Restart server after changing `.env`

## Process Management

### List Running Processes

```json
{
  "operation": "list_processes"
}
```

Returns:
```json
{
  "processes": [
    {
      "id": "abc123",
      "name": "my-app",
      "port": 3000,
      "status": "running",
      "pid": 12345
    }
  ]
}
```

### Get Process Logs

```json
{
  "operation": "get_logs",
  "params": {
    "process_id": "abc123",
    "lines": 100
  }
}
```

### Stop All Processes

Stop each process individually:

```json
{
  "operation": "stop_process",
  "params": {
    "process_id": "abc123"
  }
}
```

Or use terminal:
```bash
pkill -f "npm run dev"
```

## Testing Endpoints

### Manual Testing

**GET request:**
```bash
curl http://localhost:3000/api/users
```

**POST request:**
```bash
curl -X POST http://localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}'
```

**With authentication:**
```bash
curl http://localhost:3000/profile \
  -H "Authorization: Bearer eyJhbGc..."
```

### Automated Testing

Use the same endpoints in your test suite:

```typescript
import { test } from 'vitest';

test('GET /api/users returns users', async () => {
  const response = await fetch('http://localhost:3000/api/users');
  const users = await response.json();

  expect(response.status).toBe(200);
  expect(Array.isArray(users)).toBe(true);
});
```

## Next Steps

After local testing:
- **Deploy to staging:** Test in production-like environment
- **Deploy to production:** Use `deploy_railway` or `deploy_cloudflare`
- **Set up monitoring:** Add error tracking and logging
- **Configure CI/CD:** Automate testing and deployment

## Useful Resources

- tsx Documentation: https://github.com/privatenumber/tsx
- Fastify Documentation: https://fastify.dev
- Node.js Debugging: https://nodejs.org/en/docs/guides/debugging-getting-started
