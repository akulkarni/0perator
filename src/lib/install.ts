import { exec } from "node:child_process";
import { join } from "node:path";
import { promisify } from "node:util";
import { packageRoot } from "../config.js";
import { installMCPForClient } from "./mcpInstall.js";
import { getPackageRunner } from "./packageManager.js";

const execAsync = promisify(exec);

export interface InstallOptions {
  devMode?: boolean;
}

/**
 * Install Tiger MCP for the given IDE client
 */
export async function installTigerMcp(clientName: string): Promise<void> {
  try {
    await execAsync(`tiger mcp install ${clientName} --no-backup`);
  } catch (err) {
    const error = err as Error & { stderr?: string };
    // Ignore if already installed
    if (!error.stderr?.includes("already exists")) {
      throw new Error(`Failed to install Tiger MCP: ${error.message}`);
    }
  }
}

/**
 * Install 0perator MCP for the given IDE client
 * Uses native TypeScript implementation
 */
export async function install0peratorMcp(
  clientName: string,
  options: InstallOptions = {},
): Promise<void> {
  let command: string;
  let args: string[];

  // Detect package runner (npx, bunx, pnpm dlx)
  const runner = await getPackageRunner(process.cwd());
  const runnerParts = runner.split(" ");

  if (options.devMode) {
    // Dev mode: use package runner with tsx to run source file
    const srcPath = join(packageRoot, "src", "index.ts");
    command = runnerParts[0];
    args = [...runnerParts.slice(1), "tsx", srcPath, "mcp", "start"];
  } else {
    // Production: use package runner to run the installed package
    command = runnerParts[0];
    args = [...runnerParts.slice(1), "0perator", "mcp", "start"];
  }

  await installMCPForClient({
    clientName,
    serverName: "0perator",
    command,
    args,
    createBackup: false,
  });
}

/**
 * Install both Tiger and 0perator MCP servers
 */
export async function installBoth(
  clientName: string,
  options: InstallOptions = {},
): Promise<void> {
  await installTigerMcp(clientName);
  await install0peratorMcp(clientName, options);
}
