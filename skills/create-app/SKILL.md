---
name: create-app
description: 'Use this skill whenever creating a new application. IMPORTANT: This should be the FIRST thing you read when starting a new project. IMPORTANT: Read this before planning or brainstorming.'
---

# Create App Implementation Plan

> **For Claude Code:** Follow this plan phase-by-phase. If any step fails, notify the user and ask for next steps.

**Phase Summary Template:** At the end of each phase, output to the user a detailed summary (2-3 paragraphs) that includes:
- What was accomplished in this phase
- Key files/directories created or modified
- Important values to remember (service_id, schema names, etc.)
- Brief preview of what's coming next

**Goal:** Scaffold a production-ready fullstack web application with database, optional auth, and polished UI.

**Architecture:** T3 stack (Next.js + tRPC + Drizzle) with Timescale Cloud database and shadcn/ui components.

**Tech Stack:** Next.js, tRPC, Drizzle ORM, Timescale Cloud (PostgreSQL), Better Auth, shadcn/ui, Tailwind CSS

---

## Getting Started

Tell the user:

"Let's build a minimal v0/demo version of your app. We'll focus on the core features needed to get something working, then iterate from there.

Here's how we'll build this:
- 🎯 **Phase 1: Understand the product** - I'll ask a few questions about what you're building
- 🏗️ **Phase 2: Scaffold the app** - Create a cloud database and app with Next.js, tRPC, and Drizzle
- 📋 **Phase 3: Plan the architecture** - Generate schema, API, and frontend plans in parallel
- 🔐 **Phase 4: Configure auth** (if needed) - Set up user authentication
- 🛠️ **Phase 5: Build everything** - Database, backend, and frontend in parallel
- ✅ **Phase 6: Run, verify, and commit** - Make sure everything works and save to git

**Optional hardening (after commit):**
- 🧪 **Phase 7: Add testing** - Integration tests with Vitest
- 🔍 **Phase 8: Strict checks** - Stricter TypeScript and linting

Let's go!"

---

## Phase 1: Understand the Product

**Pre-flight check:** If a `CLAUDE.md` file already exists in the current directory, stop and tell the user: "A CLAUDE.md already exists here. Please run this from an empty directory or one without an existing project."

**Key Principles:**
- One question at a time - Don't overwhelm with multiple questions
- Multiple choice preferred - Easier to answer than open-ended when possible
- YAGNI ruthlessly - Remove unnecessary features from all designs
- Explore alternatives - Always propose 2-3 approaches before settling
- Incremental validation - Present design in sections, validate each
- Be flexible - Go back and clarify when something doesn't make sense

1. **Determine app type** - If not clear from the prompt, ask: "Is this a multi-user app (requires user accounts/login)?"

2. **Gather auth requirements (if multi-user)** - Ask the user: "Which authentication methods do you want? Pick one or more:" Email signup, GitHub OAuth, Google OAuth

3. **Confirm app name** - Propose a sensible app name based on the user's request. The name should be lowercase, use hyphens instead of spaces (e.g., `todo-app`, `fitness-tracker`), and appropriate for a directory name. Ask: "I'll name the project `<proposed-name>`. Does that work, or would you prefer something else?"

4. **Understand what product you are building** - Try to understand the project from the user prompt then ask questions one at a time to refine the idea. Focus on what the product will do, NOT technical details.

Once you understand what you're building, present the **product brief** to the user for confirmation:

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

After the user confirms the product brief, ask: "Are there any features not in the v0/demo that might affect how we build this? For example: offline support, real-time sync, multi-tenancy, or specific integrations. These won't be built now, but knowing them helps us make the right architectural choices upfront."

If yes, create and confirm a list of "future features".

---
## Phase 2: Project Setup

1. Use the `create_database` MCP tool to provision a new Timescale Cloud database
2. Store the returned `service_id` - you'll need it later
3. Use the `create_web_app` MCP tool with:
   - `app_name` confirmed in Phase 1
   - `use_auth: true` if multi-user app
   - `product_brief` from Phase 1
   - `future_features` from Phase 1 (if any)
4. Write a pointer `CLAUDE.md` in the current directory (before changing into the app):
   ```markdown
   ## In Progress

   Building an app at `./<app_name>/`. See `./<app_name>/CLAUDE.md` for details.
   ```
5. Change into the app directory: `cd <app_name>`
6. Write a brief `CLAUDE.md` with the product context (planning agents will read this):
   ```markdown
   # <App Name>

   ## Product Brief
   <product brief from Phase 1>

   ## Features (v0/demo)
   <minimal features list from Phase 1>

   ## Future Features
   <future features from Phase 1, or "None identified yet">

   ## Database
   Service ID: <service_id from step 2>

   ## Status
   We are in the middle of building an app using the `create-app` skill from 0perator. If you've just loaded this file for a new session ask the user:
   "Looks like you are in the middle of building an app, continue where you left off?" If yes, then load the `create-app` 0perator skill and resume at the current phase.

   Current Phase: Phase 3
   ```
