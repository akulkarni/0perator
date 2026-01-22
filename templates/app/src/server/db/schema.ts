import { pgSchema } from "drizzle-orm/pg-core";
import { env } from "~/env";

/**
 * Use a dedicated PostgreSQL schema for this app's tables.
 * @see https://orm.drizzle.team/docs/schemas
 */
export const schema = pgSchema(env.DATABASE_SCHEMA);

/**
 * Helper to create tables in the app's schema.
 * Always use this instead of pgTable to ensure tables are in the correct schema.
 */
const createTable = schema.table;
