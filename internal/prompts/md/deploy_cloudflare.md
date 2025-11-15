---
title: Deploy to Cloudflare Pages
description: Deploy static sites and serverless functions to Cloudflare Pages with custom domains and environment variables
tags: [deploy, cloudflare, static, frontend, serverless-functions, edge, cdn]
category: deployment
dependencies: []
related: [create_web_app, deploy_railway]
---

# Deploy to Cloudflare Pages

Deploy static sites and serverless functions to Cloudflare Pages with global CDN distribution, custom domains, and environment variables.

## Overview

This guide covers:
- Static site deployment (HTML/CSS/JS)
- Build process setup (Vite, TypeScript)
- Cloudflare Pages Functions (serverless API)
- Custom domains and SSL
- Environment variables
- Production optimization

**Use Cloudflare Pages for:**
- ✅ Static websites and landing pages
- ✅ Frontend applications (React, Vue, vanilla JS)
- ✅ Serverless API endpoints
- ✅ Edge-rendered content

**Use Railway instead for:**
- ❌ Long-running Node.js servers (like Fastify from `create_web_app`)
- ❌ WebSocket servers
- ❌ Background jobs and workers
- ❌ Database-connected backend applications

See `deploy_railway` template for backend server deployment.

## Prerequisites

- GitHub/GitLab/Bitbucket account
- Cloudflare account (free tier available)
- Git repository with your project

## Step 1: Prepare Your Project

### For Static HTML/CSS/JS

No build step needed. Your files are ready to deploy as-is.

Project structure:
```
my-site/
├── index.html
├── style.css
├── script.js
└── assets/
    └── images/
```

### For TypeScript/Vite Projects

Install Vite and TypeScript:

```bash
npm install -D vite typescript
```

Create `vite.config.ts`:

```typescript
import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'esbuild',
  },
});
```

Add build script to `package.json`:

```json
{
  "scripts": {
    "build": "vite build",
    "preview": "vite preview"
  }
}
```

Test build locally:

```bash
npm run build
npm run preview
```

This creates a `dist/` directory with optimized production files.

## Step 2: Create Cloudflare Pages Project

### Option A: Automatic Git Deployment (Recommended)

1. Go to https://dash.cloudflare.com
2. Click **Pages** in sidebar
3. Click **Create a project**
4. Click **Connect to Git**
5. Authorize Cloudflare to access your repository
6. Select your repository
7. Configure build settings:

**For static HTML/CSS/JS:**
- Build command: (leave empty)
- Build output directory: `/`

**For Vite projects:**
- Build command: `npm run build`
- Build output directory: `dist`

**For TypeScript (no bundler):**
- Build command: `tsc`
- Build output directory: `dist` or `build`

8. Click **Save and Deploy**

Cloudflare will build and deploy your site. Future git pushes automatically trigger deployments.

### Option B: Direct Upload (Wrangler CLI)

Install Wrangler:

```bash
npm install -g wrangler
```

Login to Cloudflare:

```bash
wrangler login
```

Deploy static files:

```bash
# For static HTML/CSS/JS
wrangler pages deploy ./

# For built projects
npm run build
wrangler pages deploy ./dist
```

## Step 3: Configure Environment Variables

### In Cloudflare Dashboard

1. Go to your Pages project
2. Click **Settings** → **Environment variables**
3. Add variables:

```
VITE_API_URL=https://api.yourdomain.com
VITE_STRIPE_PUBLISHABLE_KEY=pk_live_...
```

**Important:** Only `VITE_` prefixed variables are exposed to client-side code in Vite projects.

### In Code

Access environment variables:

```typescript
// Vite projects
const apiUrl = import.meta.env.VITE_API_URL;

// Vanilla JS (set during build)
const apiUrl = process.env.VITE_API_URL;
```

### Build-time vs Runtime

**Build-time variables** (Vite):
- Embedded in built files during deployment
- Cannot change without rebuild
- Use for API URLs, public keys

**Runtime variables** (Cloudflare Pages Functions):
- Available in serverless functions
- Can change without rebuild
- Use for secrets in server-side code

## Step 4: Custom Domain Setup

### Add Custom Domain

1. Go to your Pages project
2. Click **Custom domains**
3. Click **Set up a custom domain**
4. Enter your domain: `yourdomain.com`
5. Follow DNS configuration instructions

### DNS Configuration

Add these records to your domain's DNS:

**For root domain (yourdomain.com):**
```
Type: CNAME
Name: @
Value: your-project.pages.dev
Proxy: Enabled (orange cloud)
```

