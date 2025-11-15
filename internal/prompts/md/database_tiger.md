---
title: Add Tiger Cloud Database
description: Integrate PostgreSQL/TimescaleDB from Tiger Cloud with your Node.js application using Drizzle ORM
tags: [database, postgresql, timescaledb, tiger, drizzle, sql, persistence]
category: database
dependencies: []
related: [create_web_app, auth_jwt]
---

# Add Tiger Cloud Database

Integrate a PostgreSQL or TimescaleDB database from Tiger Cloud into your Node.js application with type-safe queries using Drizzle ORM.

## Overview

This guide shows you how to:
1. Provision a database using Tiger MCP
2. Set up Drizzle ORM in your Node.js application
3. Define schemas and run migrations
4. Integrate database queries with Fastify routes
5. Handle connections, errors, and transactions

**Prerequisites:** Existing Node.js/TypeScript application (see `create_web_app` template)

## Step 1: Provision Database with Tiger MCP

Use Tiger MCP to create a PostgreSQL or TimescaleDB database service.

**Conceptual steps:**
- Use Tiger MCP to provision a new database service
- Choose appropriate CPU/memory configuration
- Select a region close to your users
- Get the connection string

Tiger MCP will return a connection string like:
```
postgres://tsdbadmin:password@xxxxx.tsdb.cloud.timescale.com:12345/tsdb
```

Save this for Step 3.

## Step 2: Install Drizzle Dependencies

```bash
npm install drizzle-orm postgres dotenv
npm install -D drizzle-kit @types/node
```

**What these do:**
- `drizzle-orm` - TypeScript ORM with SQL-like queries
- `postgres` - Fast PostgreSQL client (alternative to node-postgres)
- `drizzle-kit` - CLI for migrations and introspection
- `dotenv` - Environment variable management

Use the `execute` tool with `run_command` operation:

```json
{
  "operation": "run_command",
  "params": {
    "command": "npm install drizzle-orm postgres dotenv && npm install -D drizzle-kit @types/node",
    "cwd": "/path/to/your/app"
  }
}
```

## Step 3: Configure Environment Variables

Add database connection string to `.env`:

```bash
DATABASE_URL="postgres://tsdbadmin:password@xxxxx.tsdb.cloud.timescale.com:12345/tsdb"
NODE_ENV="development"
```

Use the `execute` tool with `create_file` or `edit_file` operation.

**Security:** Never commit `.env` to version control. Ensure `.gitignore` includes `.env`.

## Step 4: Create Database Client

Create `src/db/client.ts`:

```typescript
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

// Connection configuration
const connectionString = process.env.DATABASE_URL;

if (!connectionString) {
  throw new Error('DATABASE_URL environment variable is required');
}

// Create postgres client with connection pooling
export const client = postgres(connectionString, {
  max: 10, // Maximum pool size
  idle_timeout: 20,
  connect_timeout: 10,
});

// Create Drizzle instance
export const db = drizzle(client, { schema });

// Graceful shutdown
export async function closeDatabase() {
  await client.end();
}
```

## Step 5: Define Database Schema

### Option A: SQL-First Approach (Recommended)

**Use this when:** You want to leverage Tiger MCP's excellent schema design templates.

1. Use Tiger MCP's schema design templates to create your tables
2. Execute the SQL via Tiger MCP's query execution
3. Use `drizzle-kit introspect` to generate TypeScript schema

**Example workflow:**
- Use Tiger MCP to get schema design guidance
- Execute `CREATE TABLE` statements via Tiger MCP
- Run `npx drizzle-kit introspect` to generate `src/db/schema.ts`

### Option B: Drizzle Schema-First

**Use this when:** You want type-first development or simpler workflow.

Create `src/db/schema.ts`:

