---
title: Deploy to Railway
description: Deploy Node.js servers and long-running applications to Railway with environment variables and custom domains
tags: [deploy, railway, server, backend, fastify, nodejs, long-running, websockets]
category: deployment
dependencies: []
related: [create_web_app, database_tiger, deploy_cloudflare]
---

# Deploy to Railway

Deploy Node.js servers and long-running applications to Railway with automatic deployments, environment variables, and custom domains.

## Overview

This guide covers:
- Deploy Fastify servers (from `create_web_app`)
- Git-based automatic deployments
- Railway CLI for manual deploys
- Environment variables configuration
- Custom domain setup
- Monitoring and logs
- Production optimization

**Use Railway for:**
- ✅ Node.js servers (Fastify, Express, etc.)
- ✅ Long-running processes
- ✅ WebSocket servers
- ✅ Background jobs and workers
- ✅ Database-connected applications
- ✅ Any server that needs to stay running

**Use Cloudflare Pages instead for:**
- ❌ Static sites (HTML/CSS/JS only)
- ❌ Simple serverless functions

See `deploy_cloudflare` template for static site deployment.

## Prerequisites

- GitHub/GitLab account
- Railway account (sign up at https://railway.app)
- Git repository with your Node.js application
- Application from `create_web_app` template (or similar Node.js server)

## Step 1: Prepare Your Application

### Verify package.json Scripts

Ensure your `package.json` has a start script:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js"
  },
  "dependencies": {
    "fastify": "^4.x.x",
    "dotenv": "^16.x.x"
  }
}
```

**Important:** Railway runs `npm run build` then `npm start`.

### Configure Port Binding

Railway provides `PORT` environment variable. Update your server to use it:

```typescript
// src/index.ts or src/server.ts
import Fastify from 'fastify';
import 'dotenv/config';

const server = Fastify({ logger: true });

// Your routes...

const PORT = parseInt(process.env.PORT || '3000', 10);
const HOST = '0.0.0.0'; // Important: bind to 0.0.0.0, not localhost

async function start() {
  try {
    await server.listen({ port: PORT, host: HOST });
    console.log(`Server listening on ${HOST}:${PORT}`);
  } catch (error) {
    server.log.error(error);
    process.exit(1);
  }
}

start();
```

**Critical:** Use `0.0.0.0` as host, not `localhost` or `127.0.0.1`. Railway needs to access from external network.

### Test Locally

```bash
npm run build
PORT=3000 npm start
```

Visit `http://localhost:3000` to verify it works.

## Step 2: Create Railway Project

### Option A: Deploy from GitHub (Recommended)

1. Go to https://railway.app
2. Click **New Project**
3. Select **Deploy from GitHub repo**
4. Authorize Railway to access your repositories
5. Select your repository
6. Railway automatically:
   - Detects it's a Node.js app
   - Runs `npm install`
   - Runs `npm run build` (if build script exists)
   - Runs `npm start`
   - Assigns a public URL

Your app is now deployed! 🎉

### Option B: Deploy with Railway CLI

Install Railway CLI:

```bash
npm install -g @railway/cli
```

Login:

```bash
railway login
```

Initialize project:

```bash
cd your-app-directory
railway init
```

Select "Create new project" and follow prompts.

Deploy:

```bash
railway up
```

## Step 3: Configure Environment Variables

### Required Variables

Your app needs environment variables (from previous templates):

```bash
# Database (from database_tiger)
DATABASE_URL="postgres://tsdbadmin:password@xxxxx.tsdb.cloud.timescale.com:12345/tsdb"

# JWT Secrets (from auth_jwt)
JWT_ACCESS_SECRET="your-access-token-secret"
JWT_REFRESH_SECRET="your-refresh-token-secret"
JWT_ACCESS_EXPIRES_IN="15m"
JWT_REFRESH_EXPIRES_IN="7d"

# Email (from email_resend)
RESEND_API_KEY="re_your_api_key"
FROM_EMAIL="noreply@yourdomain.com"
FROM_NAME="Your App"

# Stripe (from payments_stripe)
STRIPE_SECRET_KEY="sk_live_your_secret_key"
STRIPE_WEBHOOK_SECRET="whsec_your_webhook_secret"

# App Configuration
NODE_ENV="production"
APP_URL="https://your-app.up.railway.app"
```

### Add Variables via Dashboard

1. Go to your Railway project
2. Click on your service
3. Click **Variables** tab
4. Click **New Variable**
5. Add each variable (name and value)
6. Click **Save**

Railway automatically restarts your service with new variables.

### Add Variables via CLI

```bash
railway variables set DATABASE_URL="postgres://..."
railway variables set JWT_ACCESS_SECRET="your-secret"
railway variables set STRIPE_SECRET_KEY="sk_live_..."
```

View all variables:

```bash
railway variables
```

### Variable Groups

For multiple environments (staging, production):

1. Create separate Railway services
2. Set different variables for each
3. Deploy same code to both

## Step 4: Custom Domain Setup

### Add Domain

