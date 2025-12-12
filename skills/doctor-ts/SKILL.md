---
name: doctor-ts
description: 'Step-by-step instructions for evaluating TypeScript code for quality, identifying issues, and providing actionable recommendations'
---

# TypeScript Code Quality Evaluation

> **For Claude:** Follow this checklist systematically. Report findings with specific file paths and line numbers. Prioritize issues by severity.

**Output Style:** Report issues as you find them - don't wait until the end. After each task, output any findings immediately so the user sees progress. Use the format below for each issue. At the end, provide a summary report.

**Issue Format:**
```
[SEVERITY] Issue Title
📍 file/path.ts:123
💬 Brief description of the problem
🔧 Recommendation
```

**Goal:** Evaluate a TypeScript codebase for quality issues, identify problems, and provide actionable recommendations.

**Scope:** Type safety, runtime validation, code organization, error handling, performance, maintainability, and best practices.

**Key Insight:** TypeScript only provides compile-time safety. The gap between "types" and "runtime reality" is where bugs hide. This evaluation focuses on both static type safety AND runtime validation patterns (using Zod or similar).

---

## Phase 1: Setup and Discovery

### Task 1: Identify Scope

**Step 1: Determine what to evaluate**

Ask the user: "What would you like me to evaluate?"
- Entire codebase
- Specific directory (e.g., `src/`)
- Specific files
- Recent changes only (git diff)
- All the changes in the current pr or feature branch

**Step 2: Identify the project type**

Check for configuration files to understand the project:
- `tsconfig.json` - TypeScript configuration
- `package.json` - Dependencies and scripts (note if Zod is installed)
- `.eslintrc.*` or `eslint.config.*` - Existing lint rules
- `biome.json` - Biome configuration

**Step 3: Check for runtime validation library**

Look in `package.json` for:
- `zod` - Runtime schema validation
- `yup`, `joi`, `io-ts` - Alternative validation libraries

If no runtime validation library is present, flag this as a potential gap and recommend Zod for projects with external data boundaries.

---

## Phase 2: Reinvented Wheels

### Task 2: Find Code That Should Use Existing Libraries

**Why this matters:** Hand-rolled implementations of common functionality are bug-prone, harder to maintain, and often miss edge cases that battle-tested libraries handle. Before analyzing code quality, identify code that shouldn't exist at all.

**Step 1: Check for installed but unused API SDKs (CRITICAL CHECK)**

This is the highest-impact check in this phase. API SDKs installed but not imported indicate raw HTTP is being used with manually-maintained types that WILL drift from the actual API.

**Perform this check:**

1. Identify API SDKs in `package.json`

**Flag as HIGH priority when:**
- An API SDK is in `package.json` but has zero imports anywhere in the codebase
- AND raw HTTP calls to that API's endpoints exist

**Why this is a high-priority issue:**
- Manual TypeScript interfaces for API responses are lies waiting to happen
- The SDK maintainers update types when APIs change; your manual types don't
- You lose automatic token refresh, pagination, rate limiting, retries, and error handling
- You're maintaining hundreds/thousands of lines of code that the SDK provides for free
- Even if there's a valid reason (e.g., runtime constraints), you need Zod validation on all responses since manual types can't be trusted

examples:

| SDK in package.json | Check for imports | Check for raw HTTP if no imports |
|---------------------|-------------------|----------------------------------|
| `googleapis` | `from "googleapis"` | `googleapis.com/calendar`, `googleapis.com/drive`, etc. |
| `stripe` | `from "stripe"` | `api.stripe.com` |
| `@aws-sdk/*` | `from "@aws-sdk/"` | `amazonaws.com` |
| `openai` | `from "openai"` | `api.openai.com` |
| `twilio` | `from "twilio"` | `api.twilio.com` |
| `@sendgrid/mail` | `from "@sendgrid/"` | `api.sendgrid.com` |
| `octokit` | `from "octokit"` or `from "@octokit/"` | `api.github.com` |
| `firebase` | `from "firebase"` | `firebaseio.com` |

