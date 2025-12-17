import { exec } from "node:child_process";
import { unlink } from "node:fs/promises";
import { join } from "node:path";
import { promisify } from "node:util";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import { writeAppTemplates } from "../../lib/templates.js";
import type { ServerContext } from "../../types.js";

const execAsync = promisify(exec);

const inputSchema = {
  app_name: z.string().describe("Application name"),
  use_auth: z.boolean().default(false).describe("Enable authentication"),
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
} as const;

const outputSchema = {
  success: z.boolean().describe("Whether the app was created successfully"),
  message: z.string().describe("Status message"),
  path: z.string().optional().describe("Path to created app"),
} as const;

type OutputSchema = {
  success: boolean;
  message: string;
  path?: string;
};

export const createWebAppFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: "create_web_app",
    config: {
      title: "Create Web App",
      description:
        "🚀 Create any web application - Build an opinionated next.js app. Get instructions for how to use this using the create-app skill.",
      inputSchema,
      outputSchema,
    },
    fn: async ({
      app_name,
      use_auth,
      product_brief,
      future_features,
    }): Promise<OutputSchema> => {
      const appName = app_name;

      try {
        // Create T3 app
        const t3Args = [
          "npx",
          "create-t3-app@latest",
          appName,
          "--noInstall", //avoids dependency conflicts that could result
          "--noGit",
          "--CI",
          "--tailwind",
          "--drizzle",
          "--trpc",
          "--dbProvider",
          "postgres",
          "--appRouter",
          "--biome",
        ];
        if (use_auth) {
          t3Args.push("--betterAuth");
        }

        await execAsync(t3Args.join(" "));

        // Remove start-database script if it exists
        try {
          await unlink(join(appName, "start-database.sh"));
        } catch {
          // Ignore if file doesn't exist
        }

        // Copy app templates (CLAUDE.md, globals.css, etc.)
        await writeAppTemplates(appName, {
          app_name: appName,
          use_auth,
          product_brief,
          future_features,
        });

        return {
          success: true,
          message: `Created app '${appName}'`,
          path: appName,
        };
      } catch (err) {
        const error = err as Error & { stderr?: string };
        return {
          success: false,
          message: `Failed to create app: ${error.message}\n${error.stderr || ""}`,
        };
      }
    },
  };
};
