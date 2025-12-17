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

### Task 1: Gather Requirements And Understand The Project

Before asking any questions tell the user:

"Let's start by planning a minimal v0/demo version of your app. We'll focus on the core features needed to get something working, then we can iterate from there.

Here's how we'll build this:
1. 🎯 **Understand the product** - I'll ask a few questions to understand what you're building
2. 🏗️ **Set up infrastructure** - Create a cloud database and scaffold the app with Next.js, tRPC, and Drizzle
3. 🔐 **Configure auth** (if needed) - Set up user authentication
4. 🗄️ **Design the database** - Create tables for your data
5. ⚙️ **Build the backend** - Create API endpoints with tRPC
6. 🧪 **Add testing** (if wanted) - Set up integration tests with Vitest
7. 🎨 **Build the frontend** - Create pages and components with shadcn/ui
8. 🔍 **Configure strict checks** (if wanted) - Set up stricter TypeScript and linting to catch AI-generated code issues, and fix any issues in the scaffold
9. ✅ **Run and verify** - Make sure everything works
10. 💾 **Commit** - Save this initial version so we can iterate from here

Let's start with understanding your product."

Stress that this will be a v0/demo version we'll iterate

DO NOT ask multiple questions in the same prompt.

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

**Step 4: Understand what product you are building**

You are building a new application so try to understand the project from the user prompt then ask questions one at a time to refine the idea.
Right now you need to understand the project from the perspective of what the product will do. DO NOT try to determine the technical details now.

Once you understand what you're building, present the **product brief** to the user for confirmation. The product brief should include:

1) **App type**: Single-user or multi-user
2) **Authentication** (if multi-user): Which methods (email, GitHub, Google)
3) **Product description**: A one to three paragraph description of what the project will do
4) **Minimal features for v0/demo**: A short bulleted list - just enough to get a working application

Example product brief:
```
**App type:** Multi-user

**Authentication:** Email signup

**Product description:**
A collaborative to-do app where users can create personal to-do lists and share them with other users. Users sign up with email, create tasks, and can invite collaborators to view or edit their lists together.

**Minimal features for v0/demo:**
- Email signup/login
- Create, edit, delete, and complete to-dos
- Share a to-do list with another user by email
- Collaborators can view and edit shared lists
```

Ask the user: "Is this product brief correct?"

After the user confirms the product brief:
- Ask the user: "Are there any features not in the v0/demo that might affect how we build this? For example: offline support, real-time sync, multi-tenancy, or specific integrations. These won't be built now, but knowing them helps us make the right architectural choices upfront."
- If yes:
  1) create a list of such features.
  2) present the list of features to the user for confirmation.
  
Let's call this list the "future features"

#### The Process
Understanding the idea:

- Ask questions one at a time to refine the idea
- Prefer multiple choice questions when possible, but open-ended is fine too
- Only one question per message - if a topic needs more exploration, break it into multiple questions
- Focus on understanding: purpose, constraints, success criteria

Exploring approaches:

- Propose 2-3 different approaches with trade-offs
- Present options conversationally with your recommendation and reasoning
- Lead with your recommended option and explain why


Key Principles:
One question at a time - Don't overwhelm with multiple questions
Multiple choice preferred - Easier to answer than open-ended when possible
YAGNI ruthlessly - Remove unnecessary features from all designs
Explore alternatives - Always propose 2-3 approaches before settling
Incremental validation - Present design in sections, validate each
Be flexible - Go back and clarify when something doesn't make sense

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
- `use_auth: true` if multi-user app (from Task 1)
- `product_brief` from Task 1, Step 4 (the product brief)
- `future_features` from Task 1, Step 4 (if any future features were identified)

**Step 2: Change into app directory**

```bash
cd <app_name>
```

**Step 3: Upgrade dependencies**

```bash
npx npm-check-updates -u --reject drizzle-orm
npm install
```

**Step 4: Read project context**

Read the `CLAUDE.md` file in the newly created app directory into your context.

---

## Phase 2: Auth Configuration (If Multi-User)

### Task 4: Configure Auth Providers

**Files:**
- Modify: `src/server/better-auth/config.ts`
- Modify: `src/env.js`, `.env` , `.env.example`

**Step 1: Pass the drizzle schemas into drizzleAdapter**

```typescript
import * as schema from "~/server/db/schema";

//when initiating the drizzle adapterb pass in the schema
  drizzleAdapter(db, {
    provider: "pg",
    schema,
})
```


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

**Step 2: Update env files**

Update the `src/env.js`,`.env` and `.env.example` files to set the environment variables for the auth providers.

---

## Phase 3: Database Schema

### Task 5: Set Up Database Connection

**Step 1: Wait for database to be ready**

Check that the database status is `READY` using the `service_get` MCP tool with the `service_id` from Task 2.
If not ready, poll every 10 seconds for up to 2 minutes.

**Step 2: Set up app schema**

Use the `setup_app_schema` MCP tool with:
- `application_directory`: "."
- `service_id` from Task 2
- `app_name` (use the same name, converted to lowercase with underscores)

This creates:
- A PostgreSQL schema named after the app
- A database user with the same name and limited permissions
- Writes `DATABASE_URL` and `DATABASE_SCHEMA` to `.env`

---

### Task 6: Configure Schema Support

**Step 1: Add a DATABASE_SCHEMA env variable**

In `src/env.js` add DATABASE_SCHEMA variable (use the `schema_name` returned by `setup_app_schema` as default) to both the server and runtimeEnv sections

**Step 2: Change drizzle config to obey schema**

Modify `drizzle.config.ts` to remove the tablesFilter and add a schemasFilter with the value of the DATABASE_SCHEMA env variable.

