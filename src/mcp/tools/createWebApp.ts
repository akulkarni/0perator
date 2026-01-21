import { exec } from "node:child_process";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import { writeAppTemplates } from "../../lib/templates.js";
import type { ServerContext } from "../../types.js";

const execAsync = (cmd: string) =>
  new Promise<{ stdout: string; stderr: string }>((resolve, reject) => {
    exec(cmd, (error, stdout, stderr) => {
      if (error) reject(Object.assign(error, { stderr }));
      else resolve({ stdout, stderr });
    });
  });

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
        // Copy app template
        await writeAppTemplates(appName, {
          app_name: appName,
          use_auth,
          product_brief,
          future_features,
        });

        // Install dependencies
        await execAsync(`cd ${appName} && npm install`);

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