```typescript
import { pgTable, bigserial, text, timestamp, numeric, index } from 'drizzle-orm/pg-core';

// Example: Users table
export const users = pgTable('users', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  email: text('email').notNull().unique(),
  name: text('name').notNull(),
  createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
}, (table) => ({
  emailIdx: index('users_email_idx').on(table.email),
  createdAtIdx: index('users_created_at_idx').on(table.createdAt),
}));

// Example: Orders table
export const orders = pgTable('orders', {
  id: bigserial('id', { mode: 'number' }).primaryKey(),
  userId: bigserial('user_id', { mode: 'number' }).notNull().references(() => users.id),
  total: numeric('total', { precision: 10, scale: 2 }).notNull(),
  status: text('status').notNull().default('pending'),
  createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
}, (table) => ({
  userIdIdx: index('orders_user_id_idx').on(table.userId),
  createdAtIdx: index('orders_created_at_idx').on(table.createdAt),
}));
```

**Drizzle type mappings:**
- `bigserial` → `BIGINT GENERATED ALWAYS AS IDENTITY`
- `text` → `TEXT`
- `timestamp(..., { withTimezone: true })` → `TIMESTAMPTZ`
- `numeric(p, s)` → `NUMERIC(p, s)`

Then push schema to database:

```bash
npx drizzle-kit push
```

## Step 6: Configure Drizzle Kit

Create `drizzle.config.ts` in project root:

```typescript
import { defineConfig } from 'drizzle-kit';
import * as dotenv from 'dotenv';

dotenv.config();

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
});
```

## Step 7: Integrate with Fastify

Update `src/server.ts` to include database connection:

```typescript
import Fastify from 'fastify';
import { db, closeDatabase } from './db/client';

const server = Fastify({ logger: true });

// Make db available in routes
server.decorate('db', db);

// Graceful shutdown
server.addHook('onClose', async () => {
  await closeDatabase();
});

// Health check with database
server.get('/health', async (request, reply) => {
  try {
    // Simple query to verify database connection
    await db.execute('SELECT 1');
    return { status: 'ok', database: 'connected' };
  } catch (error) {
    reply.status(503);
    return { status: 'error', database: 'disconnected' };
  }
});

export default server;
```

Add TypeScript declaration for the decorator in `src/types/fastify.d.ts`:

```typescript
import { PostgresJsDatabase } from 'drizzle-orm/postgres-js';
import * as schema from '../db/schema';

declare module 'fastify' {
  interface FastifyInstance {
    db: PostgresJsDatabase<typeof schema>;
  }
}
```

## Step 8: Create CRUD Routes

Create `src/routes/users.ts`:

```typescript
import { FastifyPluginAsync } from 'fastify';
import { eq } from 'drizzle-orm';
import { users } from '../db/schema';
import { z } from 'zod';

// Validation schemas
const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});

const userIdSchema = z.object({
  id: z.coerce.number().int().positive(),
});

const usersRoutes: FastifyPluginAsync = async (server) => {
  // Create user
  server.post('/users', async (request, reply) => {
    try {
      const data = createUserSchema.parse(request.body);

      const [newUser] = await server.db
        .insert(users)
        .values(data)
        .returning();

      reply.status(201).send(newUser);
    } catch (error) {
      if (error instanceof z.ZodError) {
        reply.status(400).send({ error: 'Invalid input', details: error.errors });
        return;
      }

      // Handle unique constraint violation
      if (error.code === '23505') {
        reply.status(409).send({ error: 'Email already exists' });
        return;
      }

      server.log.error(error);
      reply.status(500).send({ error: 'Internal server error' });
    }
  });

  // List users
  server.get('/users', async (request, reply) => {
    try {
      const allUsers = await server.db
        .select()
        .from(users)
        .orderBy(users.createdAt);

      reply.send(allUsers);
    } catch (error) {
      server.log.error(error);
      reply.status(500).send({ error: 'Internal server error' });
    }
  });

  // Get user by ID
  server.get('/users/:id', async (request, reply) => {
    try {
      const { id } = userIdSchema.parse(request.params);

      const [user] = await server.db
        .select()
        .from(users)
        .where(eq(users.id, id))
        .limit(1);

      if (!user) {
        reply.status(404).send({ error: 'User not found' });
        return;
      }

      reply.send(user);
    } catch (error) {
      if (error instanceof z.ZodError) {
        reply.status(400).send({ error: 'Invalid user ID' });
        return;
      }

      server.log.error(error);
      reply.status(500).send({ error: 'Internal server error' });
    }
  });

  // Update user
  server.patch('/users/:id', async (request, reply) => {
    try {
      const { id } = userIdSchema.parse(request.params);
      const data = createUserSchema.partial().parse(request.body);

      const [updatedUser] = await server.db
        .update(users)
        .set(data)
        .where(eq(users.id, id))
        .returning();

      if (!updatedUser) {
        reply.status(404).send({ error: 'User not found' });
        return;
      }

      reply.send(updatedUser);
    } catch (error) {
      if (error instanceof z.ZodError) {
        reply.status(400).send({ error: 'Invalid input', details: error.errors });
        return;
      }

      server.log.error(error);
      reply.status(500).send({ error: 'Internal server error' });
    }
  });

  // Delete user
  server.delete('/users/:id', async (request, reply) => {
    try {
      const { id } = userIdSchema.parse(request.params);

      const [deletedUser] = await server.db
        .delete(users)
        .where(eq(users.id, id))
        .returning();

      if (!deletedUser) {
        reply.status(404).send({ error: 'User not found' });
        return;
      }

      reply.status(204).send();
    } catch (error) {
      server.log.error(error);
      reply.status(500).send({ error: 'Internal server error' });
    }
  });
};

export default usersRoutes;
```