**Step 3: Update schema table definitions**

In `src/server/db/schema.ts`, remove pgTableCreator pattern and instead create all tables (including auth tables, if present) using createTable:

```typescript
export const dbSchema = pgSchema(env.DATABASE_SCHEMA);
const createTable = dbSchema.table;
```
Note: make sure the schema is exported

---

### Task 7: Design Database Schema

**Files:**
- Modify: `src/server/db/schema.ts`

**Step 1: Remove example post table**

Delete the example `post` table definition - it was only there as a template.

**Step 2: Design tables for the app**

Based on the user's app requirements, add the necessary Drizzle table definitions to `src/server/db/schema.ts`.

---

### Task 8: Push Schema to Database

**Step 1: Push schema**

```bash
npm run db:push
```

---

## Phase 4: Backend Implementation

### Task 9: Implement tRPC Backend

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

## Phase 5: Backend Testing

Ask the user (yes/no) **default: yes**: "Do you want to add backend testing? This sets up isolated integration tests that run against a separate database schema - so you can confidently iterate without breaking things."
Only continue with this phase if the answer is Yes. Otherwise, skip Task 10,11,12.

### Task 10: Set Up Testing Infrastructure

**Step 1: Set up testing**

Use the `setup_testing` MCP tool to set up integration testing:

```
setup_testing(application_directory: ".", service_id: "<service_id from Task 2>")
```

This creates:
- A `test_schema` schema isolated from production data
- A `test_user` database user with permissions only on the test schema
- `vitest.config.ts` configured to load `.env.test`
- `src/test/global-setup.ts` that runs `drizzle-kit push` before tests
- `.env.test` with `DATABASE_URL` pointing to the test schema

### Task 11: Install Vitest and Add Scripts

**Step 1: Install Vitest**

```bash
npm install -D vitest dotenv
```

**Step 2: Add test scripts to package.json**

```json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest"
  }
}
```

### Task 12: Write Integration Tests for tRPC Routers

**Step 1: Write router tests**

Create a test file for each router. Example `src/server/api/routers/example.test.ts`:

```typescript
import { describe, it, expect } from "vitest";
import { appRouter } from "~/server/api/root";
import { createCallerFactory } from "~/server/api/trpc";
import { db } from "~/server/db";

const createCaller = createCallerFactory(appRouter);

const caller = createCaller({
  session: null,
  db,
  headers: new Headers(),
});

describe("exampleRouter", () => {

  describe("getAll", () => {
    it("returns empty array when no items exist", async () => {
      const result = await caller.example.getAll();
      expect(result).toEqual([]);
    });
  });

  describe("create", () => {
    it("creates a new item", async () => {
      const result = await caller.example.create({ title: "Test" });
      expect(result.title).toBe("Test");
    });
  });
});
```

**Step 2: Run tests**

```bash
npm test
```

Ensure all tests pass before proceeding.

---

## Phase 6: Frontend Implementation

### Task 13: Install Required shadcn Components

**Step 1: Install shadcn**

```bash
npx shadcn@latest init --base-color=neutral
```

**Step 2: Set Orange Theme**

```
cp src/styles/globals.css.orange src/styles/globals.css
```

**Step 2: Identify needed components**

Determine which shadcn components are needed for the app (button, card, input, form, table, etc.)

**Step 3: Install components**

```bash
npx shadcn@latest add button card input label form
```

---

### Task 14: Implement Frontend Pages

**Files:**
- Create/Modify: `src/app/**/*.tsx`
- Create: `src/components/*.tsx`

**Step 1: Create page components**

Build the pages needed for your app using shadcn components. Make sure all buttons have a type.

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

## Phase 7: Stricter Checks

Ask the user (yes/no) **Default: Yes**: "Do you want to enable stricter TypeScript checks? These catch bugs that standard TypeScript misses - especially useful when iterating quickly with AI assistance."
Only continue with this phase if the answer is Yes. Otherwise, skip Task 15.

### Task 15: Configure Stricter TypeScript and Linting

**Step 1: Add stricter compiler options**

Add these additional strict options to `tsconfig.json` under `compilerOptions`:

```json
{
  "compilerOptions": {
    // ... existing options ...

    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noImplicitOverride": true,
    "forceConsistentCasingInFileNames": true,
    "exactOptionalPropertyTypes": true,
    "useUnknownInCatchVariables": true
  }
}
```

**Step 2: Add check script to package.json**

```json
{
  "scripts": {
    "check": "biome check . && tsc --noEmit -p tsconfig.check.json"
  }
}
```

Note: tsconfig.check.json already exists don't try to create it.

**Step 3: Auto-fix issues**

```bash
npm run check:unsafe
```

Fix any remaining issues. NEVER disable any checks in biome, tsconfig.json or tsconfig.check.json. Instead, fix the code to not violate the check.

**Step 3: Run checks**

```bash
npm run check
```

---

## Phase 8: Run and Verify

### Task 16: Run and Verify

**Step 1: Start the dev server**

```bash
npm run dev
```

**Step 2: Open in browser**

Use the `open_app` MCP tool to open http://localhost:3000 in a browser and verify the app works as expected.

---

### Task 17: Finish Up

**Step 1: Review CLAUDE.md**

Read the `CLAUDE.md` file. Make sure it is accurate. Fix if needed.

**Step 2: Offer to commit**

Ask the user "Do you want to commit this initial version to git?".

If yes, then run the following commands:
```bash
git init
git add .
git commit -m "Initial commit: <app_name>"
```

---

### Task 18: Summarization

**Step 1: Highlight the next steps**

Highlight the next steps a user can take:
- Plan out the next steps for the app development using superpowers:brainstorming
- Use the deploy-app skill to deploy the app