**Step 2: Check for hand-rolled utilities**

Look for custom implementations of functionality that well-known libraries handle better:

| Pattern | Should Use Instead |
|---------|-------------------|
| Custom date formatting/parsing | `date-fns`, `dayjs`, or `Intl.DateTimeFormat` |
| Custom deep clone/merge | `lodash/cloneDeep`, `structuredClone()` |
| Custom debounce/throttle | `lodash/debounce`, `lodash/throttle` |
| Custom UUID generation | `uuid`, `nanoid`, or `crypto.randomUUID()` |
| Custom slug generation | `slugify` |
| Custom retry logic | `p-retry`, `async-retry` |
| Custom query string parsing | `qs`, `URLSearchParams` |
| Custom CSV/JSON parsing | `papaparse`, native `JSON.parse` with schema |
| Custom email/URL validation regex | `zod`, `validator.js` |
| Custom HTTP client wrapper | `ky`, `got`, or framework-specific (tRPC, etc.) |
| Custom state machine | `xstate`, `robot` |
| Custom form validation | `react-hook-form`, `formik` + Zod |
| Custom caching logic | `lru-cache`, `node-cache` |

**Step 3: Check for reimplemented language features**

Look for:
- Custom `isEmpty`, `isNil`, `isObject` checks → Use `lodash` or write proper type guards
- Custom array utilities (unique, flatten, groupBy) → Use native methods or `lodash`
- Custom `sleep`/`delay` functions → Fine, but check if a library already provides one
- Custom `retry` or `poll` functions → `p-retry`, `p-poll`

**Step 4: Check for crypto/security implementations**

Flag as **CRITICAL** if you find:
- Custom password hashing → Use `bcrypt`, `argon2`
- Custom encryption → Use `crypto` built-in or established libraries
- Custom JWT handling → Use `jose`, `jsonwebtoken`
- Custom session management → Use framework-provided solutions
- Custom CSRF protection → Use framework middleware

**Why this is critical:** Security code must be written by experts. Custom implementations almost always have vulnerabilities.

**Step 5: Check for raw HTTP calls to APIs without installed SDKs**

If no SDK is installed, look for direct `fetch`/`axios` calls to well-known APIs that have official SDKs:

| Raw HTTP to... | Should Use Instead |
|----------------|-------------------|
| Stripe API | `stripe` SDK |
| AWS services | `@aws-sdk/*` |
| GitHub API | `octokit` |
| Slack API | `@slack/web-api` |
| Twilio API | `twilio` SDK |
| SendGrid API | `@sendgrid/mail` |
| OpenAI API | `openai` SDK |
| Google APIs | `googleapis` or specific SDKs |
| Firebase | `firebase` SDK |

**Why this matters:**
- SDKs handle authentication, retries, rate limiting, and error handling correctly
- SDKs provide TypeScript types that match the actual API responses
- Raw HTTP calls require manual type definitions that can drift from the real API
- SDKs are updated when APIs change; raw calls silently break

**Also check for:**
- APIs with published OpenAPI/Swagger specs - consider generating a typed client instead of manual `fetch` calls
- GraphQL APIs being called with raw HTTP instead of a GraphQL client (`@apollo/client`, `urql`, `graphql-request`)

Flag as **HIGH** priority when an official SDK could replace significant custom code (auth handling, retry logic, error parsing). Flag as **MEDIUM** priority for simpler cases where raw HTTP works but an SDK would be cleaner.

**Step 6: Check for other ignored installed dependencies**

Cross-reference `package.json` dependencies with actual usage:

```bash
# For each utility library, check if it's imported
grep -r "from ['\"]lodash" --include="*.ts" .
grep -r "from ['\"]date-fns" --include="*.ts" .
grep -r "from ['\"]zod" --include="*.ts" .
```

Flag as **MEDIUM** priority when:
- `lodash` is installed but custom utilities exist anyway
- `date-fns` is installed but custom date parsing is used
- `zod` is installed but manual validation code exists in the same boundary

**Step 7: Evaluate whether to recommend new dependencies**