Register routes in `src/server.ts`:

```typescript
import usersRoutes from './routes/users';

// After other setup
server.register(usersRoutes);
```

## Step 9: Transaction Handling

For operations requiring multiple queries, use transactions:

```typescript
import { db } from '../db/client';
import { users, orders } from '../db/schema';

// Example: Create user and initial order in transaction
server.post('/users-with-order', async (request, reply) => {
  try {
    const result = await db.transaction(async (tx) => {
      // Insert user
      const [newUser] = await tx
        .insert(users)
        .values({ email: 'test@example.com', name: 'Test User' })
        .returning();

      // Insert order for user
      const [newOrder] = await tx
        .insert(orders)
        .values({
          userId: newUser.id,
          total: '99.99',
          status: 'pending',
        })
        .returning();

      return { user: newUser, order: newOrder };
    });

    reply.status(201).send(result);
  } catch (error) {
    server.log.error(error);
    reply.status(500).send({ error: 'Transaction failed' });
  }
});
```

**Transaction behavior:**
- All queries succeed together or all fail together
- Automatic rollback on error
- Essential for data consistency

## Step 10: Advanced Queries

### Joins

```typescript
import { eq } from 'drizzle-orm';

const usersWithOrders = await db
  .select({
    userId: users.id,
    userName: users.name,
    orderId: orders.id,
    orderTotal: orders.total,
  })
  .from(users)
  .leftJoin(orders, eq(users.id, orders.userId))
  .where(eq(users.id, 1));
```

### Aggregations

```typescript
import { count, sum } from 'drizzle-orm';

const orderStats = await db
  .select({
    userId: orders.userId,
    totalOrders: count(orders.id),
    totalSpent: sum(orders.total),
  })
  .from(orders)
  .groupBy(orders.userId);
```

### Time-based queries (TimescaleDB)

```typescript
import { sql, gte } from 'drizzle-orm';

// Last 24 hours
const recentOrders = await db
  .select()
  .from(orders)
  .where(gte(orders.createdAt, sql`NOW() - INTERVAL '24 hours'`));
```

### Raw SQL (for advanced TimescaleDB features)

```typescript
import { sql } from 'drizzle-orm';

// Use time_bucket for aggregations
const hourlySales = await db.execute(sql`
  SELECT
    time_bucket('1 hour', created_at) AS hour,
    COUNT(*) as order_count,
    SUM(total) as total_sales
  FROM orders
  WHERE created_at >= NOW() - INTERVAL '7 days'
  GROUP BY hour
  ORDER BY hour DESC