**For subdomain (www.yourdomain.com):**
```
Type: CNAME
Name: www
Value: your-project.pages.dev
Proxy: Enabled (orange cloud)
```

**SSL/TLS:** Cloudflare provides free SSL certificates automatically.

## Step 5: Cloudflare Pages Functions (Serverless API)

Add serverless functions to your static site.

### Project Structure

```
my-site/
├── functions/          # Serverless functions
│   ├── api/
│   │   └── hello.ts   # /api/hello endpoint
│   └── contact.ts     # /contact endpoint
├── public/            # Static files
│   ├── index.html
│   └── style.css
└── package.json
```

### Simple Function Example

Create `functions/api/hello.ts`:

```typescript
export async function onRequest(context) {
  return new Response(JSON.stringify({
    message: 'Hello from Cloudflare Pages!',
    timestamp: new Date().toISOString(),
  }), {
    headers: {
      'Content-Type': 'application/json',
    },
  });
}
```

**URL:** `https://yourdomain.com/api/hello`

### Function with Request Data

Create `functions/api/contact.ts`:

```typescript
interface Env {
  // Environment variables available in functions
  RESEND_API_KEY: string;
}

export async function onRequestPost(context: { request: Request; env: Env }) {
  try {
    const { name, email, message } = await context.request.json();

    // Validate input
    if (!name || !email || !message) {
      return new Response(JSON.stringify({
        error: 'Missing required fields'
      }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    // Send email (example with Resend)
    const response = await fetch('https://api.resend.com/emails', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${context.env.RESEND_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        from: 'noreply@yourdomain.com',
        to: 'contact@yourdomain.com',
        subject: `Contact form: ${name}`,
        text: `From: ${name} (${email})\n\n${message}`,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to send email');
    }

    return new Response(JSON.stringify({
      success: true,
      message: 'Message sent successfully'
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    return new Response(JSON.stringify({
      error: 'Failed to send message'
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}
```

### CORS for API Functions

Add CORS headers for cross-origin requests:

```typescript
export async function onRequest(context) {
  // Handle OPTIONS preflight
  if (context.request.method === 'OPTIONS') {
    return new Response(null, {
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    });
  }

  // Your function logic
  const response = new Response(JSON.stringify({ data: 'example' }), {
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
    },
  });

  return response;
}
```

### Function Limitations

- **Execution time:** 10ms CPU time (free), 50ms (paid)
- **Request size:** 100MB
- **Response size:** 25MB
- **No persistent connections:** Each request is isolated

For complex backends, use `deploy_railway` instead.

## Step 6: Frontend Code Example