Consider suggesting libraries when:
- Custom code is > 50 lines for common functionality
- Multiple bugs or edge cases exist in custom implementations
- The custom code duplicates well-known library functionality

Be conservative - don't recommend adding dependencies for trivial utilities (< 10 lines) unless they're error-prone (dates, crypto, validation).

---

## Phase 3: Type Safety Analysis

### Task 3: Check TypeScript Strictness

**Files:** `tsconfig.json`

**Step 1: Review compiler options**

Check for these critical strict mode settings:

| Setting | Recommended | Why |
|---------|-------------|-----|
| `strict` | `true` | Enables all strict type checks |
| `noImplicitAny` | `true` | Prevents implicit `any` types |
| `strictNullChecks` | `true` | Catches null/undefined errors |
| `noUncheckedIndexedAccess` | `true` | Safer array/object access |
| `exactOptionalPropertyTypes` | `true` | Stricter optional properties |

**Step 2: Report strictness issues**

If strict mode is not fully enabled, flag this as a **HIGH** priority issue.

---

### Task 4: Find Type Safety Violations

**Step 1: Search for `any` usage**

Search for explicit `any` types:
```
: any
as any
<any>
```

Categorize findings:
- **Legitimate:** Third-party library types, truly dynamic data (but even these should prefer `unknown` + validation)
- **Avoidable:** Lazy typing that should use proper types or Zod schemas

**Why this matters:** Even with strict TypeScript, developers bypass safety using `any`. In projects with Zod available, most `any` usages can be replaced with `unknown` + schema validation, which provides both compile-time AND runtime safety.

**Step 2: Search for type assertions**

Look for potentially unsafe type assertions:
```
as unknown as
as SomeType
!
```

Flag these patterns:
- **`as SomeType`** on external data - Should use Zod validation instead
- **`value as unknown as SomeType`** - Double cast indicates type system being bypassed
- **`value!`** on values from external sources - Should validate instead of assert

**Why this matters:** Type assertions tell the compiler "trust me" but provide no runtime guarantee. For external data, this is a bug waiting to happen.

**Step 3: Check for `@ts-ignore` and `@ts-expect-error`**

Search for TypeScript directive comments:
```
@ts-ignore
@ts-expect-error
@ts-nocheck
```

Each should have a comment explaining why it's necessary. Consider whether refactoring or adding Zod validation could eliminate the need for the directive.

---

## Phase 4: Runtime Validation (Zod Analysis)

### Task 5: Check External Data Boundaries

**Why this matters:** External data should be validated with Zod before being treated as trusted. TypeScript types are erased at runtime - only validation provides actual safety.

**Step 1: Find unvalidated external data**

Search for code that reads from external sources WITHOUT Zod validation:

| Source | What to look for |
|--------|------------------|
| JSON parsing | `JSON.parse(...)` without schema validation |
| HTTP responses | `fetch().then(res => res.json())`, `axios.get/post` results cast directly |
| API routes | `req.body`, `req.query`, `req.params` used without validation |
| Environment | `process.env.*` accessed directly throughout codebase |
| Database | ORM/SQL results cast to specific types without validation |
| WebSocket/queues | Message payloads assumed to match expected shape |

**Step 2: Identify unsafe patterns**

Flag as **HIGH** priority when:
- Data from outside the process is given a strong static type without any Zod validation
- `as SomeType` is used on parsed JSON or API responses
- Environment variables are accessed directly and trusted without central config validation

**Recommended pattern:**

```typescript
// BAD: Trusting external data
const user = await res.json() as User;

// GOOD: Validate at the boundary
const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
});

const user = UserSchema.parse(await res.json());
type User = z.infer<typeof UserSchema>; // Derive type from schema
```

---

### Task 6: Check Environment Variable Handling

**Why this matters:** Environment variables are a classic source of bugs. They should be validated once, centrally, with a Zod schema. Using `process.env.*` directly throughout the codebase is fragile.

**Step 1: Search for direct env access**

