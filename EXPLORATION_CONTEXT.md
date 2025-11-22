# Action Tools Architecture Exploration

**Status:** IMPLEMENTED & EVOLVED (Direct Tools + Recipes approach adopted)

## Executive Summary

After exploring multiple architectures and considering Anthropic's MCP code execution patterns, we evolved from a template-based system to a direct tools + recipes approach. This provides better agentic ergonomics, faster execution, and cleaner AI interactions.

## Journey & Key Decisions

### 1. Initial Problem (Template Approach)
- Claude interpreted templates inconsistently (ignored pnpm, used npm instead)
- Context window limits (can't load 20+ templates efficiently)
- Combination explosion (2^N possible combinations)
- No reusability between runs
- 5+ minute execution times for full SaaS apps

### 2. Action Tools Exploration
- **Branch:** `action-tools-architecture`
- Initial concept: Fast MCP tools that generate code directly
- Target: 30-60 seconds for common patterns
- Two-tier system considered (actions for 80%, templates for 20%)

### 3. Anthropic's MCP Code Execution Article (Nov 2024)
- **Reference:** [Building Effective Agents](https://www.anthropic.com/news/building-effective-agents)
- Key insight: Code execution tools can enable AI to solve O(N²) complexity problems
- Their approach: Let Claude write and execute code dynamically
- **Our assessment:** While powerful for general problems, our domain (web app scaffolding) benefits more from pre-built, optimized tools that complete in seconds rather than minutes

### 4. Critical Architecture Decision

**User Question:** "Do we even need actions?"

This led to a fundamental shift:
- Dropped the action abstraction layer
- Moved to **direct MCP tools** for immediate operations
- Added **recipes** for composition (text-based, community-friendly)

### 5. Recipe System Design

User requirement: "Super easy so others can contribute"

Final format (minimal YAML):
```yaml
name: SaaS Starter
desc: Complete SaaS app with Next.js, PostgreSQL, auth, and payments
inputs:
  app_name: string = my-saas
  auth: jwt = jwt
steps:
  - create_nextjs_app name={{app_name}} typescript=true tailwind=true
  - setup_postgres_free name={{app_name}}_db
  - add_jwt_auth
  - add_stripe_payments mode=subscription
```

## Current Architecture (v2.0.0)

### Direct Tools (10 implemented)

**Generic tools with smart defaults:**
- `create_web_app` - Defaults to Next.js, accepts framework param
- `setup_database` - Defaults to PostgreSQL free tier, accepts type param

**Specific tools:**
- `create_nextjs_app`, `create_react_app`, `create_express_api`
- `setup_postgres_free` - Tiger Cloud with fixed CLI integration
- `setup_sqlite` - Local database for development

**Feature tools (placeholders):**
- `add_jwt_auth`, `add_stripe_payments`

**Orchestration:**
- `operator` - For complex multi-step operations

### Key Implementation Fixes

1. **Tiger CLI Integration**
   - Fixed: `--json` → `-o json`
   - Fixed: Removed `--addons` flag for free tier
   - Philosophy change: "Tiger Cloud for everything" (not just production)
   - Free tier defaults: shared CPU, auto-includes time-series + AI extensions

2. **Agentic Ergonomics**
   - Removed decision points (no more "which tier?" questions)
   - Smart defaults everywhere (Next.js for web, Postgres for DB)
   - Tool descriptions guide Claude to preferred choices

3. **MCP SDK Compatibility**
   - Kept operator wrapper pattern for backward compatibility
   - Direct tools exposed for cleaner Claude interactions
   - Typed handlers for MCP compliance

## Performance Results

**Before (Templates):** 5+ minutes for full SaaS app
**After (Direct Tools):** <1 minute for same result
- No template interpretation overhead
- Direct execution paths
- Parallel tool execution where possible

## Lessons Learned

1. **Abstraction vs Directness**
   - Too many abstraction layers confuse AI agents
   - Direct tools with clear names work better
   - Keep escape hatches (operator) for complex cases

2. **Defaults Matter**
   - Every decision point slows down AI agents
   - Good defaults (Next.js, Postgres) cover 80% of cases
   - Let users be explicit only when needed

3. **Community Contribution**
   - Text-based formats (YAML recipes) lower barriers
   - Minimal syntax reduces errors
   - Variable substitution keeps it flexible

4. **Code Execution vs Pre-built Tools**
   - Anthropic's approach: General purpose, flexible, slower
   - Our approach: Domain-specific, fast, predictable
   - Right tool for right job (we're optimizing for speed)

## Future Considerations

1. **Recipe Library**
   - Build community recipe repository
   - Categories: SaaS, Blog, E-commerce, API, etc.
   - Version management for recipes

2. **Tool Implementation**
   - Complete JWT auth implementation
   - Add Stripe payments integration
   - More deployment targets beyond Railway

3. **Local-First Options**
   - SQLite implemented for local development
   - Consider Docker Postgres for prod parity
   - Offline-capable workflows

## Key Constraint (Maintained)

"Type as little as possible" - Natural language interface remains non-negotiable.

## Files & Locations

- **Direct tools:** `internal/server/tools_direct.go`
- **Operator wrapper:** `internal/server/tools_v2.go`
- **Tool implementations:** `internal/tools/`
- **Action implementations:** `internal/actions/implementations/`
- **Recipe system:** `internal/recipes/`
- **Recipe files:** `recipes/*.yaml`

## Decision Record

**November 2024:** Adopted Direct Tools + Recipes architecture
- Simpler than action abstraction
- Faster execution than templates
- Better agentic ergonomics
- Community-friendly contribution model