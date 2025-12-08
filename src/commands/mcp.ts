import { Command } from 'commander';
import { startMcpServer } from '../mcp/server.js';

export function createMcpCommand(): Command {
  const mcp = new Command('mcp').description('MCP server commands');

  mcp
    .command('start')
    .description('Start the MCP server')
    .action(async () => {
      await startMcpServer();
    });

  return mcp;
}
