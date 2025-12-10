---
name: create-app
description: 'Use this skill whenever creating a new application. IMPORTANT: This should be the FIRST thing you read when starting a new project. IMPORTANT: Read this before planning or brainstorming.'
---

# Create App Implementation Plan

> **For Claude:** Follow this plan task-by-task. If any step fails, notify the user and ask for next steps.

**Goal:** Scaffold a production-ready fullstack web application with database, optional auth, and polished UI.

**Architecture:** T3 stack (Next.js + tRPC + Drizzle) with Timescale Cloud database and shadcn/ui components.

**Tech Stack:** Next.js, tRPC, Drizzle ORM, Timescale Cloud (PostgreSQL), Better Auth, shadcn/ui, Tailwind CSS

---

## Phase 1: Project Setup

### Task 1: Gather Requirements

**Step 1: Determine app type**

If it is not clear from the prompt, ask the user: "Is this a multi-user app (requires user accounts/login)?"

**Step 2: Gather auth requirements (if multi-user)**

Ask the user: "Which authentication methods do you want? Pick one or more:"
- Email signup
- GitHub OAuth
- Google OAuth

**Step 3: Confirm app name**

Propose a sensible app name based on the user's request. The name should be:
- Lowercase
- Use hyphens instead of spaces (e.g., `todo-app`, `fitness-tracker`)
- Appropriate for a directory name

Ask the user: "I'll name the project `<proposed-name>`. Does that work, or would you prefer something else?"

---

### Task 2: Create Database

**Step 1: Create database**

Use the `create_database` MCP tool to provision a new Timescale Cloud database.

**Step 2: Save the service_id**

Store the returned `service_id` - you'll need it for the next task.

---

### Task 3: Scaffold Web App

**Step 1: Create the web app**

Use the `create_web_app` MCP tool with:
- `app_name` confirmed in Task 1, Step 3
- `db_service_id` from Task 2
- `use_auth: true` if multi-user app (from Task 1)

**Step 2: Change into app directory**

```bash
cd <app_name>
```

**Step 3: Read project context**

Read the `CLAUDE.md` file in the newly created app directory into your context.

---

## Phase 2: Auth Configuration (If Multi-User)

### Task 4: Configure Auth Providers

**Files:**
- Modify: `src/server/better-auth/config.ts`
- Modify: `src/env.js`

**Step 1: Edit auth config**

Update the Better Auth configuration to enable only the providers the user requested:

```typescript
// Example for email + github
export const authConfig = {
  providers: [
    emailProvider(),
    githubProvider({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),
  ],
};
```

**Step 2: Update env.js**

Update the src/env.js file to set the environment variables for the auth providers.

---

## Phase 3: Database Schema

### Task 5: Fix Schema Table Prefix

**Files:**
- Modify: `src/server/db/schema.ts`
- Read: `drizzle.config.ts`

**Step 1: Check drizzle config**

Read `drizzle.config.ts` and note the `tablesFilter` or prefix setting.

**Step 2: Update schema prefix**

In `src/server/db/schema.ts`, replace the `pg_drizzle` prefix with whatever prefix is configured in `drizzle.config.ts`.

---

### Task 6: Design Database Schema

**Files:**
- Modify: `src/server/db/schema.ts`

**Step 1: Remove example post table**

Delete the example `post` table definition - it was only there as a template.

**Step 2: Design tables for the app**

Based on the user's app requirements, add the necessary Drizzle table definitions to `src/server/db/schema.ts`.

---

### Task 7: Push Schema to Database

**Step 1: Wait for database to be ready**

Run `tiger service list -o json` and check that the database service has `"status": "READY"`. If not, wait and retry in a loop for up to 2 minutes.

**Step 2: Push schema**

```bash
npm run db:push
```

---

## Phase 4: Backend Implementation

### Task 8: Implement tRPC Backend

**Files:**
- Create/Modify: `src/server/api/routers/*.ts`
- Modify: `src/server/api/root.ts`

**Step 1: Remove example post router**

Delete any tRPC routes that reference the old post model.

**Step 2: Create routers**

Add tRPC routers for CRUD operations on your data models. Follow the patterns in existing routers.

**Step 3: Register routers**

Add new routers to `src/server/api/root.ts`.

---

## Phase 5: Frontend Implementation

### Task 9: Install Required shadcn Components

**Step 1: Identify needed components**

Determine which shadcn components are needed for the app (button, card, input, form, table, etc.)

**Step 2: Install components**

```bash
npx shadcn@latest add button card input label form
```

Note: `shadcn init` was already run. Only add individual components.

---

### Task 10: Implement Frontend Pages

**Files:**
- Create/Modify: `src/app/**/*.tsx`
- Create: `src/components/*.tsx`

**Step 1: Create page components**

Build the pages needed for your app using shadcn components.

**Step 2: Connect to backend**

Use tRPC hooks to fetch and mutate data from your routers.

**Step 3: Create sign-in component (if multi-user)**

Build a reusable sign-in form component at `src/components/auth/sign-in-form.tsx` using shadcn components that supports all auth methods the user requested:
- Email: email/password form fields
- GitHub: "Sign in with GitHub" button
- Google: "Sign in with Google" button

**Step 4: Fix color scheme**

Replace any hardcoded T3 template colors with shadcn CSS variables. Examples:

| Replace | With |
|---------|------|
| `bg-gradient-to-b from-slate-900 to-slate-800` | `bg-background` |
| `text-white` | `text-foreground` |
| `bg-white/10` | `bg-muted` |
| `border-white/20` | `border-border` |

---

## Phase 6: Run and Verify

### Task 11: Run and Verify

**Step 1: Start the dev server**

```bash
npm run dev
```

**Step 2: Open in browser**

Use the `open_app` MCP tool to open http://localhost:3000 in a browser and verify the app works as expected.

---

### Task 12: Finish Up

**Step 1: Offer to commit**

Ask the user if they would like to commit the app to git (and highlight the question). Don't include the question with the summary of what you did.

If yes, then run the following commands:
```bash
git init
git add .
git commit -m "Initial commit: <app_name>"
```

---

### Task 13: Summarization

**Step 1: Highlight the next steps**

Highlight the next steps a user can take:
- Plan out the next steps for the app development using superpowers:brainstorming
- Use the deploy-app skill to deploy the app
