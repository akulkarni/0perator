import { exec } from "node:child_process";
import crypto from "node:crypto";
import { appendFile, readFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import { promisify } from "node:util";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import { writeBetterAuthTemplates } from "../../lib/templates.js";
import type { ServerContext } from "../../types.js";

const execAsync = promisify(exec);

const inputSchema = {
  application_directory: z
    .string()
    .describe("Path to the application directory"),
} as const;

const outputSchema = {
  success: z.boolean().describe("Whether Better Auth setup succeeded"),
  message: z.string().describe("Status message"),
} as const;

type OutputSchema = {
  success: boolean;
  message: string;
};

function generateSecret(): string {
  return Buffer.from(crypto.getRandomValues(new Uint8Array(32))).toString(
    "base64",
  );
}

export const setupBetterAuthFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: "setup_better_auth",
    config: {
      title: "Setup Better Auth",
      description:
        "🔐 Set up Better Auth authentication. Copies Better Auth config, client, server helpers, API route, tRPC integration files, and generates BETTER_AUTH_SECRET in .env.",
      inputSchema,
      outputSchema,
    },
    fn: async ({ application_directory }): Promise<OutputSchema> => {
      const appDir = resolve(process.cwd(), application_directory);
      const envPath = join(appDir, ".env");

      try {
        await writeBetterAuthTemplates(appDir);

        // Install better-auth package
        await execAsync("npm install better-auth", { cwd: appDir });

        // Check if BETTER_AUTH_SECRET already exists in .env
        let envContent = "";
        try {
          envContent = await readFile(envPath, "utf-8");
        } catch {
          // .env doesn't exist yet, that's fine
        }

        if (!envContent.includes("BETTER_AUTH_SECRET")) {
          const secret = generateSecret();
          await appendFile(
            envPath,
            `\n# Better Auth\nBETTER_AUTH_SECRET="${secret}"\n`,
          );
        }

        return {
          success: true,
          message: `Better Auth files copied, package installed, and BETTER_AUTH_SECRET added to .env. Configure auth providers in src/server/better-auth/config.ts.`,
        };
      } catch (err) {
        const error = err as Error;
        return {
          success: false,
          message: `Failed to set up Better Auth: ${error.message}`,
        };
      }
    },
  };
};
