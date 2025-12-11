# {{app_name}} - Project Guide

## Overview

Full-stack {{app_name}} app built with the T3 Stack (Next.js 15, tRPC, Drizzle ORM{{#if use_auth}}, Better Auth{{/if}}).

{{#if product_brief}}
## Product Brief

{{{product_brief}}}

{{/if}}
{{#if future_features}}
## Future Features

{{{future_features}}}

{{/if}}
## Status

The app is now in development and has not been deployed to production yet

## Tech Stack

- **Frontend**: Next.js 15.2.3 (App Router), React 19, TypeScript, Tailwind CSS 4
- **Backend**: tRPC v11{{#if use_auth}}, Better Auth v1.3{{/if}}
- **Database**: PostgreSQL with Drizzle ORM v0.41
- **State Management**: TanStack Query (React Query) v5

## Commands

```bash
npm run dev          # Start dev server with Turbo (localhost:3000)
npm run build        # Production build
npm run typecheck    # Type check without emitting
npm run db:generate  # Generate Drizzle migrations (must use this for production-deployed apps)
npm run db:migrate   # Run pending migrations (must use this for production-deployed apps)
npm run db:push      # Push schema changes (only do this while the app hasn't been deployed to production)
npm run db:studio    # Open Drizzle Studio UI
```

## Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── api/
{{#if use_auth}}
│   │   ├── auth/[...all]/ # Better Auth handler
{{/if}}
│   │   └── trpc/[trpc]/   # tRPC endpoint
│   ├── _components/       # Client components
│   ├── page.tsx           # Home page (Server Component)
│   └── layout.tsx         # Root layout with providers
├── server/
│   ├── api/
│   │   ├── root.ts        # tRPC router aggregation
│   │   ├── trpc.ts        # tRPC context & procedures
│   │   └── routers/       # tRPC sub-routers
│   ├── db/
│   │   ├── schema.ts      # Drizzle schema definitions
│   │   └── index.ts       # Database connection
{{#if use_auth}}
│   └── better-auth/       # Auth configuration
{{/if}}
├── trpc/
│   ├── server.ts          # RSC tRPC helpers
│   ├── react.tsx          # Client tRPC provider
│   └── query-client.ts    # React Query config
├── lib/utils.ts           # Utility functions (cn)
└── env.js                 # Environment validation
```

## Key Patterns

### Import Aliases
Use `~/` for all imports from `src/`:
```typescript
import { db } from "~/server/db";
import { api } from "~/trpc/server";
```

### Server vs Client Components
- Pages are Server Components by default
- Mark client components with `"use client"` directive
- Client components go in `_components/` directory

### tRPC Procedures
- `publicProcedure`: No authentication required
{{#if use_auth}}
- `protectedProcedure`: Requires valid session (throws UNAUTHORIZED)
{{/if}}

```typescript
// In routers
export const postRouter = createTRPCRouter({
  getAll: publicProcedure.query(async ({ ctx }) => { ... }),
{{#if use_auth}}
  create: protectedProcedure
    .input(z.object({ title: z.string().min(1) }))
    .mutation(async ({ ctx, input }) => { ... }),
{{/if}}
});
```

### Database Queries
Use Drizzle ORM with the schema from `~/server/db/schema`:
```typescript
import { db } from "~/server/db";
import { todos } from "~/server/db/schema";

// Query
await db.query.todos.findMany({ where: eq(todos.createdById, userId) });

// Insert
await db.insert(todos).values({ title, createdById: userId });
```

{{#if use_auth}}
### Authentication
Session available in tRPC context for protected procedures:
```typescript
// ctx.session contains user info in protectedProcedure
const userId = ctx.session.user.id;
```

{{/if}}
## Development Notes

- Dev server has artificial 100-500ms delay to catch data waterfalls
- Database connection is cached in dev to avoid HMR reconnection issues
- Use `db:push` for quick schema iteration, `db:migrate` for production

## Styling

- Tailwind CSS with CSS variables for theming
- Use `cn()` utility from `~/lib/utils` for conditional classes
- Dark mode supported via `dark:` prefix

### shadcn/ui Components

This project uses [shadcn/ui](https://ui.shadcn.com/) for UI components. Components are installed to `src/components/ui/`.

**Adding new components:**
```bash
npx shadcn@latest add button card input form table
```

**Using components:**
```typescript
import { Button } from "~/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "~/components/ui/card";
```

**Color tokens:** Use semantic shadcn colors instead of hardcoded Tailwind colors:
- `bg-background` / `text-foreground` - main bg/text
- `bg-muted` / `text-muted-foreground` - secondary bg/text
- `bg-card` / `text-card-foreground` - card surfaces
- `bg-primary` / `text-primary-foreground` - primary actions
- `bg-destructive` / `text-destructive-foreground` - destructive actions
- `border-border` - borders
- `ring-ring` - focus rings

Browse available components at https://ui.shadcn.com/docs/components

## Adding New Features

### New tRPC Router
1. Create router in `src/server/api/routers/`
2. Add to `appRouter` in `src/server/api/root.ts`

### New Database Table
1. Add schema in `src/server/db/schema.ts`
2. Run `npm run db:generate` then `npm run db:migrate`

### New Page
1. Create `page.tsx` in `src/app/[route]/`
2. Use Server Components for data fetching
3. Create client components in `_components/` as needed

## Type Safety

- tRPC provides end-to-end type inference
- Zod schemas validate inputs at runtime
- Environment variables validated via `@t3-oss/env-nextjs`
