import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { spawn } from "node:child_process";
import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import * as dotenv from "dotenv";
import type { ServerContext } from "../../types.js";

const vercelEnvironments = ["production", "preview", "development"] as const;
type VercelEnvironment = (typeof vercelEnvironments)[number];

const inputSchema = {
  application_directory: z
    .string()
    .describe("Path to the application directory containing .env and Vercel project"),
  env_file: z
    .string()
    .default(".env")
    .describe("Path to .env file (relative to application_directory)"),
  environments: z
    .array(z.enum(vercelEnvironments))
    .default(["production", "preview"])
    .describe(
      "Vercel environments to upload to (default: production and preview)"
    ),
} as const;

const outputSchema = {
  success: z.boolean().describe("Whether all env vars were uploaded"),
  message: z.string().describe("Status message"),
  uploaded: z
    .array(z.string())
    .optional()
    .describe("List of uploaded variable names"),
  failed: z
    .array(z.object({ name: z.string(), error: z.string() }))
    .optional()
    .describe("List of failed variables with error messages"),
  skipped_empty: z
    .array(z.string())
    .optional()
    .describe("List of variable names skipped because they had empty values"),
} as const;

type FailedVar = {
  name: string;
  error: string;
};

type OutputSchema = {
  success: boolean;
  message: string;
  uploaded?: string[];
  failed?: FailedVar[];
  skipped_empty?: string[];
};

interface ParsedEnvResult {
  vars: Record<string, string>;
  skippedEmpty: string[];
}

function readEnvFile(appDir: string, envFilePath: string): ParsedEnvResult {
  const absolutePath = resolve(appDir, envFilePath);

  if (!existsSync(absolutePath)) {
    throw new Error(`.env file not found at: ${absolutePath}`);
  }

  const raw = readFileSync(absolutePath, "utf8");
  const parsed = dotenv.parse(raw);

  const vars: Record<string, string> = {};
  const skippedEmpty: string[] = [];

  for (const [key, value] of Object.entries(parsed)) {
    if (!key.trim()) continue;
    if (value === undefined || value === null || value === "") {
      skippedEmpty.push(key);
      continue;
    }
    vars[key] = value;
  }

  return { vars, skippedEmpty };
}

function runVercelEnvAddSingle(
  appDir: string,
  name: string,
  value: string,
  vercelEnv: VercelEnvironment
): Promise<void> {
  return new Promise((resolve, reject) => {
    const args = ["vercel", "env", "add", name, vercelEnv, "--cwd", appDir, "--sensitive", "--force"];

    const child = spawn("npx", args, {
      cwd: appDir,
      stdio: ["pipe", "pipe", "pipe"],
      env: {
        ...process.env,
        VERCEL_TELEMETRY_DISABLED: "1",
      },
    });

    let stderr = "";
    child.stderr.on("data", (data: Buffer) => {
      stderr += data.toString();
    });

    child.on("error", (err) => {
      reject(err);
    });

    child.on("exit", (code) => {
      if (code === 0) {
        resolve();
      } else {
        const errorMsg = stderr.trim() || `exit code ${code ?? "unknown"}`;
        reject(new Error(errorMsg));
      }
    });

    // Send the secret through stdin so it never appears on the command line
    child.stdin.write(value);
    child.stdin.write("\n");
    child.stdin.end();
  });
}

async function runVercelEnvAdd(
  appDir: string,
  name: string,
  value: string,
  environments: VercelEnvironment[]
): Promise<void> {
  for (const env of environments) {
    await runVercelEnvAddSingle(appDir, name, value, env);
  }
}

export const uploadEnvToVercelFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: "upload_env_to_vercel",
    config: {
      title: "Upload Env to Vercel",
      description:
        "Upload environment variables from a .env file to Vercel project",
      inputSchema,
      outputSchema,
    },
    fn: async ({ application_directory, env_file, environments }): Promise<OutputSchema> => {
      const appDir = resolve(process.cwd(), application_directory);

      let parsed: ParsedEnvResult;
      try {
        parsed = readEnvFile(appDir, env_file);
      } catch (err) {
        const error = err as Error;
        return {
          success: false,
          message: error.message,
        };
      }

      const { vars: envVars, skippedEmpty } = parsed;

      const varNames = Object.keys(envVars);
      if (varNames.length === 0) {
        return {
          success: false,
          message: "No environment variables with values found in .env file",
          skipped_empty: skippedEmpty.length > 0 ? skippedEmpty : undefined,
        };
      }

      const uploaded: string[] = [];
      const failed: FailedVar[] = [];

      for (const [name, value] of Object.entries(envVars)) {
        try {
          await runVercelEnvAdd(appDir, name, value, environments);
          uploaded.push(name);
        } catch (err) {
          const error = err as Error;
          failed.push({ name, error: error.message });
        }
      }

      const skipped_empty = skippedEmpty.length > 0 ? skippedEmpty : undefined;

      if (failed.length === 0) {
        return {
          success: true,
          message: `Uploaded ${uploaded.length} environment variables to Vercel`,
          uploaded,
          skipped_empty,
        };
      } else if (uploaded.length === 0) {
        return {
          success: false,
          message: "Failed to upload any environment variables",
          failed,
          skipped_empty,
        };
      } else {
        return {
          success: false,
          message: `Partially completed: ${uploaded.length} uploaded, ${failed.length} failed`,
          uploaded,
          failed,
          skipped_empty,
        };
      }
    },
  };
};