1. Go to your Railway service
2. Click **Settings** tab
3. Scroll to **Domains**
4. Click **Generate Domain** (gets you a `*.up.railway.app` domain)
5. Or click **Custom Domain** to add your own

### Configure DNS for Custom Domain

Add a CNAME record to your domain's DNS:

```
Type: CNAME
Name: api (or @ for root domain)
Value: your-app.up.railway.app
TTL: Auto or 3600
```

**Note:** Railway handles SSL certificates automatically via Let's Encrypt.

### Update APP_URL Variable

After adding custom domain, update `APP_URL`:

```bash
railway variables set APP_URL="https://api.yourdomain.com"
```

## Step 5: Automatic Deployments

### Git Integration

Every push to your main branch automatically triggers a deployment:

```bash
git add .
git commit -m "Add new feature"
git push origin main
```

Railway:
1. Detects the push
2. Pulls latest code
3. Runs `npm install`
4. Runs `npm run build`
5. Runs `npm start`
6. Routes traffic to new deployment

### Branch Deployments

Deploy feature branches for testing:

1. Create branch: `git checkout -b feature/new-api`
2. Push: `git push origin feature/new-api`
3. In Railway, click **New Service**
4. Select same repo, different branch
5. Configure to deploy from `feature/new-api`

Each branch gets its own URL.

## Step 6: Monitoring and Logs

### View Logs

**Via Dashboard:**
1. Go to your service
2. Click **Deployments** tab
3. Click on active deployment
4. View real-time logs

**Via CLI:**
```bash
railway logs
```

Follow logs in real-time:
```bash
railway logs --follow
```

### Metrics

Railway dashboard shows:
- CPU usage
- Memory usage
- Network traffic
- Request count
- Deployment history

### Health Checks

Add health check endpoint:

```typescript
// In your Fastify server
server.get('/health', async (request, reply) => {
  // Check database connection
  try {
    await db.execute('SELECT 1');
    return {
      status: 'healthy',
      timestamp: new Date().toISOString(),
      database: 'connected',
    };
  } catch (error) {
    reply.status(503);
    return {
      status: 'unhealthy',
      database: 'disconnected',
    };
  }
});
```

Railway automatically pings this endpoint (if it exists) to verify service health.

## Step 7: Database Connection

### Using Tiger Cloud Database

Railway works seamlessly with external databases like Tiger Cloud:

```typescript
// src/db/client.ts (from database_tiger template)
import postgres from 'postgres';
import { drizzle } from 'drizzle-orm/postgres-js';

const connectionString = process.env.DATABASE_URL;

if (!connectionString) {
  throw new Error('DATABASE_URL environment variable is required');
}

export const client = postgres(connectionString, {
  max: 10, // Connection pool size
  idle_timeout: 20,
  connect_timeout: 10,
});

export const db = drizzle(client);
```

**Important:** Set `DATABASE_URL` in Railway environment variables (from Tiger Cloud connection string).

### Connection Pooling

Use connection pooling for better performance:

```typescript
const client = postgres(connectionString, {
  max: 10, // Maximum 10 connections
  idle_timeout: 20,
  connect_timeout: 10,
});
```

Railway automatically handles connection scaling.

## Step 8: Webhook Configuration

If using webhooks (Stripe, etc.), configure them to point to your Railway URL:

### Stripe Webhooks

1. Go to https://dashboard.stripe.com/webhooks
2. Click **Add endpoint**
3. Enter URL: `https://your-app.up.railway.app/webhooks/stripe`
4. Select events to listen for
5. Copy webhook signing secret
6. Add to Railway: `railway variables set STRIPE_WEBHOOK_SECRET="whsec_..."`

### Testing Webhooks Locally

Use Railway CLI to tunnel webhooks to local development:

```bash
# Terminal 1: Run app locally
npm run dev

# Terminal 2: Forward webhooks
railway run --port 3000
```

This creates a temporary public URL that forwards to your local server.

## Step 9: Production Optimization

### Build Optimization

Optimize TypeScript compilation:

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "outDir": "./dist",
    "rootDir": "./src",
    "removeComments": true,
    "sourceMap": false,
    "declaration": false
  }
}
```

### Environment-Specific Code

```typescript
const isProd = process.env.NODE_ENV === 'production';

const server = Fastify({
  logger: isProd ? true : {
    level: 'debug',
    transport: {
      target: 'pino-pretty'
    }
  }
});
```

### Error Handling

```typescript
// Global error handler
server.setErrorHandler((error, request, reply) => {
  server.log.error(error);

  // Don't expose internal errors in production
  const isProd = process.env.NODE_ENV === 'production';

  reply.status(error.statusCode || 500).send({
    error: isProd ? 'Internal server error' : error.message,
    statusCode: error.statusCode || 500,
  });
});

