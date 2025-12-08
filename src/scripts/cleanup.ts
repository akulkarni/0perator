#!/usr/bin/env node
/**
 * Cleanup script that runs during `npm uninstall -g 0perator`
 * Removes configuration files and directories.
 */
import { rm } from 'fs/promises';
import { join } from 'path';
import { homedir } from 'os';

async function cleanup(): Promise<void> {
  // Remove config directory
  const configDir = join(homedir(), '.config', '0perator');
  try {
    await rm(configDir, { recursive: true, force: true });
    console.log('0perator: Removed config directory');
  } catch {
    // Directory might not exist, that's fine
  }

  console.log('0perator: Cleanup complete');
  console.log(
    "0perator: Please manually remove '0perator' from your IDE's MCP configuration",
  );
}

cleanup().catch((err) => {
  console.error('0perator: Cleanup failed:', err);
  // Don't exit with error code - let npm uninstall continue
});
