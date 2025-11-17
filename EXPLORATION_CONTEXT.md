# Action Tools Architecture Exploration

**Status:** EXPLORATION PHASE (not committed to this approach yet)

## Why We're Exploring This

Current template approach works (5.5 min for full SaaS app) but has scalability concerns:
- Claude interprets templates inconsistently (ignores pnpm, uses npm instead)
- Context window limits (can't load 20+ templates efficiently)
- Combination explosion (2^N possible combinations)
- No reusability between runs
- Unpredictable behavior as templates grow

## What We're Exploring

**Alternative: Action Tools Approach**
- Fast MCP tools that generate code directly (not through interpretation)
- Target: 30-60 seconds for common patterns
- Still keep templates for custom/unusual requests (Two-Tier system)

## Current State

- **Branch:** `action-tools-architecture`
- **Main branch:** Still has working template approach (don't delete!)
- **Previous branch deleted:** `speed-optimizations-pnpm-free-db` (had pnpm/drizzle-kit optimizations)

## Test Results from Template Approach

Last test: 5 minutes 29 seconds total
- Database: 30 sec (free tier)
- File creation: 1 min
- npm install: 9 sec (cached)
- Migration: 5 sec
- Debugging db.execute: 3 min (could be avoided with better templates)

## Questions to Answer During Exploration

1. Can action tools be faster and more predictable?
2. Will they scale better to many features (50+ tools)?
3. Do they provide enough flexibility?
4. Are they worth the engineering effort vs improving templates?
5. Should we do hybrid (action tools for 80% case, templates for 20%)?

## Next Steps

1. Design what an action tool interface would look like
2. Prototype ONE action tool (e.g., `create_saas_app`)
3. Compare pros/cons vs template approach
4. Make decision: templates, action tools, or hybrid

## Key Constraint

Must maintain 0perator's promise: "Type as little as possible" - natural language interface is non-negotiable.

## Resources

- Templates location: `internal/prompts/md/`
- MCP server code: `cmd/0perator-mcp/`
- Current templates: create_web_app, database_tiger, auth_jwt, payments_stripe, deploy_railway, etc.
