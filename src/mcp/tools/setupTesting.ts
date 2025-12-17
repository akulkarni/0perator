import { exec } from "node:child_process";
import { existsSync } from "node:fs";
import { writeFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import { promisify } from "node:util";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import postgres from "postgres";
import { z } from "zod";
import { writeTestingTemplates } from "../../lib/templates.js";
import type { ServerContext } from "../../types.js";

const execAsync = promisify(exec);

const inputSchema = {
  application_directory: z
    .string()
    .describe("Path to the application directory"),
  service_id: z.string().describe("Tiger Cloud service ID for the database"),
  schema_name: z
    .string()
    .regex(
      /^[a-z][a-z0-9_]*$/,
      "Schema name must be lowercase alphanumeric with underscores, starting with a letter.",
    )
    .describe(
      "Name of the test schema to create. Should be the name of the app schema prefixed with test_",
    ),
  test_user: z
    .string()
    .regex(
      /^[a-z][a-z0-9_]*$/,
      "User name must be lowercase alphanumeric with underscores, starting with a letter.",
    )
    .describe(
      "Name of the test database user. Should be the name of the app user prefixed with test_",
    ),
} as const;

const outputSchema = {
  success: z.boolean().describe("Whether test schema setup succeeded"),
  message: z.string().describe("Status message"),
  schema_name: z.string().optional().describe("Name of the created schema"),
  test_user: z.string().optional().describe("Name of the created test user"),
} as const;

type OutputSchema = {
  success: boolean;
  message: string;
  schema_name?: string | undefined;
  test_user?: string | undefined;
};

function generatePassword(length = 24): string {
  const chars =
    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  let password = "";
  for (let i = 0; i < length; i++) {
    password += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return password;
}

function buildTestConnectionString(
  originalUrl: string,
  testUser: string,
  testPassword: string,
): string {
  const parsed = new URL(originalUrl);
  parsed.username = testUser;
  parsed.password = encodeURIComponent(testPassword);
  return parsed.toString();
}

export const setupTestingFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: "setup_testing",
    config: {
      title: "Setup Testing",
      description:
        "🧪 Set up integration testing infrastructure. Creates an isolated PostgreSQL schema and user, copies Vitest config and test setup files, and writes DATABASE_URL to .env.test.local.",
      inputSchema,
      outputSchema,
    },
    fn: async ({
      application_directory,
      service_id,
      schema_name,
      test_user,
    }): Promise<OutputSchema> => {
      const appDir = resolve(process.cwd(), application_directory);
      const envTestPath = join(appDir, ".env.test.local");

      // Check if .env.test.local already exists
      if (existsSync(envTestPath)) {
        return {
          success: true,
          message: `.env.test.local already exists. Delete it and re-run if you need to regenerate.`,
          schema_name,
          test_user,
        };
      }

      // Get database connection string from Tiger
      let adminConnectionString: string;
      try {
        const { stdout: serviceJson } = await execAsync(
          `tiger service get ${service_id} --with-password -o json`,
        );
        const serviceDetails = JSON.parse(serviceJson) as {
          connection_string?: string;
        };

        if (!serviceDetails.connection_string) {
          return {
            success: false,
            message: "connection_string not found in service details",
          };
        }
        adminConnectionString = serviceDetails.connection_string;
      } catch (err) {
        const error = err as Error;
        return {
          success: false,
          message: `Failed to get service details: ${error.message}`,
        };
      }

      // Connect using postgres.js
      const sql = postgres(adminConnectionString);

      try {
        // Create test schema
        await sql.unsafe(`CREATE SCHEMA IF NOT EXISTS ${schema_name}`);

        // Check if user already exists
        const existingUser = await sql`
          SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = ${test_user}
        `;

        if (existingUser.length > 0) {
          await sql.end();
          return {
            success: false,
            message: `Test user '${test_user}' already exists but .env.test.local does not. Either create .env.test.local manually with the correct DATABASE_URL, or use a different test_user name.`,
          };
        }

        // Create new user
        const testPassword = generatePassword();
        await sql.unsafe(
          `CREATE ROLE ${test_user} WITH LOGIN PASSWORD '${testPassword}'`,
        );

        // Revoke access to public schema
        await sql.unsafe(`REVOKE ALL ON SCHEMA public FROM ${test_user}`);

        // Grant permissions to test user
        await sql.unsafe(
          `GRANT USAGE ON SCHEMA ${schema_name} TO ${test_user}`,
        );
        await sql.unsafe(
          `GRANT CREATE ON SCHEMA ${schema_name} TO ${test_user}`,
        );
        await sql.unsafe(
          `GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ${schema_name} TO ${test_user}`,
        );
        await sql.unsafe(
          `GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ${schema_name} TO ${test_user}`,
        );
        await sql.unsafe(
          `ALTER DEFAULT PRIVILEGES IN SCHEMA ${schema_name} GRANT ALL ON TABLES TO ${test_user}`,
        );
        await sql.unsafe(
          `ALTER DEFAULT PRIVILEGES IN SCHEMA ${schema_name} GRANT ALL ON SEQUENCES TO ${test_user}`,
        );

        // Set search_path for test user
        await sql.unsafe(
          `ALTER ROLE ${test_user} SET search_path TO ${schema_name}`,
        );

        await sql.end();

        // Build test connection string and write to .env.test.local
        const testDatabaseUrl = buildTestConnectionString(
          adminConnectionString,
          test_user,
          testPassword,
        );
        const envTestContent = `# Test environment - uses isolated schema '${schema_name}'\n# Generated by setup_testing - do not commit this file\nDATABASE_URL="${testDatabaseUrl}"\nDATABASE_SCHEMA="${schema_name}"\n`;
        await writeFile(envTestPath, envTestContent);

        // Copy testing template files (vitest.config.ts, src/test/global-setup.ts)
        await writeTestingTemplates(appDir);
      } catch (err) {
        await sql.end();
        const error = err as Error;
        return {
          success: false,
          message: `Failed to set up testing: ${error.message}`,
        };
      }

      return {
        success: true,
        message: `Created test schema '${schema_name}' and user '${test_user}'. Vitest config and .env.test.local written.`,
        schema_name,
        test_user,
      };
    },
  };
};
