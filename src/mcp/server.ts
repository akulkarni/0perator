import { stdioServerFactory } from "@tigerdata/mcp-boilerplate";
import { context, serverInfo } from "./serverInfo.js";
import { apiFactories } from "./tools/index.js";

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
