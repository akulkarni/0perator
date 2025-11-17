# 0perator Design Principles

## Core Philosophy

0perator enables developers to build real, production-capable applications with as few words as possible, in as little time as possible. A developer should be able to say "Build a basic SaaS app with auth and Postgres" and have a running application in minutes.

## Design Principles

### 1. **Balanced Optimization**
- Optimize heavily for speed on common patterns (< 2 minutes)
- Maintain flexibility for edge cases (5-10 minutes)
- Two-tier system: fast path + flexible path
- Don't sacrifice quality for speed, but prioritize speed for common cases

### 2. **Outcome: Running Application**
- Primary goal: User can immediately use the app
- Deploy locally or to cloud within minutes
- Not just code generation - deliver a full running system
- "Build it" means "build it AND run it"

### 3. **Sensible Defaults Over Configuration**
- Use best practices and proven tech stacks by default
- Prioritize speed over perfect customization
- Allow customization when explicitly requested
- Users shouldn't need to configure everything upfront

### 4. **Scalability Through Composition**
- ✅ **Features compose cleanly** - No conflicts when combining features
- ✅ **Easy to add new features** - Simple to extend with new capabilities
- ✅ **Performance stays fast** - Common cases don't slow down as system grows
- Support 50+ features without degrading experience

### 5. **Edge Cases Get Proper Treatment**
- Unusual requests deserve time and care
- Don't force-fit into standard patterns
- Quality over speed for edge cases
- Take 10+ minutes if needed to do it right

### 6. **High-Level Transparency**
- Tell users what was built
- Summarize key features added (e.g., "Built web app with JWT auth + Tiger Postgres")
- Don't overwhelm with implementation details
- Users shouldn't need to understand internals

### 7. **No DSL - Markdown + Examples**
- Contributors write markdown guides with code examples
- No special syntax, configuration languages, or DSLs to learn
- Lower barrier = more community contributions
- If you built it, you can contribute it

### 8. **LLM-Native Architecture**
- Improves automatically as base models improve
- System interprets guides better over time
- No rewrites needed when GPT-5/Claude-4/Opus-5 arrives
- Architecture that gets better with AI progress

### 9. **Magical User Experience**
- Feels effortless: "just say what you want"
- Produces real, production-capable applications
- "Wow, I can't believe that worked" factor
- Should feel like magic, not like configuration

## Architectural Implications

### Two-Tier Pattern System

**Tier 1: Optimized Patterns (Fast Path)**
- Hand-optimized markdown guides for common use cases
- Same markdown format as community patterns
- Performance-tuned for speed (< 2 minutes)
- Examples: web apps, auth, databases, payments, deployment
- Gets better as LLMs improve at interpretation

**Tier 2: Community Patterns (Flexible Path)**
- Anyone can contribute markdown + examples
- LLM interprets naturally
- Slower but works for any feature (5-10 minutes)
- Handles edge cases and unusual requests

**Key Insight:** Both tiers use the same contribution format (markdown), but Tier 1 patterns receive additional performance optimization. As LLMs improve, Tier 2 patterns automatically get faster without requiring changes to the patterns themselves.

### Example User Experience

**Common Pattern (Fast Path):**
```
User: "Build a SaaS app with auth and Postgres"

System:
1. Recognizes common pattern
2. Uses optimized implementations
3. Composes features cleanly
4. Returns: "Built web app with JWT auth + Tiger Postgres database. Running at localhost:3000"

Time: < 2 minutes, fully deployed and usable
```

**Edge Case (Flexible Path):**
```
User: "Build a SaaS app but use Passport.js for auth and add custom OAuth flow"

System:
1. Recognizes deviation from standard
2. Uses flexible approach (templates + LLM interpretation)
3. Takes time to build exactly what was requested
4. Returns: Running application with custom auth

Time: 5-10 minutes, but built exactly as requested
```

## Success Metrics

- **Speed**: Common patterns complete in under 2 minutes
- **Quality**: Production-capable code, not just boilerplate
- **Usability**: User types natural language, gets running app
- **Composability**: Can combine any features without conflicts
- **Extensibility**: Easy for community to add new patterns
- **Magic**: "I can't believe that just worked" reactions

## Non-Goals

- ❌ Supporting every possible configuration option
- ❌ Generating code without running it
- ❌ Requiring users to understand the implementation
- ❌ Creating a DSL or configuration language
- ❌ Forcing users into specific tech stacks (unless speed-critical)
