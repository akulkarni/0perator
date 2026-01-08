import { stdioServerFactory } from "@tigerdata/mcp-boilerplate";
import { context, serverInfo } from "./serverInfo.js";
import { getApiFactories } from "./tools/index.js";

export interface McpServerOptions {
  disableViewSkill?: boolean;
}

/**
 * Start the MCP server in stdio mode
 */
export async function startMcpServer(
  options: McpServerOptions = {}
): Promise<void> {
  const factoryOptions: import("./tools/index.js").GetApiFactoriesOptions = {};
  if (options.disableViewSkill) {
    factoryOptions.disableViewSkill = true;
  }
  const apiFactories = await getApiFactories(factoryOptions);

  await stdioServerFactory({
    ...serverInfo,
    context,
    apiFactories,
  });
}