7. Output a phase summary to the user using the template.


## Phase 3: Planning

1. **Run planning agents in parallel** - Use the Task tool to launch these three agents simultaneously, passing the **full application path** in each prompt:
   - `0perator:schema-designer` - Creates `<app_path>/docs/plan/schema.sql`
   - `0perator:api-planner` - Creates `<app_path>/docs/plan/trpc-api.md`
   - `0perator:frontend-planner` - Creates `<app_path>/docs/plan/frontend-plan.md`

   **Important:** Include the full path (e.g., `/path/to/todo-app/docs/plan/`) in each agent's prompt, not relative paths.

2. **Verify and fix plans** - After all three complete, use the Task tool to run:
   - `0perator:plan-verifier` - Cross-checks all plans and fixes inconsistencies

3. **Review with user** - Briefly summarize the planned schema, API, and frontend structure.

4. **Update status** - Update `CLAUDE.md` to set `Current Phase: Phase 4`

5. Output a phase summary to the user using the template.

---

## Phase 4: Auth Configuration (If Multi-User)

Skip this phase if the app is single-user.

1. Use the Task tool to run the `0perator:auth-configurator` agent, passing the **full application path** in the prompt.
2. **Update status** - Update `CLAUDE.md` to set `Current Phase: Phase 5`
3. Output a phase summary to the user using the template.

---

## Phase 5: Build Database, Backend, and Frontend

1. **Set up frontend scaffolding** - Run the `0perator:frontend-builder` agent to clean up T3 scaffold and install shadcn/ui components.

2. **Run build agents in parallel**, passing the **full application path** in each prompt:
   - `0perator:database-setup` - Also pass the `service_id` from Phase 2
   - `0perator:backend-builder` - Creates tRPC routers
   - `0perator:frontend-builder-group` (multiple) - Read `docs/plan/frontend-plan.md` to identify sections, then spawn one agent for **Global** and one for each **Route Group** (e.g., `(admin)`). Each prompt should include the full path and which section to build (e.g., "Build the Global section (shared components and flat routes) for the app at /path/to/app").

3. **After all agents complete**, run the `0perator:code-checker` agent to fix formatting and verify checks pass.

4. **Update status** - Update `CLAUDE.md` to set `Current Phase: Phase 6`

5. Output a phase summary to the user using the template.

---

## Phase 6: Run, Verify, and Commit

1. Start the dev server: `npm run dev`

2. Use the `open_app` MCP tool to open http://localhost:3000 in a browser and verify the app works as expected

3. Use the `write_claude_md` MCP tool to generate the project guide:
   - `application_directory`: "."
   - `app_name`: from Phase 1
   - `use_auth`: Whether auth is enabled
   - `product_brief`: from Phase 1
   - `future_features`: from Phase 1 (if any)
   - `db_schema`: from `setup_app_schema` in Phase 5
   - `db_user`: from `setup_app_schema` in Phase 5

4. Read the generated `CLAUDE.md` file. Make sure it is accurate. Fix if needed.

5. Ask the user "Do you want to commit this initial version to git?". If yes:
   ```bash
   git init
   git add .
   git commit -m "Initial commit: <app_name>"
   ```

6. Tell the user:

"🎉 Congrats! Your app is set up and committed. You have a working demo you can iterate on.

**Next steps:**
- 🛡️ **Harden** - Add testing and stricter checks (recommended)
- 🧠 **Brainstorm** - Plan your next features
- 🚀 **Deploy** - Ship to Vercel

**Why harden?** These checks act like a reward signal for AI-assisted development - they catch mistakes early and help guide me toward correct solutions faster. Without them, bugs can compound silently.

Shall we add hardening now?"

If the user wants to skip hardening, the skill is complete.

---

## Phase 7: Backend Testing (Optional Hardening)

Ask the user (yes/no): "Do you want to add backend testing?"

If no, skip this phase.

1. Use the `view_skill` MCP tool to read the `add-backend-testing` skill
2. Follow the skill with `service_id` from Phase 2
3. Offer to git commit these changes

---

## Phase 8: Stricter Checks (Optional Hardening)

Ask the user (yes/no): "Do you want to enable stricter TypeScript checks?"

If no, skip this phase.

1. Use the `view_skill` MCP tool to read the `add-strict-checks` skill
2. Follow the skill
3. Offer to git commit these changes