Look for:
- `process.env.FOO` used in multiple files
- `Deno.env.get` or similar
- Non-null assertions on env vars (`process.env.FOO!`)
- Type casts on env vars

**Step 2: Check for central config**

Look for a dedicated config module that:
- Uses a Zod schema to validate env vars on startup
- Exports typed config that the rest of the app imports
- Fails fast if required env vars are missing

Flag as **MEDIUM** priority if env vars are accessed directly in many places without central validation.

**Recommended pattern:**

```typescript
// src/config.ts
const EnvSchema = z.object({
  NODE_ENV: z.enum(["development", "test", "production"]),
  DATABASE_URL: z.string().url(),
  PORT: z.string().transform(Number).optional(),
});

const env = EnvSchema.parse(process.env);

export const CONFIG = {
  ...env,
  PORT: env.PORT ?? 3000,
} as const;

// Other files import CONFIG, never use process.env directly
```

---

### Task 7: Check Zod Usage Patterns (If Zod is present)

**Step 1: Find validation results that are ignored**

Search for patterns where Zod validation is performed but the result isn't used:

```typescript
// BAD: Validating but using the original value
Schema.parse(value);
doSomething(value); // Should use the parse result!

// BAD: safeParse without using the result
const result = Schema.safeParse(value);
doSomething(value); // Should use result.data!
```

**Why this matters:** Calling Zod purely for side effects defeats the purpose. The parsed value IS the validated, typed value - using the original input means you're still working with unvalidated data.

Flag as **HIGH** priority - this is validation theater that provides false confidence.

**Recommended pattern:**

```typescript
// GOOD: Use the validated result
const result = Schema.safeParse(input);
if (!result.success) {
  throw new ValidationError(result.error);
}
const data = result.data; // Use this, not input

// OR
const data = Schema.parse(input); // Throws on invalid
```

---

### Task 8: Check for Type-Schema Drift

**Why this matters:** When a Zod schema and a separate TypeScript interface describe the same domain object but are defined independently, they can drift apart over time, causing subtle bugs.

**Step 1: Find potential drift**

Look for:
- Paired names like `UserSchema` and `User`, `UserDto` and `UserZodSchema`
- Manual interfaces that mirror Zod schemas by hand
- Zod schemas used for validation while code uses a different hand-maintained type

**Step 2: Identify drift indicators**

- Schema requires a property that the type marks as optional
- Type includes a property the schema doesn't validate
- Different nullability between schema and type
- Different union members

Flag as **MEDIUM** priority when types and schemas appear to describe the same thing but are maintained separately.

**Recommended pattern:**

```typescript
// BAD: Duplicate definitions that can drift
const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
});

interface User {  // Manually maintained - can drift!
  id: string;
  name: string;
  email?: string;  // Oops, schema doesn't have this
}

// GOOD: Derive type from schema
const UserSchema = z.object({
  id: z.string(),
  name: z.string(),
});

type User = z.infer<typeof UserSchema>;
```

---

### Task 9: Check Schema-Logic Alignment

**Why this matters:** Sometimes a Zod schema allows more states than the consuming code handles. This isn't a type error, but it's a design bug.

**Step 1: Find schema-logic mismatches**

Look for:
- Zod unions/enums where consuming code only handles some cases
- Optional fields in schema that code assumes are present
- Code with generic `else` branches that don't distinguish between states

```typescript
// Schema allows three states
const StatusSchema = z.enum(["pending", "approved", "rejected"]);

// But code only distinguishes two
if (status === "approved") {
  applyDiscount();
} else {
  // Bug: treats "pending" same as "rejected"
}
```

**Step 2: Check for unconstrained generics with external data**

Look for exported functions like:

```typescript
// BAD: Generic with no runtime constraint
function getConfig<T>(key: string): T {
  return storage[key] as T;  // Lies to the type system
}
```

**Recommended pattern:**

```typescript
// GOOD: Tie generic to Zod schema
function parseWith<T>(schema: z.ZodType<T>, input: unknown): T {
  return schema.parse(input);
}
```

Flag schema-logic mismatches as **MEDIUM** priority.

