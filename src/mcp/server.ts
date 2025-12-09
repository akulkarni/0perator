import { stdioServerFactory } from "@tigerdata/mcp-boilerplate";
import { context, serverInfo } from "./serverInfo.js";
import { getApiFactories } from "./tools/index.js";

/**
 * Start the MCP server in stdio mode
 */
export async function startMcpServer(): Promise<void> {
  const apiFactories = await getApiFactories();

  await stdioServerFactory({
    ...serverInfo,
    context,
    apiFactories,
  });
}
