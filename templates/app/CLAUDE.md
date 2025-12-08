# Todo App - Project Guide

## Overview

Full-stack [[.Name]] management app built with the T3 Stack (Next.js 15, tRPC, Drizzle ORM, Better Auth).

## Status

The app is now in development and has not been deployed to production yet

## Tech Stack

- **Frontend**: Next.js 15.2.3 (App Router), React 19, TypeScript, Tailwind CSS 4
- **Backend**: tRPC v11, Better Auth v1.3
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
│   │   ├── auth/[...all]/ # Better Auth handler
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
│   └── better-auth/       # Auth configuration
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
- `protectedProcedure`: Requires valid session (throws UNAUTHORIZED)

```typescript
// In routers
export const postRouter = createTRPCRouter({
  getAll: publicProcedure.query(async ({ ctx }) => { ... }),
  create: protectedProcedure
    .input(z.object({ title: z.string().min(1) }))
    .mutation(async ({ ctx, input }) => { ... }),
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

### Authentication
Session available in tRPC context for protected procedures:
```typescript
// ctx.session contains user info in protectedProcedure
const userId = ctx.session.user.id;
```


## Development Notes

- Dev server has artificial 100-500ms delay to catch data waterfalls
- Database connection is cached in dev to avoid HMR reconnection issues
- Use `db:push` for quick schema iteration, `db:migrate` for production

## Styling

- Tailwind CSS with CSS variables for theming
- Use `cn()` utility from `~/lib/utils` for conditional classes
- Dark mode supported via `dark:` prefix

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
