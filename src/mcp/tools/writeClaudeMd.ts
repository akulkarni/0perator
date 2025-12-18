import { resolve } from "node:path";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import { writeClaudeMdTemplate } from "../../lib/templates.js";
import type { ServerContext } from "../../types.js";

const inputSchema = {
  application_directory: z
    .string()
    .describe("Path to the application directory"),
  app_name: z.string().describe("Application name"),
  use_auth: z.boolean().default(false).describe("Whether auth is enabled"),
  product_brief: z
    .string()
    .optional()
    .describe("Description of the product and minimal features for v0/demo"),
  future_features: z
    .string()
    .optional()
    .describe(
      "Features deferred to later that may affect architectural decisions",
    ),
  db_schema: z
    .string()
    .optional()
    .describe("Database schema name (from setup_app_schema)"),
  db_user: z
    .string()
    .optional()
    .describe("Database user name (from setup_app_schema)"),
} as const;

const outputSchema = {
  success: z.boolean().describe("Whether CLAUDE.md was created successfully"),
  message: z.string().describe("Status message"),
} as const;

type OutputSchema = {
  success: boolean;
  message: string;
};

export const writeClaudeMdFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: "write_claude_md",
    config: {
      title: "Write CLAUDE.md",
      description:
        "📝 Generate the CLAUDE.md project guide file for a scaffolded app. Call this at the end of app setup after all configuration is complete.",
      inputSchema,
      outputSchema,
    },
    fn: async ({
      application_directory,
      app_name,
      use_auth,
      product_brief,
      future_features,
      db_schema,
      db_user,
    }): Promise<OutputSchema> => {
      const appDir = resolve(process.cwd(), application_directory);
      try {
        await writeClaudeMdTemplate(appDir, {
          app_name,
          use_auth,
          product_brief,
          future_features,
          db_schema,
          db_user,
        });

        return {
          success: true,
          message: `Created CLAUDE.md for '${app_name}'`,
        };
      } catch (err) {
        const error = err as Error;
        return {
          success: false,
          message: `Failed to create CLAUDE.md: ${error.message}`,
        };
      }
    },
  };
};
