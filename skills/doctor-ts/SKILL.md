---
name: doctor-ts
description: 'Step-by-step instructions for evaluating TypeScript code for quality, identifying issues, and providing actionable recommendations'
---

# TypeScript Code Quality Evaluation

> **For Claude:** Follow this checklist systematically. Report findings with specific file paths and line numbers. Prioritize issues by severity.

**Goal:** Evaluate a TypeScript codebase for quality issues, identify problems, and provide actionable recommendations.

**Scope:** Type safety, code organization, error handling, performance, maintainability, and best practices.

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
- `package.json` - Dependencies and scripts
- `.eslintrc.*` or `eslint.config.*` - Existing lint rules
- `biome.json` - Biome configuration

---

## Phase 2: Type Safety Analysis

### Task 2: Check TypeScript Strictness

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

### Task 3: Find Type Safety Violations

**Step 1: Search for `any` usage**

Search for explicit `any` types:
```
: any
as any
<any>
```

Categorize findings:
- **Legitimate:** Third-party library types, truly dynamic data
- **Avoidable:** Lazy typing that should use proper types

**Step 2: Search for type assertions**

Look for potentially unsafe type assertions:
```
as unknown as
as SomeType
!
```

Flag non-null assertions (`!`) that aren't justified by context.

**Step 3: Check for `@ts-ignore` and `@ts-expect-error`**

Search for TypeScript directive comments:
```
@ts-ignore
@ts-expect-error
@ts-nocheck
```

Each should have a comment explaining why it's necessary.

---

## Phase 3: Code Organization

### Task 4: Analyze Module Structure

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

### Task 5: Review Export Patterns

**Step 1: Check for default exports**

Flag default exports as a **LOW** priority issue - named exports are generally preferred for:
- Better refactoring support
- Clearer imports
- Better tree-shaking

**Step 2: Check for unused exports**

Identify exported functions, types, or constants that aren't imported anywhere.

---

## Phase 4: Error Handling

### Task 6: Evaluate Error Handling

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

## Phase 5: Code Quality Patterns

### Task 7: Identify Anti-Patterns

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

### Task 8: Review Type Definitions

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

## Phase 6: Performance Considerations

### Task 9: Identify Performance Issues

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

## Phase 7: Dependency Analysis

### Task 10: Review Dependencies

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

## Phase 8: Report Generation

### Task 11: Generate Quality Report

**Step 1: Compile findings**

Organize issues by severity:

**Critical (Must Fix)**
- Type safety holes that could cause runtime errors
- Security vulnerabilities
- Missing error handling that could crash the app

**High (Should Fix)**
- Strict mode violations
- Empty catch blocks
- Significant code duplication

**Medium (Consider Fixing)**
- Code organization issues
- Complex functions that should be refactored
- Missing type annotations in key areas

**Low (Nice to Have)**
- Style inconsistencies
- Minor anti-patterns
- Documentation gaps

**Step 2: Provide actionable recommendations**

For each issue, provide:
1. File path and line number
2. Description of the problem
3. Specific recommendation for fixing
4. Example of the fix (when helpful)

**Step 3: Highlight positive patterns**

Note well-implemented patterns that should be continued:
- Good type safety practices
- Clean code organization
- Effective error handling

---

## Phase 9: Optional Deep Dives

### Task 12: Offer Additional Analysis

Ask the user if they want deeper analysis on:
- Test coverage and test quality
- API design and contracts
- Documentation completeness
- Accessibility (for frontend code)
- Bundle size analysis