---

## Phase 5: Code Organization

### Task 10: Analyze Module Structure

**Step 1: Check for barrel exports**

Look for `index.ts` files that re-export. Evaluate if they:
- Cause circular dependencies
- Increase bundle size unnecessarily
- Make imports harder to trace

**Step 2: Check import organization**

Look for:
- Circular dependencies (imports that form loops)
- Deep relative imports (`../../../`)
- Missing path aliases that could simplify imports

**Step 3: Review file organization**

Check if:
- Related code is colocated
- Files are reasonably sized (flag files > 500 lines)
- Naming conventions are consistent

---

### Task 11: Review Export Patterns

**Step 1: Check for default exports**

Flag default exports as a **LOW** priority issue - named exports are generally preferred for:
- Better refactoring support
- Clearer imports
- Better tree-shaking

**Step 2: Check for unused exports**

Identify exported functions, types, or constants that aren't imported anywhere.

---

## Phase 6: Error Handling

### Task 12: Evaluate Error Handling

**Step 1: Check for empty catch blocks**

Search for:
```typescript
catch (e) {}
catch (_) {}
catch {
}
```

Empty catch blocks silently swallow errors - flag as **HIGH** priority.

**Step 2: Check error typing**

Look for:
```typescript
catch (e: any)
catch (error)  // implicit any
```

Errors should be properly typed or use `unknown` with type guards.

**Step 3: Review async error handling**

Check for:
- Unhandled promise rejections (missing `.catch()` or try/catch)
- Fire-and-forget async calls without error handling
- Missing error boundaries in React components (if applicable)

---

## Phase 7: Code Quality Patterns

### Task 13: Identify Anti-Patterns

**Step 1: Check for common anti-patterns**

Look for:

| Pattern | Issue | Severity |
|---------|-------|----------|
| `== null` or `!= null` | Use `=== null \|\| === undefined` or nullish coalescing | LOW |
| Nested ternaries | Hard to read, use if/else or early returns | MEDIUM |
| Magic numbers/strings | Should be named constants | LOW |
| Deep nesting (> 3 levels) | Extract functions or use early returns | MEDIUM |
| Long functions (> 50 lines) | Break into smaller functions | MEDIUM |
| Long parameter lists (> 4 params) | Use options object | LOW |

**Step 2: Check for code duplication**

Identify repeated code blocks that could be extracted into shared utilities.

---

### Task 14: Review Type Definitions

**Step 1: Check interface vs type usage**

Review consistency:
- Are interfaces used for object shapes?
- Are type aliases used for unions/intersections?
- Is there a consistent pattern throughout the codebase?

**Step 2: Check for overly complex types**

Flag types that are:
- Deeply nested (> 3 levels)
- Excessively long (> 20 properties without grouping)
- Using complex conditional types without documentation

**Step 3: Review generic usage**

Check that:
- Generics have meaningful names (not just `T` for complex cases)
- Constraints are properly applied
- Generics aren't overused where simple types would work

---

## Phase 8: Logic and Semantic Bugs

### Task 15: Find Logic Bugs That Pass Type Checking

**Why this matters:** Even with strict TypeScript AND Zod, logic can be wrong. Focus on business rules and invariants that types and schemas don't guarantee.

**Step 1: Check for inverted conditions**

Look for suspicious boolean logic:
- Conditions on flags, roles, or statuses that look inverted
- `!isPremium` where `isPremium` seems intended (or vice versa)
- Negations that don't match the variable/function name

**Step 2: Check array handling**

Look for:
- Array indexing without checking length (even after Zod validation)
- `array[0]` without checking if array is empty
- Off-by-one errors in loops or slices

**Step 3: Check exhaustiveness**

Look for:
- Switch statements on discriminated unions with `default` instead of exhaustive cases
- If/else chains that don't cover all enum values
- Generic `else` branches that should distinguish between states

**Recommended pattern for exhaustive switches:**

