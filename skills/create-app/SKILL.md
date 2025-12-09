---
name: create-app
description: 'Step-by-step plan for creating a new fullstack web application with database, auth, and shadcn UI'
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

## Phase 2: UI Refinement

### Task 4: Install Required shadcn Components

**Files:**
- Modify: `package.json` (dependencies added automatically)
- Create: `src/components/ui/*.tsx` (component files)

**Step 1: Identify needed components**

Review existing pages and determine which shadcn components are needed (button, card, input, form, etc.)

**Step 2: Install components**

```bash
npx shadcn@latest add button card input label form
```

Note: `shadcn init` was already run. Only add individual components.

---

### Task 5: Refactor Pages to Use shadcn Components

**Files:**
- Modify: `src/app/page.tsx`
- Modify: Any other pages in `src/app/`

**Step 1: Replace HTML elements with shadcn components**

Replace native elements with shadcn equivalents:
- `<button>` → `<Button>`
- `<input>` → `<Input>`
- Cards/containers → `<Card>`, `<CardHeader>`, `<CardContent>`

**Step 2: Verify imports**

Ensure all components are imported from `@/components/ui/*`

---

### Task 6: Fix Color Scheme

**Files:**
- Modify: `src/app/page.tsx`
- Modify: All pages in `src/app/`

**Step 1: Remove T3 template colors**

Find and replace hardcoded colors with shadcn CSS variables. Examples:

| Replace | With |
|---------|------|
| `bg-gradient-to-b from-slate-900 to-slate-800` | `bg-background` |
| `text-white` | `text-foreground` |
| `bg-white/10` | `bg-muted` |
| `border-white/20` | `border-border` |

Look for any hardcoded Tailwind colors (slate, gray, white, etc.) and replace with semantic shadcn tokens.

**Step 2: Verify consistency**

Check every page uses shadcn color tokens, not hardcoded colors.

---

## Phase 3: Auth Configuration (If Multi-User)

### Task 7: Configure Auth Providers

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

### Task 8: Create Sign-In Component

**Files:**
- Create: `src/components/auth/sign-in-form.tsx`
- Modify: Sign-in page to use the new component

**Step 1: Create sign-in form component**

Build a reusable sign-in form component using shadcn components that supports all auth methods the user requested.

**Step 2: Include all requested methods**

- Email: email/password form fields
- GitHub: "Sign in with GitHub" button
- Google: "Sign in with Google" button

---

## Phase 4: Database Schema

### Task 9: Fix Schema Table Prefix

**Files:**
- Modify: `src/server/db/schema.ts`
- Read: `drizzle.config.ts`

**Step 1: Check drizzle config**

Read `drizzle.config.ts` and note the `tablesFilter` or prefix setting.

**Step 2: Update schema prefix**

In `src/server/db/schema.ts`, replace the `pg_drizzle` prefix with whatever prefix is configured in `drizzle.config.ts`.

---

### Task 10: Remove Example Post Model

**Files:**
- Modify: `src/server/db/schema.ts`

**Step 1: Delete post table**

Remove the example `post` table definition - it was only there as a template.

**Step 2: Remove related code**

Delete any tRPC routes or components that reference the post model.

---

## Phase 5: App Implementation

### Task 11: Implement Database Schema

**Files:**
- Modify: `src/server/db/schema.ts`

**Step 1: Design and add tables**

Based on the user's app requirements, add the necessary Drizzle table definitions to `src/server/db/schema.ts`.

**Step 2: Push schema to database**

```bash
npm run db:push
```

---

### Task 12: Implement tRPC Backend

**Files:**
- Create/Modify: `src/server/api/routers/*.ts`
- Modify: `src/server/api/root.ts`

**Step 1: Create routers**

Add tRPC routers for CRUD operations on your data models. Follow the patterns in existing routers.

**Step 2: Register routers**

Add new routers to `src/server/api/root.ts`.

---

### Task 13: Implement Frontend Pages

**Files:**
- Create/Modify: `src/app/**/*.tsx`
- Create: `src/components/*.tsx`

**Step 1: Create page components**

Build the pages needed for your app using shadcn components.

**Step 2: Connect to backend**

Use tRPC hooks to fetch and mutate data from your routers.

**Step 3: Style with shadcn**

Ensure all UI uses shadcn components and color tokens.

---

### Task 14: Run and Verify

**Step 1: Start the dev server**

```bash
npm run dev
```

**Step 2: Open in browser**

Use the `open_app` MCP tool to open http://localhost:3000 in a browser and verify the app works as expected.

**Step 3: Offer to commit**

Ask the user if they would like to commit the app to git (and highlight the question). Don't include the question with the summary of what you did.

If yes, then run the following commands:
```bash
git init
git add .
git commit -m "Initial commit: <app_name>"
```

**Step 4: Highlight the next steps**

Highlight the next steps a user can take:
- Plan out the next steps for the app development using superpowers:brainstorming 
- Use the depoy-app skill to deploy the app
