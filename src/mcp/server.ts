import { stdioServerFactory } from '@tigerdata/mcp-boilerplate';
import { apiFactories } from './tools/index.js';
import { context, serverInfo } from './serverInfo.js';

/**
 * Start the MCP server in stdio mode
 */
export async function startMcpServer(): Promise<void> {
  await stdioServerFactory({
    ...serverInfo,
    context,
    apiFactories,
  });
}
