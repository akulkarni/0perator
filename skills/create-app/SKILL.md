---
name: create-app
description: 'Use this skill whenever creating a new application. IMPORTANT: This should be the FIRST thing you read when starting a new project. IMPORTANT: Read this before planning or brainstorming.'
---

# Create App Implementation Plan

> **For Claude Code:** Follow this plan task-by-task. If any step fails, notify the user and ask for next steps. When I tell you to use a subagent, that means you MUST use the Task Tool. 

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
6. 🎨 **Build the frontend** - Create pages and components with shadcn/ui
7. ✅ **Run and verify** - Make sure everything works
8. 💾 **Commit** - Save this initial version so we can iterate from here

**Optional hardening (after initial commit):**
9. 🧪 **Add testing** - Set up integration tests with Vitest
10. 🔍 **Configure strict checks** - Set up stricter TypeScript and linting to catch AI-generated code issues

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

---

## Phase 2: Auth Configuration (If Multi-User)

### Task 4: Configure Auth Providers

**IMPORTANT: Spawn a subagent for the following:** The subagent should configure auth by:

1. Pass the drizzle schemas into `drizzleAdapter` in `src/server/better-auth/config.ts`:
   ```typescript
   import * as schema from "~/server/db/schema";

   drizzleAdapter(db, {
     provider: "pg",
     schema,
   })
   ```

2. Update the Better Auth configuration to enable only the providers the user requested (email, GitHub, Google)

3. Update `src/env.js`, `.env`, and `.env.example` with the required environment variables for the auth providers

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

**IMPORTANT: Spawn a subagent for the following:** The subagent should implement the tRPC backend by:

1. Removing any example/post router that references the old post model
2. Creating tRPC routers for CRUD operations on the app's data models in `src/server/api/routers/`
3. Registering new routers in `src/server/api/root.ts`
4. Verify with `npx tsc --noEmit -p tsconfig.server.json` (checks only server code, avoids frontend errors)

---

## Phase 5: Frontend Implementation

### Task 10: Implement Frontend

**IMPORTANT: Spawn a subagent for the following:** The subagent should implement the frontend by:

1. Installing and configuring shadcn:
   ```bash
   npx shadcn@latest init --base-color=neutral
   cp src/styles/globals.css.orange src/styles/globals.css
   ```

2. Installing required shadcn components (button, card, input, form, table, etc.):
   ```bash
   npx shadcn@latest add <component1> <component2> ...
   ```

3. Building the pages needed for the app using shadcn components (ensure all buttons have a type attribute)

4. Connecting pages to the backend using tRPC hooks to fetch and mutate data

5. Creating a sign-in form component at `src/components/auth/sign-in-form.tsx` (if multi-user) supporting all requested auth methods (email, GitHub, Google)

6. Replacing hardcoded T3 template colors with shadcn CSS variables:
   - `bg-gradient-to-b from-slate-900 to-slate-800` → `bg-background`
   - `text-white` → `text-foreground`
   - `bg-white/10` → `bg-muted`
   - `border-white/20` → `border-border`

7. Verify with `npm run build` and fix any errors

---

## Phase 6: Run, Verify, and Commit

### Task 11: Run and Verify

**Step 1: Start the dev server**

```bash
npm run dev
```

**Step 2: Open in browser**

Use the `open_app` MCP tool to open http://localhost:3000 in a browser and verify the app works as expected.

---

### Task 12: Finish Up

**Step 1: Generate CLAUDE.md**

Use the `write_claude_md` MCP tool to generate the project guide:
- `application_directory`: "."
- `app_name`: The app name from Task 1
- `use_auth`: Whether auth is enabled (from Task 1)
- `product_brief`: The product brief from Task 1, Step 4
- `future_features`: The future features from Task 1, Step 4 (if any)
- `db_schema`: The schema name returned by `setup_app_schema` in Task 5
- `db_user`: The user name returned by `setup_app_schema` in Task 5

**Step 2: Review CLAUDE.md**

Read the generated `CLAUDE.md` file. Make sure it is accurate. Fix if needed.

**Step 3: Run checks**

Run `npm run check:unsafe` to auto-fix formatting issues, then verify `npm run check` passes.

**Step 4: Offer to commit**

Ask the user "Do you want to commit this initial version to git?".

If yes, then run the following commands:
```bash
git init
git add .
git commit -m "Initial commit: <app_name>"
```

---

### Task 13: Congratulate and Offer Next Steps

Tell the user:

"🎉 Congrats! Your app is set up and committed. You have a working demo you can iterate on.

**🛡️ Optional hardening (recommended):**
These checks act like a reward signal for AI-assisted development - they catch mistakes early and help guide me toward correct solutions faster. Without them, bugs can compound silently:
- **Backend testing** - Integration tests with an isolated test database
- **Stricter TypeScript** - Additional type checks that catch common AI-generated code issues

Would you like to set these up now? (You can always ask for these later)

**Or skip and continue with:**
- 🧠 **Brainstorm** - Plan your next features
- 🚀 **Deploy** - Ship to Vercel"

If the user wants to skip hardening, the skill is complete.

---

## Phase 7: Backend Testing (Optional Hardening)

Ask the user (yes/no): "Do you want to add backend testing?"

If no, skip this phase.

**IMPORTANT: Spawn a subagent for the following:** The subagent should:

1. Read `CLAUDE.md` to understand the project
2. Use the `view_skill` MCP tool to read the `add-backend-testing` skill
3. Follow the skill with `service_id` from Task 2

---

## Phase 8: Stricter Checks (Optional Hardening)

Ask the user (yes/no): "Do you want to enable stricter TypeScript checks?"

If no, skip this phase.

**IMPORTANT: Spawn a subagent for the following:** The subagent should:

1. Read `CLAUDE.md` to understand the project
2. Use the `view_skill` MCP tool to read the `add-strict-checks` skill
3. Follow the skill