```typescript
function assertNever(x: never): never {
  throw new Error(`Unexpected value: ${x}`);
}

switch (status) {
  case "pending": return handlePending();
  case "approved": return handleApproved();
  case "rejected": return handleRejected();
  default: return assertNever(status); // Compile error if case missed
}
```

Flag logic bugs as **HIGH** priority when you can describe a concrete plausible bug scenario.

---

## Phase 9: Performance Considerations

### Task 16: Identify Performance Issues

**Step 1: Check for synchronous blocking operations**

Look for:
- `fs.readFileSync` and other sync file operations
- Large synchronous loops
- Blocking the event loop

**Step 2: Review data structures**

Check for:
- Arrays used where Sets would be more efficient (for lookups)
- Repeated array operations that could be combined
- Missing memoization for expensive computations

**Step 3: Check React-specific issues (if applicable)**

Look for:
- Missing `useMemo`/`useCallback` for expensive operations
- Inline object/array creation in JSX props
- Missing `key` props in lists
- Unnecessary re-renders from unstable references

---

## Phase 10: Dependency Analysis

### Task 17: Review Dependencies

**Files:** `package.json`

**Step 1: Check for outdated dependencies**

Note any dependencies that might be significantly outdated.

**Step 2: Check for security concerns**

Flag any dependencies known to have security issues.

**Step 3: Check for unnecessary dependencies**

Identify dependencies that:
- Duplicate functionality already in the codebase
- Are only used in one place and could be replaced with native code
- Have been abandoned or are unmaintained

---

## Phase 11: Summary Report

### Task 18: Generate Summary Report

After completing all phases, provide a summary that consolidates all issues.

**Step 1: Reprint all issues**

List every issue found, grouped by severity, using the same format as during the evaluation:

```
## All Issues

### Critical
[CRITICAL] Issue Title
📍 file/path.ts:123
💬 Brief description of the problem
🔧 Recommendation

### High
[HIGH] Issue Title
📍 file/path.ts:456
💬 Brief description of the problem
🔧 Recommendation

... (continue for Medium, Low)
```

**Step 2: Compile findings**

Reference these severity categories when classifying issues:

**Critical (Must Fix)**
- Custom crypto/security implementations (use established libraries)
- External data used without runtime validation
- Type safety holes that could cause runtime errors
- Security vulnerabilities
- Missing error handling that could crash the app

**High (Should Fix)**
- API SDK installed but unused (raw HTTP used instead, manual types will drift from actual API)
- Reinvented wheels for complex functionality (dates, validation, state machines)
- Strict mode violations
- Empty catch blocks
- Zod validation results ignored
- Logic bugs with concrete failure scenarios
- Significant code duplication

**Medium (Consider Fixing)**
- Installed dependencies not being used where they should be
- Type-schema drift
- Schema-logic mismatches
- Direct env var access without central config
- Code organization issues
- Complex functions that should be refactored

**Low (Nice to Have)**
- Style inconsistencies
- Minor anti-patterns
- Documentation gaps
- Default exports

**Step 3: Provide summary statistics**

After the full issue list, add:
```
## Statistics

Found X issues across Y files:
- Critical: N
- High: N
- Medium: N
- Low: N

Top 3 priorities:
1. [Most important issue to fix first]
2. [Second most important]
3. [Third most important]
```

**Step 4: Highlight positive patterns**

Note well-implemented patterns that should be continued:
- Good type safety practices
- Proper Zod validation at boundaries
- Clean code organization
- Effective error handling

---

## Phase 12: Save and Next Steps

### Task 19: Offer to Save Report

Ask the user: "Would you like me to save this report to `docs/issues-{date}.md`?"

If yes:
- Create the `docs/` directory if it doesn't exist
- Write the full summary report (all issues, statistics, positive patterns) to `docs/issues-YYYY-MM-DD.md` (e.g., `docs/issues-2025-01-15.md`)
- Include a timestamp at the top of the file

### Task 20: Offer Additional Analysis

Ask the user if they want deeper analysis on:
- Test coverage and test quality
- API design and contracts
- Documentation completeness
- Accessibility (for frontend code)
- Bundle size analysis