// Graceful shutdown
process.on('SIGTERM', async () => {
  server.log.info('SIGTERM received, closing server gracefully');
  await server.close();
  process.exit(0);
});
```

### Memory Management

Set Node.js memory limits:

```json
// package.json
{
  "scripts": {
    "start": "node --max-old-space-size=512 dist/index.js"
  }
}
```

## Step 10: Scaling

### Vertical Scaling (More Resources)

1. Go to service **Settings**
2. Scroll to **Resources**
3. Adjust:
   - CPU allocation
   - Memory limit
4. Click **Save**

### Horizontal Scaling (Multiple Instances)

Railway doesn't support automatic horizontal scaling on free tier, but you can:

1. Deploy same app as separate service
2. Use load balancer (external)
3. Or upgrade to Railway Pro for replicas

## CLI Commands Reference

```bash
# Login
railway login

# Initialize project
railway init

# Deploy
railway up

# Set variable
railway variables set KEY="value"

# View logs
railway logs
railway logs --follow

# Open dashboard
railway open

# Link to existing project
railway link

# Run command with Railway environment
railway run npm start

# SSH into service
railway shell
```

## Troubleshooting

### Build Fails

**Check build logs:**
- Missing dependencies? Run `npm install` locally
- TypeScript errors? Run `npm run build` locally
- Wrong Node version? Add `.nvmrc` or set in Railway settings

### App Crashes on Start

**Common issues:**
- Port binding: Use `0.0.0.0`, not `localhost`
- Missing environment variables: Check Railway variables
- Database connection fails: Verify `DATABASE_URL`

### 502 Bad Gateway

- App not listening on correct PORT
- App crashed during startup (check logs)
- Health check failing

### Environment Variables Not Working

- Variables set but not saving? Click **Save** button
- App not seeing variables? Restart service
- Still not working? Redeploy service

### Database Connection Fails

- Verify `DATABASE_URL` is correct
- Check Tiger Cloud firewall/IP restrictions
- Test connection locally first
- Verify SSL/TLS settings if required

## Cost Considerations

**Free Tier (Hobby Plan):**
- $5/month credit
- 512MB RAM
- Shared CPU
- 100GB bandwidth
- Unlimited projects

**Paid Tier (Developer/Team):**
- More resources
- Priority support
- Multiple replicas
- Advanced features

Most small apps fit in free tier.

## Deployment Checklist

- [ ] `package.json` has `start` script
- [ ] Server binds to `0.0.0.0:${PORT}`
- [ ] All environment variables set in Railway
- [ ] Build succeeds locally (`npm run build`)
- [ ] Health check endpoint added
- [ ] Error handling implemented
- [ ] Logs are readable (structured logging)
- [ ] Database connection string added
- [ ] Stripe webhook URL updated (if using)
- [ ] Custom domain configured (optional)
- [ ] SSL certificate verified
- [ ] Monitoring set up

## Common Patterns

### Multi-Service Architecture

Deploy multiple services:

**Frontend (Cloudflare Pages):**
- Static site or SPA
- Fast global delivery

**Backend (Railway):**
- API server
- Database connections
- Business logic

**Workers (Railway):**
- Background jobs
- Scheduled tasks
- Queue processing

### Environment-Based Deployments

**Staging:**
```bash
railway service staging
railway variables set NODE_ENV="staging"
railway variables set DATABASE_URL="staging-db-url"
```

**Production:**
```bash
railway service production
railway variables set NODE_ENV="production"
railway variables set DATABASE_URL="prod-db-url"
```

## Migration from Other Platforms

### From Heroku

Railway is Heroku-like:
- Same `Procfile` support (optional)
- Same environment variable pattern
- Similar CLI commands
- Often just works with existing setup

### From Vercel/Netlify

If you had serverless functions, convert to regular Express/Fastify routes:

```typescript
// Before (Vercel serverless)
// api/hello.ts
export default function handler(req, res) {
  res.json({ message: 'Hello' });
}

// After (Railway server)
// src/server.ts
server.get('/api/hello', async (request, reply) => {
  return { message: 'Hello' };
});
```

## When to Use Railway vs Cloudflare

**Railway for:**
- ✅ Backend APIs (Fastify, Express)
- ✅ Database-connected apps
- ✅ Long-running servers
- ✅ WebSockets
- ✅ Background workers
- ✅ Traditional Node.js apps

**Cloudflare Pages for:**
- ✅ Static sites
- ✅ SPAs (React, Vue, etc.)
- ✅ Simple serverless functions
- ✅ Edge-rendered content
- ✅ Marketing sites

**Use Both:**
- Frontend on Cloudflare (fast CDN)
- Backend on Railway (full server)
- Connect via API calls

## Next Steps

- **Add monitoring:** Integrate with Sentry, LogRocket, etc.
- **Set up CI/CD:** GitHub Actions for tests before deploy
- **Add caching:** Redis on Railway (separate service)
- **Background jobs:** Add Bull/BullMQ for job queue
- **WebSockets:** Add Socket.io or native WebSockets

## Useful Resources

- Railway Documentation: https://docs.railway.app
- Railway CLI: https://docs.railway.app/develop/cli
- Railway Templates: https://railway.app/templates
- Discord Community: https://discord.gg/railway