`);
```

## TimescaleDB Integration

For time-series workloads (metrics, events, logs), use Tiger MCP's hypertable templates:

1. Use Tiger MCP to get the `setup_hypertable` template
2. Follow the guidance to create hypertables with compression and continuous aggregates
3. Use Drizzle for application queries, raw SQL for TimescaleDB-specific features

**When to use hypertables:**
- High-volume insert workloads (1M+ rows)
- Time-based queries (last hour, last day, last month)
- Need compression (90%+ storage savings)
- Aggregation queries (dashboards, analytics)

## Testing the Integration

```typescript
// src/db/client.test.ts
import { db, closeDatabase } from './client';
import { users } from './schema';

async function testConnection() {
  try {
    // Test basic query
    const result = await db.select().from(users).limit(1);
    console.log('✓ Database connection successful');

    // Test insert
    const [newUser] = await db
      .insert(users)
      .values({ email: 'test@example.com', name: 'Test' })
      .returning();
    console.log('✓ Insert successful:', newUser);

    // Cleanup
    await db.delete(users).where(eq(users.id, newUser.id));
    console.log('✓ Delete successful');

  } catch (error) {
    console.error('✗ Database test failed:', error);
  } finally {
    await closeDatabase();
  }
}

testConnection();
```

Run with:
```bash
npx tsx src/db/client.test.ts
```

## Common Patterns

### Error Handling

```typescript
// Utility function for consistent error handling
function handleDatabaseError(error: any, reply: FastifyReply) {
  // Unique constraint violation
  if (error.code === '23505') {
    reply.status(409).send({ error: 'Resource already exists' });
    return;
  }

  // Foreign key violation
  if (error.code === '23503') {
    reply.status(400).send({ error: 'Referenced resource not found' });
    return;
  }

  // Not null violation
  if (error.code === '23502') {
    reply.status(400).send({ error: 'Required field missing' });
    return;
  }

  // Generic error
  server.log.error(error);
  reply.status(500).send({ error: 'Internal server error' });
}
```

### Pagination

```typescript
import { asc } from 'drizzle-orm';

const page = 1;
const pageSize = 20;

const paginatedUsers = await db
  .select()
  .from(users)
  .orderBy(asc(users.createdAt))
  .limit(pageSize)
  .offset((page - 1) * pageSize);
```

### Connection Health Check

```typescript
// Periodically verify connection
setInterval(async () => {
  try {
    await db.execute(sql`SELECT 1`);
  } catch (error) {
    server.log.error('Database health check failed:', error);
  }
}, 30000); // Every 30 seconds
```

## Migration Workflow

### Generate migration from schema changes

```bash
npx drizzle-kit generate
```

### Apply migrations

```bash
npx drizzle-kit migrate
```

### View current schema

```bash
npx drizzle-kit studio
```

This opens a web UI at `https://local.drizzle.studio` to browse your database.

## Troubleshooting

### Connection issues

- Verify `DATABASE_URL` is correct
- Check firewall rules allow connection to Tiger Cloud
- Ensure database service is running (use Tiger MCP to check)

### Type errors

- Run `npx drizzle-kit generate` after schema changes
- Restart TypeScript server in your IDE

### Slow queries

- Add indexes on frequently queried columns
- Use Tiger MCP's query analysis tools
- Consider continuous aggregates for TimescaleDB

### Migration conflicts

- Always pull latest schema before making changes
- Use transactions in migrations for safety
- Test migrations on non-production data first

## Next Steps

- **Authentication:** Add user authentication with `auth_jwt` template
- **Advanced queries:** Explore Drizzle's query builder features
- **Monitoring:** Set up query logging and performance tracking
- **Scaling:** Configure read replicas via Tiger MCP

## Useful Resources

- Drizzle ORM Documentation: https://orm.drizzle.team/docs/overview
- Tiger MCP schema templates: Use `get_prompt_template` for PostgreSQL and TimescaleDB guidance
- Drizzle Kit CLI: https://orm.drizzle.team/kit-docs/overview