Call your Cloudflare Functions from frontend:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Contact Form</title>
</head>
<body>
  <form id="contact-form">
    <input type="text" id="name" placeholder="Name" required>
    <input type="email" id="email" placeholder="Email" required>
    <textarea id="message" placeholder="Message" required></textarea>
    <button type="submit">Send</button>
  </form>

  <script>
    document.getElementById('contact-form').addEventListener('submit', async (e) => {
      e.preventDefault();

      const data = {
        name: document.getElementById('name').value,
        email: document.getElementById('email').value,
        message: document.getElementById('message').value,
      };

      try {
        const response = await fetch('/api/contact', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (response.ok) {
          alert('Message sent successfully!');
          e.target.reset();
        } else {
          alert(`Error: ${result.error}`);
        }
      } catch (error) {
        alert('Failed to send message. Please try again.');
      }
    });
  </script>
</body>
</html>
```

## Step 7: Production Optimization

### Performance Best Practices

**Minify assets:**
```javascript
// vite.config.ts
export default defineConfig({
  build: {
    minify: 'esbuild', // or 'terser' for smaller files
    cssMinify: true,
  },
});
```

**Image optimization:**
- Use modern formats (WebP, AVIF)
- Compress images before upload
- Use Cloudflare Images (paid) for automatic optimization

**Caching:**
Cloudflare automatically caches static assets. Configure cache headers:

Create `_headers` file in your output directory:

```
/*
  Cache-Control: public, max-age=31536000, immutable

/*.html
  Cache-Control: public, max-age=0, must-revalidate

/api/*
  Cache-Control: no-cache
```

### Security Headers

Create `_headers` file:

```
/*
  X-Frame-Options: DENY
  X-Content-Type-Options: nosniff
  X-XSS-Protection: 1; mode=block
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Redirects

Create `_redirects` file:

```
# Redirect old paths
/old-page /new-page 301

# SPA fallback (all routes to index.html)
/* /index.html 200

# Redirect www to non-www
https://www.yourdomain.com/* https://yourdomain.com/:splat 301
```

## Step 8: CI/CD with Git

Every git push automatically triggers a deployment:

```bash
git add .
git commit -m "Update homepage"
git push origin main
```

Cloudflare builds and deploys automatically.

### Branch Previews

Every branch gets a preview URL:
- `main` branch → `yourdomain.com`
- `develop` branch → `develop.your-project.pages.dev`
- Pull requests → `pr-123.your-project.pages.dev`

Great for testing before merging.

## Step 9: Monitoring and Logs

### View Deployment Logs

1. Go to your Pages project
2. Click **Deployments**
3. Click on a deployment
4. View build and function logs

### Analytics

1. Click **Analytics** in your project
2. View:
   - Page views
   - Requests
   - Bandwidth usage
   - Function invocations

### Function Logs

Add logging to functions:

```typescript
export async function onRequest(context) {
  console.log('Function invoked:', context.request.url);

  // Your logic...

  return new Response('OK');
}
```

View logs in deployment details.

## Common Patterns

### Environment-Specific Configuration

```typescript
const config = {
  apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:3000',
  stripeKey: import.meta.env.VITE_STRIPE_KEY,
};

// Use config.apiUrl in your code
```

### API Proxy (Avoid CORS)

Create `functions/api/proxy/[...path].ts`:

```typescript
export async function onRequest(context) {
  const url = new URL(context.request.url);
  const path = context.params.path.join('/');

  // Proxy to your backend
  const response = await fetch(`https://your-backend.railway.app/${path}${url.search}`, {
    method: context.request.method,
    headers: context.request.headers,
    body: context.request.body,
  });

  return response;
}
```

Now `/api/proxy/users` → `https://your-backend.railway.app/users`

## Troubleshooting

### Build Failures

**Check build logs** in deployment details:
- Missing dependencies? Add to `package.json`
- Build command incorrect? Update in settings
- Environment variables missing? Add in dashboard

### Functions Not Working

- Check function file path matches URL structure
- Verify function exports `onRequest` or `onRequestGet/Post/etc`
- Check environment variables are set
- Review function logs for errors

### Custom Domain Not Working

- Verify DNS records are correct
- Wait for DNS propagation (up to 24 hours)
- Ensure CNAME points to `your-project.pages.dev`
- Check Cloudflare proxy is enabled (orange cloud)

### 404 Errors for SPA Routes

Add `_redirects` file with SPA fallback:
```
/* /index.html 200
```

## Migration from Other Platforms

### From Vercel

- Similar deployment process
- Update environment variables
- Update custom domain DNS
- Update `_redirects` format if needed

### From Netlify

- Similar `_redirects` and `_headers` format
- Functions syntax is different (update to Cloudflare format)
- Environment variables need migration

## Cost Considerations

**Free Tier:**
- Unlimited requests
- Unlimited bandwidth
- 500 builds per month
- 20,000 functions requests per day

**Paid Tier ($20/month):**
- 5,000 builds per month
- 100,000 functions requests per day
- Advanced features

Most projects fit in free tier.

## When to Use Cloudflare vs Railway

**Use Cloudflare Pages when:**
- ✅ Static site or SPA (React, Vue, etc.)
- ✅ Landing pages and marketing sites
- ✅ Simple serverless APIs
- ✅ Edge-rendered content
- ✅ Global CDN distribution priority

**Use Railway when:**
- ✅ Node.js server (Fastify, Express)
- ✅ Database connections (from `database_tiger`)
- ✅ WebSocket servers
- ✅ Background jobs
- ✅ Long-running processes
- ✅ Complex backend logic

**Use Both Together:**
- Frontend on Cloudflare Pages (fast global delivery)
- Backend on Railway (full server capabilities)
- Connect via API calls

## Next Steps

- **Add analytics:** Cloudflare Web Analytics
- **Add WAF:** Web Application Firewall for security
- **Add Workers:** More powerful edge computing
- **Add R2:** Object storage for files
- **Deploy backend:** Use `deploy_railway` for server deployment

## Useful Resources

- Cloudflare Pages Docs: https://developers.cloudflare.com/pages
- Pages Functions: https://developers.cloudflare.com/pages/functions
- Wrangler CLI: https://developers.cloudflare.com/workers/wrangler
- Examples: https://github.com/cloudflare/pages-example-projects
