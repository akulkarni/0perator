import { exec } from 'child_process';
import { promisify } from 'util';
import { join } from 'path';
import { packageRoot } from '../config.js';

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
    if (!error.stderr?.includes('already exists')) {
      throw new Error(`Failed to install Tiger MCP: ${error.message}`);
    }
  }
}

/**
 * Install 0perator MCP for the given IDE client
 * Uses tiger-cli's mcpinstall under the hood
 */
export async function install0peratorMcp(
  clientName: string,
  options: InstallOptions = {},
): Promise<void> {
  let command: string;
  let args: string;

  if (options.devMode) {
    // Dev mode: use npx tsx with source file
    const srcPath = join(packageRoot, 'src', 'index.ts');
    command = 'npx';
    args = `tsx ${srcPath} mcp start`;
  } else {
    // Production: use the installed binary
    command = process.argv[1];
    args = 'mcp start';
  }

  // Use tiger CLI to install MCP config
  await execAsync(
    `tiger mcp install-raw ${clientName} --name 0perator --command "${command}" --args "${args}" --no-backup`,
  );
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
